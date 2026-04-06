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
)

type ProjectWorkerService struct {
	workerRepo  *repository.ProjectWorkerRepository
	projectRepo *repository.ProjectRepository
	memberRepo  *repository.ProjectMemberRepository
	auditRepo   *repository.AuditLogRepository
	userRepo    *repository.UserRepository
}

func NewProjectWorkerService(
	workerRepo *repository.ProjectWorkerRepository,
	projectRepo *repository.ProjectRepository,
	memberRepo *repository.ProjectMemberRepository,
	auditRepo *repository.AuditLogRepository,
	userRepo *repository.UserRepository,
) *ProjectWorkerService {
	return &ProjectWorkerService{
		workerRepo:  workerRepo,
		projectRepo: projectRepo,
		memberRepo:  memberRepo,
		auditRepo:   auditRepo,
		userRepo:    userRepo,
	}
}

func (s *ProjectWorkerService) logAudit(ctx context.Context, userID uint64, action string, entityID uint64, details string) {
	_, err := s.auditRepo.Create(ctx, &model.AuditLog{
		UserID:     userID,
		Action:     action,
		EntityType: "project_worker",
		EntityID:   entityID,
		Details:    details,
	})
	if err != nil {
		log.Printf("audit log error: %v", err)
	}
}

func (s *ProjectWorkerService) Create(ctx context.Context, req *request.CreateProjectWorkerRequest, userID uint64, role string) (*response.ProjectWorkerResponse, error) {
	// Verify project exists
	_, err := s.projectRepo.FindByID(ctx, req.ProjectID)
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

	worker := &model.ProjectWorker{
		ProjectID: req.ProjectID,
		FullName:  req.FullName,
		Role:      req.Role,
		Phone:     req.Phone,
		DailyWage: req.DailyWage,
		IsActive:  true,
		AddedBy:   userID,
	}

	id, err := s.workerRepo.Create(ctx, worker)
	if err != nil {
		return nil, fmt.Errorf("create project worker: %w", err)
	}

	s.logAudit(ctx, userID, "CREATE", id, fmt.Sprintf("name=%s, role=%s", req.FullName, req.Role))

	return s.GetByID(ctx, id)
}

func (s *ProjectWorkerService) ListByProject(ctx context.Context, projectID uint64) ([]response.ProjectWorkerResponse, error) {
	// Verify project exists
	_, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("project not found")
		}
		return nil, err
	}

	workers, err := s.workerRepo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	result := make([]response.ProjectWorkerResponse, 0, len(workers))
	for _, w := range workers {
		resp := toWorkerResponse(&w)
		if user, err := s.userRepo.FindByID(ctx, w.AddedBy); err == nil {
			resp.AdderName = user.FullName
		}
		result = append(result, resp)
	}
	return result, nil
}

func (s *ProjectWorkerService) GetByID(ctx context.Context, id uint64) (*response.ProjectWorkerResponse, error) {
	worker, err := s.workerRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("worker not found")
		}
		return nil, err
	}
	resp := toWorkerResponse(worker)
	if user, err := s.userRepo.FindByID(ctx, worker.AddedBy); err == nil {
		resp.AdderName = user.FullName
	}
	return &resp, nil
}

func (s *ProjectWorkerService) Update(ctx context.Context, id uint64, req *request.UpdateProjectWorkerRequest, userID uint64) (*response.ProjectWorkerResponse, error) {
	worker, err := s.workerRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("worker not found")
		}
		return nil, err
	}

	if req.FullName != "" {
		worker.FullName = req.FullName
	}
	if req.Role != "" {
		worker.Role = req.Role
	}
	if req.Phone != "" {
		worker.Phone = req.Phone
	}
	if req.DailyWage > 0 {
		worker.DailyWage = req.DailyWage
	}
	if req.IsActive != nil {
		worker.IsActive = *req.IsActive
	}

	if err := s.workerRepo.Update(ctx, worker); err != nil {
		return nil, fmt.Errorf("update worker: %w", err)
	}

	s.logAudit(ctx, userID, "UPDATE", id, "")

	return s.GetByID(ctx, id)
}

func (s *ProjectWorkerService) Delete(ctx context.Context, id uint64, userID uint64) error {
	worker, err := s.workerRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("worker not found")
		}
		return err
	}

	if err := s.workerRepo.Delete(ctx, id); err != nil {
		return err
	}

	s.logAudit(ctx, userID, "DELETE", id, fmt.Sprintf("name=%s", worker.FullName))
	return nil
}

func toWorkerResponse(w *model.ProjectWorker) response.ProjectWorkerResponse {
	return response.ProjectWorkerResponse{
		ID:        w.ID,
		ProjectID: w.ProjectID,
		FullName:  w.FullName,
		Role:      w.Role,
		Phone:     w.Phone,
		DailyWage: w.DailyWage,
		IsActive:  w.IsActive,
		AddedBy:   w.AddedBy,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}
