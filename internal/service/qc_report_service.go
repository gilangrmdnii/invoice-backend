package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gilangrmdnii/invoice-backend/internal/dto/request"
	"github.com/gilangrmdnii/invoice-backend/internal/dto/response"
	"github.com/gilangrmdnii/invoice-backend/internal/model"
	"github.com/gilangrmdnii/invoice-backend/internal/repository"
	"github.com/gilangrmdnii/invoice-backend/internal/sse"
)

type QCReportService struct {
	reportRepo  *repository.QCReportRepository
	projectRepo *repository.ProjectRepository
	memberRepo  *repository.ProjectMemberRepository
	auditRepo   *repository.AuditLogRepository
	notifRepo   *repository.NotificationRepository
	userRepo    *repository.UserRepository
	sseHub      *sse.Hub
}

func NewQCReportService(
	reportRepo *repository.QCReportRepository,
	projectRepo *repository.ProjectRepository,
	memberRepo *repository.ProjectMemberRepository,
	auditRepo *repository.AuditLogRepository,
	notifRepo *repository.NotificationRepository,
	userRepo *repository.UserRepository,
	sseHub *sse.Hub,
) *QCReportService {
	return &QCReportService{
		reportRepo:  reportRepo,
		projectRepo: projectRepo,
		memberRepo:  memberRepo,
		auditRepo:   auditRepo,
		notifRepo:   notifRepo,
		userRepo:    userRepo,
		sseHub:      sseHub,
	}
}

func (s *QCReportService) logAudit(ctx context.Context, userID uint64, action string, entityID uint64, details string) {
	_, err := s.auditRepo.Create(ctx, &model.AuditLog{
		UserID:     userID,
		Action:     action,
		EntityType: "qc_report",
		EntityID:   entityID,
		Details:    details,
	})
	if err != nil {
		log.Printf("audit log error: %v", err)
	}
}

func (s *QCReportService) notifyRoles(ctx context.Context, roles []string, title, message string, notifType model.NotificationType, refID uint64) {
	users, err := s.userRepo.FindByRoles(ctx, roles)
	if err != nil {
		log.Printf("find users by roles error: %v", err)
		return
	}
	for _, u := range users {
		id, err := s.notifRepo.Create(ctx, &model.Notification{
			UserID:      u.ID,
			Title:       title,
			Message:     message,
			Type:        notifType,
			ReferenceID: &refID,
		})
		if err != nil {
			log.Printf("notification error: %v", err)
			continue
		}
		s.sseHub.Publish(u.ID, sse.Event{
			Type: string(notifType),
			Data: map[string]interface{}{"id": id, "title": title, "message": message, "reference_id": refID},
		})
	}
}

func parseDateNullable(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		// try RFC3339
		if t2, err2 := time.Parse(time.RFC3339, *s); err2 == nil {
			return &t2
		}
		return nil
	}
	return &t
}

func buildItemsModel(items []request.QCReportItemRequest) ([]model.QCReportItem, float64) {
	var total float64
	out := make([]model.QCReportItem, 0, len(items))
	for i, it := range items {
		status := model.QCStatusNone
		if it.Status != "" {
			status = model.QCItemStatus(it.Status)
		}
		subtotal := float64(it.Quantity) * it.UnitPrice
		total += subtotal
		sortOrder := it.SortOrder
		if sortOrder == 0 {
			sortOrder = i
		}
		out = append(out, model.QCReportItem{
			Category:  model.QCItemCategory(it.Category),
			Status:    status,
			Label:     it.Label,
			Quantity:  it.Quantity,
			UnitPrice: it.UnitPrice,
			Subtotal:  subtotal,
			SortOrder: sortOrder,
		})
	}
	return out, total
}

