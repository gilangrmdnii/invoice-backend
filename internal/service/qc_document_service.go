package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/gilangrmdnii/invoice-backend/internal/dto/request"
	"github.com/gilangrmdnii/invoice-backend/internal/dto/response"
	"github.com/gilangrmdnii/invoice-backend/internal/model"
	"github.com/gilangrmdnii/invoice-backend/internal/repository"
	"github.com/gilangrmdnii/invoice-backend/internal/sse"
)

type QCDocumentService struct {
	docRepo     *repository.QCDocumentRepository
	projectRepo *repository.ProjectRepository
	memberRepo  *repository.ProjectMemberRepository
	auditRepo   *repository.AuditLogRepository
	notifRepo   *repository.NotificationRepository
	userRepo    *repository.UserRepository
	sseHub      *sse.Hub
}

func NewQCDocumentService(
	docRepo *repository.QCDocumentRepository,
	projectRepo *repository.ProjectRepository,
	memberRepo *repository.ProjectMemberRepository,
	auditRepo *repository.AuditLogRepository,
	notifRepo *repository.NotificationRepository,
	userRepo *repository.UserRepository,
	sseHub *sse.Hub,
) *QCDocumentService {
	return &QCDocumentService{
		docRepo:     docRepo,
		projectRepo: projectRepo,
		memberRepo:  memberRepo,
		auditRepo:   auditRepo,
		notifRepo:   notifRepo,
		userRepo:    userRepo,
		sseHub:      sseHub,
	}
}

func (s *QCDocumentService) logAudit(ctx context.Context, userID uint64, action string, entityID uint64, details string) {
	_, err := s.auditRepo.Create(ctx, &model.AuditLog{
		UserID:     userID,
		Action:     action,
		EntityType: "qc_document",
		EntityID:   entityID,
		Details:    details,
	})
	if err != nil {
		log.Printf("audit log error: %v", err)
	}
}

func (s *QCDocumentService) notifyRoles(ctx context.Context, roles []string, title, message string, notifType model.NotificationType, refID uint64) {
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

func (s *QCDocumentService) Create(ctx context.Context, req *request.CreateQCDocumentRequest, userID uint64, role string) (*response.QCDocumentResponse, error) {
	// Verify project exists
	project, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("project not found")
		}
		return nil, err
	}

	// Field roles must be a member of the project
	if model.IsFieldRole(role) {
		isMember, err := s.memberRepo.Exists(ctx, req.ProjectID, userID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, fmt.Errorf("not a member of this project")
		}
	}

	doc := &model.QCDocument{
		ProjectID:    req.ProjectID,
		Title:        req.Title,
		Description:  req.Description,
		DocumentType: model.DocumentType(req.DocumentType),
		FileURL:      req.FileURL,
		UploadedBy:   userID,
	}

	id, err := s.docRepo.Create(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("create qc document: %w", err)
	}

	s.logAudit(ctx, userID, "CREATE", id, fmt.Sprintf("type=%s, title=%s, project=%s", req.DocumentType, req.Title, project.Name))
	s.notifyRoles(ctx, []string{"FINANCE", "OWNER"}, "New QC Document Uploaded",
		fmt.Sprintf("A new %s document '%s' has been uploaded to project %s", req.DocumentType, req.Title, project.Name),
		model.NotifQCDocumentCreated, id)

	return s.GetByID(ctx, id)
}

func (s *QCDocumentService) List(ctx context.Context, userID uint64, role string, projectID *uint64) ([]response.QCDocumentResponse, error) {
	var docs []model.QCDocument
	var err error

	if projectID != nil {
		docs, err = s.docRepo.FindByProjectID(ctx, *projectID)
	} else if model.IsFieldRole(role) {
		projects, err := s.projectRepo.FindByMemberUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		projectIDs := make([]uint64, len(projects))
		for i, p := range projects {
			projectIDs[i] = p.ID
		}
		docs, err = s.docRepo.FindByProjectIDs(ctx, projectIDs)
		if err != nil {
			return nil, err
		}
	} else {
		docs, err = s.docRepo.FindAll(ctx)
	}
	if err != nil {
		return nil, err
	}

	result := make([]response.QCDocumentResponse, 0, len(docs))
	for _, d := range docs {
		resp := toQCDocumentResponse(&d)
		if user, err := s.userRepo.FindByID(ctx, d.UploadedBy); err == nil {
			resp.UploaderName = user.FullName
		}
		result = append(result, resp)
	}
	return result, nil
}

func (s *QCDocumentService) GetByID(ctx context.Context, id uint64) (*response.QCDocumentResponse, error) {
	doc, err := s.docRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("qc document not found")
		}
		return nil, err
	}
	resp := toQCDocumentResponse(doc)
	if user, err := s.userRepo.FindByID(ctx, doc.UploadedBy); err == nil {
		resp.UploaderName = user.FullName
	}
	return &resp, nil
}

func (s *QCDocumentService) Update(ctx context.Context, id uint64, req *request.UpdateQCDocumentRequest, userID uint64, role string) (*response.QCDocumentResponse, error) {
	doc, err := s.docRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("qc document not found")
		}
		return nil, err
	}

	// Field roles can only update their own documents
	if model.IsFieldRole(role) && doc.UploadedBy != userID {
		return nil, fmt.Errorf("not authorized to update this document")
	}

	if req.Title != "" {
		doc.Title = req.Title
	}
	if req.Description != "" {
		doc.Description = req.Description
	}
	if req.DocumentType != "" {
		doc.DocumentType = model.DocumentType(req.DocumentType)
	}
	if req.FileURL != "" {
		doc.FileURL = req.FileURL
	}

	if err := s.docRepo.Update(ctx, doc); err != nil {
		return nil, fmt.Errorf("update qc document: %w", err)
	}

	s.logAudit(ctx, userID, "UPDATE", id, "")
	return s.GetByID(ctx, id)
}

func (s *QCDocumentService) Delete(ctx context.Context, id uint64, userID uint64, role string) error {
	doc, err := s.docRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("qc document not found")
		}
		return err
	}

	// Field roles can only delete their own documents
	if model.IsFieldRole(role) && doc.UploadedBy != userID {
		return fmt.Errorf("not authorized to delete this document")
	}

	if err := s.docRepo.Delete(ctx, id); err != nil {
		return err
	}

	s.logAudit(ctx, userID, "DELETE", id, "")
	return nil
}

func toQCDocumentResponse(d *model.QCDocument) response.QCDocumentResponse {
	return response.QCDocumentResponse{
		ID:           d.ID,
		ProjectID:    d.ProjectID,
		Title:        d.Title,
		Description:  d.Description,
		DocumentType: string(d.DocumentType),
		FileURL:      d.FileURL,
		UploadedBy:   d.UploadedBy,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}
