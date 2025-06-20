package files

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/dedyoc/g0x0/internal/config"
	"github.com/dedyoc/g0x0/internal/utils"
)

type Handler struct {
	db     *sql.DB
	config *config.Config
	repo   *Repository
}

func NewHandler(db *sql.DB, cfg *config.Config) *Handler {
	return &Handler{
		db:     db,
		config: cfg,
		repo:   NewRepository(db),
	}
}

func (h *Handler) Index(c echo.Context) error {
	return c.File("web/templates/index.html")
}

func (h *Handler) Upload(c echo.Context) error {
	// Handle file upload
	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No file provided")
	}

	if file.Size > h.config.MaxFileSize {
		return echo.NewHTTPError(http.StatusRequestEntityTooLarge, "File too large")
	}

	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Cannot open file")
	}
	defer src.Close()

	// Calculate SHA256
	hash := sha256.New()
	if _, err := io.Copy(hash, src); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Cannot hash file")
	}
	sha256sum := fmt.Sprintf("%x", hash.Sum(nil))

	// Reset file pointer
	src.Seek(0, 0)

	// Check expiration
	expiration := h.calculateExpiration(c.FormValue("expires"), file.Size)

	// Check if file exists
	existing, err := h.repo.GetBySHA256(sha256sum)
	if err != nil && err != sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	var fileRecord *File
	var isNew bool

	if existing != nil && !existing.Removed {
		// File exists, update expiration
		existing.ExpiresAt = expiration
		if err := h.repo.Update(existing); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Cannot update file")
		}
		fileRecord = existing
		isNew = false
	} else {
		// New file
		fileRecord = &File{
			ID:           uuid.New(),
			SHA256:       sha256sum,
			OriginalName: file.Filename,
			MimeType:     file.Header.Get("Content-Type"),
			FileSize:     file.Size,
			UploadIP:     c.RealIP(),
			UserAgent:    c.Request().UserAgent(),
			MgmtToken:    utils.GenerateToken(32),
			ExpiresAt:    expiration,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if c.FormValue("secret") == "true" {
			fileRecord.Secret = utils.GenerateToken(h.config.SecretBytes)
		}

		// Save to database
		if err := h.repo.Create(fileRecord); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Cannot save file record")
		}
		// Save file to disk
		if err := h.saveFile(src, sha256sum); err != nil {
			log.Printf("Error saving file %s: %v", sha256sum, err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Cannot save file")
		}
		log.Printf("New file uploaded: %s (%s)", fileRecord.OriginalName, sha256sum)

		isNew = true
	}

	response := map[string]interface{}{
		"url":     h.buildFileURL(fileRecord),
		"expires": fileRecord.ExpiresAt.Unix(),
	}

	if isNew {
		response["token"] = fileRecord.MgmtToken
	}

	return c.JSON(http.StatusOK, response)
}

func (h *Handler) Get(c echo.Context) error {
	id := c.Param("id")

	// Try to parse as file ID
	if fileID, err := uuid.Parse(id); err == nil {
		file, err := h.repo.GetByID(fileID)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "File not found")
		}

		return h.serveFile(c, file, "")
	}

	// Try URL shortening
	return echo.NewHTTPError(http.StatusNotFound, "Not found")
}

func (h *Handler) GetSecret(c echo.Context) error {
	secret := c.Param("secret")
	id := c.Param("id")

	fileID, err := uuid.Parse(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Invalid file ID")
	}

	file, err := h.repo.GetByID(fileID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "File not found")
	}

	return h.serveFile(c, file, secret)
}

func (h *Handler) Manage(c echo.Context) error {
	id := c.Param("id")
	token := c.FormValue("token")

	fileID, err := uuid.Parse(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid file ID")
	}

	file, err := h.repo.GetByID(fileID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "File not found")
	}

	if file.MgmtToken != token {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
	}

	if c.FormValue("delete") == "true" {
		file.Removed = true
		if err := h.repo.Update(file); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Cannot delete file")
		}

		// Remove from filesystem
		os.Remove(filepath.Join(h.config.StoragePath, file.SHA256))

		return c.NoContent(http.StatusOK)
	}

	return echo.NewHTTPError(http.StatusBadRequest, "Invalid action")
}

// ...existing helper methods...
func (h *Handler) calculateExpiration(expiresParam string, fileSize int64) time.Time {
	maxLifespan := h.getMaxLifespan(fileSize)
	maxExpiration := time.Now().Add(maxLifespan)

	if expiresParam == "" {
		return maxExpiration
	}

	if hours, err := strconv.Atoi(expiresParam); err == nil {
		requested := time.Now().Add(time.Duration(hours) * time.Hour)
		if requested.Before(maxExpiration) {
			return requested
		}
	}

	return maxExpiration
}

func (h *Handler) getMaxLifespan(fileSize int64) time.Duration {
	ratio := float64(fileSize) / float64(h.config.MaxFileSize)
	lifespanRange := h.config.MaxExpiration - h.config.MinExpiration
	return h.config.MinExpiration + time.Duration(float64(lifespanRange)*(1-ratio))
}

func (h *Handler) saveFile(src io.Reader, sha256sum string) error {
	if err := os.MkdirAll(h.config.StoragePath, 0755); err != nil {
		return err
	}

	dst, err := os.Create(filepath.Join(h.config.StoragePath, sha256sum))
	log.Printf("Saving file to %s", filepath.Join(h.config.StoragePath, sha256sum))
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func (h *Handler) serveFile(c echo.Context, file *File, secret string) error {
	if file.Removed {
		return echo.NewHTTPError(http.StatusGone, "File removed")
	}

	if file.Secret != "" && file.Secret != secret {
		return echo.NewHTTPError(http.StatusNotFound, "File not found")
	}

	if time.Now().After(file.ExpiresAt) {
		return echo.NewHTTPError(http.StatusGone, "File expired")
	}

	filePath := filepath.Join(h.config.StoragePath, file.SHA256)
	return c.File(filePath)
}

func (h *Handler) buildFileURL(file *File) string {
	if file.Secret != "" {
		return fmt.Sprintf("/s/%s/%s", file.Secret, file.ID.String())
	}
	return fmt.Sprintf("/%s", file.ID.String())
}
