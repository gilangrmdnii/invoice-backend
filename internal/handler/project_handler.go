package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/gilangrmdnii/invoice-backend/internal/dto/request"
	"github.com/gilangrmdnii/invoice-backend/internal/middleware"
	"github.com/gilangrmdnii/invoice-backend/internal/service"
	"github.com/gilangrmdnii/invoice-backend/pkg/response"
	"github.com/gilangrmdnii/invoice-backend/pkg/validator"
)

type ProjectHandler struct {
	projectService *service.ProjectService
}

func NewProjectHandler(projectService *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

func (h *ProjectHandler) Create(c *fiber.Ctx) error {
	var req request.CreateProjectRequest
	if err := validator.ParseAndValidate(c, &req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	userID := middleware.GetUserID(c)
	result, err := h.projectService.Create(c.Context(), &req, userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to create project")
	}

	return response.Success(c, fiber.StatusCreated, "project created successfully", result)
}

func (h *ProjectHandler) List(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	projects, err := h.projectService.List(c.Context(), userID, role)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to list projects")
	}

	return response.Success(c, fiber.StatusOK, "projects retrieved successfully", projects)
}

func (h *ProjectHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid project id")
	}

	project, err := h.projectService.GetByID(c.Context(), id)
	if err != nil {
		if err.Error() == "project not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to get project")
	}

	return response.Success(c, fiber.StatusOK, "project retrieved successfully", project)
}

func (h *ProjectHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid project id")
	}

	var req request.UpdateProjectRequest
	if err := validator.ParseAndValidate(c, &req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	project, err := h.projectService.Update(c.Context(), id, &req)
	if err != nil {
		if err.Error() == "project not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to update project")
	}

	return response.Success(c, fiber.StatusOK, "project updated successfully", project)
}

func (h *ProjectHandler) GetPlan(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid project id")
	}

	items, err := h.projectService.GetPlan(c.Context(), id)
	if err != nil {
		if err.Error() == "project not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to get plan")
	}

	return response.Success(c, fiber.StatusOK, "plan retrieved successfully", items)
}

func (h *ProjectHandler) UpdatePlan(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid project id")
	}

	var req request.UpdateProjectPlanRequest
	if err := validator.ParseAndValidate(c, &req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	items, err := h.projectService.UpdatePlan(c.Context(), id, &req)
	if err != nil {
		if err.Error() == "project not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to update plan")
	}

	return response.Success(c, fiber.StatusOK, "plan updated successfully", items)
}

func (h *ProjectHandler) AddMember(c *fiber.Ctx) error {
	projectID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid project id")
	}

	var req request.AddMemberRequest
	if err := validator.ParseAndValidate(c, &req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	member, err := h.projectService.AddMember(c.Context(), projectID, req.UserID)
	if err != nil {
		switch err.Error() {
		case "project not found", "user not found", "member not found":
			return response.Error(c, fiber.StatusNotFound, err.Error())
		case "only SPV users can be added as project members":
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		case "user is already a member of this project":
			return response.Error(c, fiber.StatusConflict, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to add member")
	}

	return response.Success(c, fiber.StatusCreated, "member added successfully", member)
}

func (h *ProjectHandler) RemoveMember(c *fiber.Ctx) error {
	projectID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid project id")
	}

	userID, err := strconv.ParseUint(c.Params("userId"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid user id")
	}

	if err := h.projectService.RemoveMember(c.Context(), projectID, userID); err != nil {
		if err.Error() == "member not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to remove member")
	}

	return response.Success(c, fiber.StatusOK, "member removed successfully", nil)
}

func (h *ProjectHandler) ListMembers(c *fiber.Ctx) error {
	projectID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid project id")
	}

	members, err := h.projectService.ListMembers(c.Context(), projectID)
	if err != nil {
		if err.Error() == "project not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to list members")
	}

	return response.Success(c, fiber.StatusOK, "members retrieved successfully", members)
}
