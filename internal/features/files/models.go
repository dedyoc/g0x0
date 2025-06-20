package files

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID           uuid.UUID `json:"id"`
	SHA256       string    `json:"sha256"`
	OriginalName string    `json:"original_name"`
	MimeType     string    `json:"mime_type"`
	FileSize     int64     `json:"file_size"`
	UploadIP     string    `json:"upload_ip"`
	UserAgent    string    `json:"user_agent"`
	Secret       string    `json:"secret,omitempty"`
	MgmtToken    string    `json:"mgmt_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	Removed      bool      `json:"removed"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