func buildRecruitersModel(recruiters []request.QCRecruiterPerformanceRequest) []model.QCRecruiterPerformance {
	out := make([]model.QCRecruiterPerformance, 0, len(recruiters))
	for i, r := range recruiters {
		sortOrder := r.SortOrder
		if sortOrder == 0 {
			sortOrder = i
		}
		out = append(out, model.QCRecruiterPerformance{
			RecruiterName: r.RecruiterName,
			Total:         r.Total,
			OKPerpi:       r.OKPerpi,
			DOPerpi:       r.DOPerpi,
			OKQC:          r.OKQC,
			DOQC:          r.DOQC,
			Notes:         r.Notes,
			SortOrder:     sortOrder,
		})
	}
	return out
}

func (s *QCReportService) Create(ctx context.Context, req *request.CreateQCReportRequest, userID uint64, role string) (*response.QCReportResponse, error) {
	project, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("project not found")
		}
		return nil, err
	}

	// Field roles must be member of the project
	if model.IsFieldRole(role) {
		isMember, err := s.memberRepo.Exists(ctx, req.ProjectID, userID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, fmt.Errorf("not a member of this project")
		}
	}

	items, total := buildItemsModel(req.Items)
	recruiters := buildRecruitersModel(req.Recruiters)

	rep := &model.QCReport{
		ProjectID:                 req.ProjectID,
		QCUserID:                  req.QCUserID,
		SPVNames:                  req.SPVNames,
		ProjectType:               model.QCProjectType(req.ProjectType),
		Methodology:               model.QCMethodology(req.Methodology),
		City:                      req.City,
		Area:                      model.QCArea(req.Area),
		ExecutionStartDate:        parseDateNullable(req.ExecutionStartDate),
		ExecutionEndDate:          parseDateNullable(req.ExecutionEndDate),
		BriefingDate:              parseDateNullable(req.BriefingDate),
		WorkStartDate:             parseDateNullable(req.WorkStartDate),
		WorkEndDate:               parseDateNullable(req.WorkEndDate),
		VisitTarget:               req.VisitTarget,
		VisitOK:                   req.VisitOK,
		TelpTarget:                req.TelpTarget,
		TelpOK:                    req.TelpOK,
		TotalAmount:               total,
		Location:                  req.Location,
		ReportDate:                parseDateNullable(req.ReportDate),
		QCSignatoryName:           req.QCSignatoryName,
		QCSignatoryTitle:          defaultString(req.QCSignatoryTitle, "Quality Control"),
		CoordinatorSignatoryName:  req.CoordinatorSignatoryName,
		CoordinatorSignatoryTitle: defaultString(req.CoordinatorSignatoryTitle, "Koordinator QC"),
		Note:                      req.Note,
		CreatedBy:                 userID,
	}

	id, err := s.reportRepo.Create(ctx, rep, items, recruiters)
	if err != nil {
		return nil, fmt.Errorf("create qc report: %w", err)
	}

	s.logAudit(ctx, userID, "CREATE", id, fmt.Sprintf("project=%s, total=%.0f", project.Name, total))
	s.notifyRoles(ctx, []string{"FINANCE", "OWNER"}, "QC Report Created",
		fmt.Sprintf("A new QC report has been created for project %s", project.Name),
		model.NotifQCReportCreated, id)

	return s.GetByID(ctx, id)
}

func (s *QCReportService) List(ctx context.Context, userID uint64, role string, projectID *uint64) ([]response.QCReportResponse, error) {
	var reports []model.QCReport
	var err error

	if projectID != nil {
		reports, err = s.reportRepo.FindByProjectID(ctx, *projectID)
	} else if model.IsFieldRole(role) {
		projects, pErr := s.projectRepo.FindByMemberUserID(ctx, userID)
		if pErr != nil {
			return nil, pErr
		}
		projectIDs := make([]uint64, len(projects))
		for i, p := range projects {
			projectIDs[i] = p.ID
		}
		reports, err = s.reportRepo.FindByProjectIDs(ctx, projectIDs)
	} else {
		reports, err = s.reportRepo.FindAll(ctx)
	}
	if err != nil {
		return nil, err
	}

	result := make([]response.QCReportResponse, 0, len(reports))
	for _, rep := range reports {
		resp, err := s.enrichReport(ctx, &rep, false)
		if err != nil {
			return nil, err
		}
		result = append(result, *resp)
	}
	return result, nil
}

