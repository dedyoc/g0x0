package files

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(file *File) error {
	query := `
        INSERT INTO files (id, sha256, original_name, mime_type, file_size, 
                          upload_ip, user_agent, secret, mgmt_token, expires_at, 
                          removed, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.db.Exec(query,
		file.ID, file.SHA256, file.OriginalName, file.MimeType, file.FileSize,
		file.UploadIP, file.UserAgent, file.Secret, file.MgmtToken, file.ExpiresAt,
		file.Removed, file.CreatedAt, file.UpdatedAt)

	return err
}

func (r *Repository) GetByID(id uuid.UUID) (*File, error) {
	query := `
        SELECT id, sha256, original_name, mime_type, file_size, upload_ip, 
               user_agent, secret, mgmt_token, expires_at, removed, created_at, updated_at
        FROM files WHERE id = $1`

	file := &File{}
	err := r.db.QueryRow(query, id).Scan(
		&file.ID, &file.SHA256, &file.OriginalName, &file.MimeType, &file.FileSize,
		&file.UploadIP, &file.UserAgent, &file.Secret, &file.MgmtToken, &file.ExpiresAt,
		&file.Removed, &file.CreatedAt, &file.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return file, nil
}

func (r *Repository) GetBySHA256(sha256 string) (*File, error) {
	query := `
        SELECT id, sha256, original_name, mime_type, file_size, upload_ip, 
               user_agent, secret, mgmt_token, expires_at, removed, created_at, updated_at
        FROM files WHERE sha256 = $1`

	file := &File{}
	err := r.db.QueryRow(query, sha256).Scan(
		&file.ID, &file.SHA256, &file.OriginalName, &file.MimeType, &file.FileSize,
		&file.UploadIP, &file.UserAgent, &file.Secret, &file.MgmtToken, &file.ExpiresAt,
		&file.Removed, &file.CreatedAt, &file.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return file, nil
}

func (r *Repository) Update(file *File) error {
	file.UpdatedAt = time.Now()

	query := `
        UPDATE files 
        SET original_name = $2, mime_type = $3, file_size = $4, upload_ip = $5,
            user_agent = $6, secret = $7, mgmt_token = $8, expires_at = $9,
            removed = $10, updated_at = $11
        WHERE id = $1`

	_, err := r.db.Exec(query,
		file.ID, file.OriginalName, file.MimeType, file.FileSize, file.UploadIP,
		file.UserAgent, file.Secret, file.MgmtToken, file.ExpiresAt,
		file.Removed, file.UpdatedAt)

	return err
}

func (r *Repository) GetExpired() ([]*File, error) {
	query := `
        SELECT id, sha256, original_name, mime_type, file_size, upload_ip, 
               user_agent, secret, mgmt_token, expires_at, removed, created_at, updated_at
        FROM files 
        WHERE expires_at < NOW() AND removed = false`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*File
	for rows.Next() {
		file := &File{}
		err := rows.Scan(
			&file.ID, &file.SHA256, &file.OriginalName, &file.MimeType, &file.FileSize,
			&file.UploadIP, &file.UserAgent, &file.Secret, &file.MgmtToken, &file.ExpiresAt,
			&file.Removed, &file.CreatedAt, &file.UpdatedAt)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}
