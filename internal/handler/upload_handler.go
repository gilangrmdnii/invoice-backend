package handler

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/gilangrmdnii/invoice-backend/pkg/response"
)

var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".pdf":  true,
}

type UploadHandler struct {
	uploadDir string
}

func NewUploadHandler(uploadDir string) *UploadHandler {
	return &UploadHandler{uploadDir: uploadDir}
}

func (h *UploadHandler) Upload(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "file is required")
	}

	// Max 5MB
	if file.Size > 5*1024*1024 {
		return response.Error(c, fiber.StatusBadRequest, "file size must be less than 5MB")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExtensions[ext] {
		return response.Error(c, fiber.StatusBadRequest, "only JPG, PNG, and PDF files are allowed")
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s-%s%s", time.Now().Format("20060102"), uuid.New().String()[:8], ext)
	savePath := filepath.Join(h.uploadDir, filename)

	if err := c.SaveFile(file, savePath); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to save file")
	}

	fileURL := "/uploads/" + filename

	return response.Success(c, fiber.StatusOK, "file uploaded successfully", fiber.Map{
		"file_url": fileURL,
	})
}