func (s *QCReportService) GetByID(ctx context.Context, id uint64) (*response.QCReportResponse, error) {
	rep, err := s.reportRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("qc report not found")
		}
		return nil, err
	}
	return s.enrichReport(ctx, rep, true)
}

func (s *QCReportService) Update(ctx context.Context, id uint64, req *request.UpdateQCReportRequest, userID uint64, role string) (*response.QCReportResponse, error) {
	rep, err := s.reportRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("qc report not found")
		}
		return nil, err
	}

	// Field roles can only update own reports
	if model.IsFieldRole(role) && rep.CreatedBy != userID {
		return nil, fmt.Errorf("not authorized to update this report")
	}

	if req.QCUserID > 0 {
		rep.QCUserID = req.QCUserID
	}
	if req.SPVNames != "" {
		rep.SPVNames = req.SPVNames
	}
	if req.ProjectType != "" {
		rep.ProjectType = model.QCProjectType(req.ProjectType)
	}
	if req.Methodology != "" {
		rep.Methodology = model.QCMethodology(req.Methodology)
	}
	if req.City != "" {
		rep.City = req.City
	}
	if req.Area != "" {
		rep.Area = model.QCArea(req.Area)
	}
	if req.ExecutionStartDate != nil {
		rep.ExecutionStartDate = parseDateNullable(req.ExecutionStartDate)
	}
	if req.ExecutionEndDate != nil {
		rep.ExecutionEndDate = parseDateNullable(req.ExecutionEndDate)
	}
	if req.BriefingDate != nil {
		rep.BriefingDate = parseDateNullable(req.BriefingDate)
	}
	if req.WorkStartDate != nil {
		rep.WorkStartDate = parseDateNullable(req.WorkStartDate)
	}
	if req.WorkEndDate != nil {
		rep.WorkEndDate = parseDateNullable(req.WorkEndDate)
	}
	if req.VisitTarget != nil {
		rep.VisitTarget = *req.VisitTarget
	}
	if req.VisitOK != nil {
		rep.VisitOK = *req.VisitOK
	}
	if req.TelpTarget != nil {
		rep.TelpTarget = *req.TelpTarget
	}
	if req.TelpOK != nil {
		rep.TelpOK = *req.TelpOK
	}
	if req.Location != "" {
		rep.Location = req.Location
	}
	if req.ReportDate != nil {
		rep.ReportDate = parseDateNullable(req.ReportDate)
	}
	if req.QCSignatoryName != "" {
		rep.QCSignatoryName = req.QCSignatoryName
	}
	if req.QCSignatoryTitle != "" {
		rep.QCSignatoryTitle = req.QCSignatoryTitle
	}
	if req.CoordinatorSignatoryName != "" {
		rep.CoordinatorSignatoryName = req.CoordinatorSignatoryName
	}
	if req.CoordinatorSignatoryTitle != "" {
		rep.CoordinatorSignatoryTitle = req.CoordinatorSignatoryTitle
	}
	if req.Note != "" {
		rep.Note = req.Note
	}

	items, total := buildItemsModel(req.Items)
	recruiters := buildRecruitersModel(req.Recruiters)
	rep.TotalAmount = total

	if err := s.reportRepo.Update(ctx, rep, items, recruiters); err != nil {
		return nil, fmt.Errorf("update qc report: %w", err)
	}

	s.logAudit(ctx, userID, "UPDATE", id, "")
	return s.GetByID(ctx, id)
}

func (s *QCReportService) Delete(ctx context.Context, id uint64, userID uint64, role string) error {
	rep, err := s.reportRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("qc report not found")
		}
		return err
	}

	if model.IsFieldRole(role) && rep.CreatedBy != userID {
		return fmt.Errorf("not authorized to delete this report")
	}

	if err := s.reportRepo.Delete(ctx, id); err != nil {
		return err
	}

	s.logAudit(ctx, userID, "DELETE", id, "")
	return nil
}

func (s *QCReportService) enrichReport(ctx context.Context, rep *model.QCReport, includeChildren bool) (*response.QCReportResponse, error) {
	resp := response.QCReportResponse{
		ID:                        rep.ID,
		ProjectID:                 rep.ProjectID,
		QCUserID:                  rep.QCUserID,
		SPVNames:                  rep.SPVNames,
		ProjectType:               string(rep.ProjectType),
		Methodology:               string(rep.Methodology),
		City:                      rep.City,
		Area:                      string(rep.Area),
		ExecutionStartDate:        rep.ExecutionStartDate,
		ExecutionEndDate:          rep.ExecutionEndDate,
		BriefingDate:              rep.BriefingDate,
		WorkStartDate:             rep.WorkStartDate,
		WorkEndDate:               rep.WorkEndDate,
		VisitTarget:               rep.VisitTarget,
		VisitOK:                   rep.VisitOK,
		TelpTarget:                rep.TelpTarget,
		TelpOK:                    rep.TelpOK,
		TotalAmount:               rep.TotalAmount,
		Location:                  rep.Location,
		ReportDate:                rep.ReportDate,
		QCSignatoryName:           rep.QCSignatoryName,
		QCSignatoryTitle:          rep.QCSignatoryTitle,
		CoordinatorSignatoryName:  rep.CoordinatorSignatoryName,
		CoordinatorSignatoryTitle: rep.CoordinatorSignatoryTitle,
		Note:                      rep.Note,
		CreatedBy:                 rep.CreatedBy,
		CreatedAt:                 rep.CreatedAt,
		UpdatedAt:                 rep.UpdatedAt,
		Items:                     []response.QCReportItemResponse{},
		Recruiters:                []response.QCRecruiterPerformanceResponse{},
	}

	if project, err := s.projectRepo.FindByID(ctx, rep.ProjectID); err == nil {
		resp.ProjectName = project.Name
	}
	if qcUser, err := s.userRepo.FindByID(ctx, rep.QCUserID); err == nil {
		resp.QCUserName = qcUser.FullName
	}
	if creator, err := s.userRepo.FindByID(ctx, rep.CreatedBy); err == nil {
		resp.CreatorName = creator.FullName
	}

	if includeChildren {
		items, err := s.reportRepo.FindItems(ctx, rep.ID)
		if err != nil {
			return nil, err
		}
		for _, it := range items {
			resp.Items = append(resp.Items, response.QCReportItemResponse{
				ID:        it.ID,
				Category:  string(it.Category),
				Status:    string(it.Status),
				Label:     it.Label,
				Quantity:  it.Quantity,
				UnitPrice: it.UnitPrice,
				Subtotal:  it.Subtotal,
				SortOrder: it.SortOrder,
				CreatedAt: it.CreatedAt,
			})
		}

		recruiters, err := s.reportRepo.FindRecruiters(ctx, rep.ID)
		if err != nil {
			return nil, err
		}
		for _, rc := range recruiters {
			resp.Recruiters = append(resp.Recruiters, response.QCRecruiterPerformanceResponse{
				ID:            rc.ID,
				RecruiterName: rc.RecruiterName,
				Total:         rc.Total,
				OKPerpi:       rc.OKPerpi,
				DOPerpi:       rc.DOPerpi,
				OKQC:          rc.OKQC,
				DOQC:          rc.DOQC,
				Notes:         rc.Notes,
				SortOrder:     rc.SortOrder,
				CreatedAt:     rc.CreatedAt,
			})
		}
	}

	return &resp, nil
}

func defaultString(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
