package upload

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/RESERPIX/hubigr/internal/security"
)

const (
	MaxAvatarSize = 2 << 20 // 2 МБ согласно ТЗ
	UploadDir     = "./uploads/avatars"
)

type AvatarUploader struct {
	baseURL string
}

func NewAvatarUploader(baseURL string) *AvatarUploader {
	// Создаем директорию если не существует
	os.MkdirAll(UploadDir, 0755)
	return &AvatarUploader{baseURL: baseURL}
}

// UploadAvatar загружает аватар согласно UC-1.2.1 (jpeg/png, до 2 МБ)
func (u *AvatarUploader) UploadAvatar(userID int64, file *multipart.FileHeader) (string, error) {
	// Проверка размера
	if file.Size > MaxAvatarSize {
		return "", fmt.Errorf("файл слишком большой, максимум 2 МБ")
	}

	// Проверка типа файла
	contentType := file.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		return "", fmt.Errorf("неподдерживаемый формат файла, только JPEG/PNG")
	}

	// Открытие файла
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Генерация уникального имени файла
	ext := getFileExtension(file.Filename)
	filename := fmt.Sprintf("%d_%s%s", userID, security.GenerateToken()[:16], ext)
	filepath := filepath.Join(UploadDir, filename)

	// Создание файла на диске
	dst, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Копирование данных
	_, err = io.Copy(dst, src)
	if err != nil {
		os.Remove(filepath) // Удаляем файл при ошибке
		return "", err
	}

	// Возвращаем URL для доступа к файлу
	avatarURL := fmt.Sprintf("%s/uploads/avatars/%s", u.baseURL, filename)
	return avatarURL, nil
}

// DeleteAvatar удаляет старый аватар
func (u *AvatarUploader) DeleteAvatar(avatarURL string) error {
	if avatarURL == "" {
		return nil
	}

	// Извлекаем имя файла из URL
	parts := strings.Split(avatarURL, "/")
	if len(parts) == 0 {
		return nil
	}
	filename := parts[len(parts)-1]
	filepath := filepath.Join(UploadDir, filename)

	// Удаляем файл если существует
	if _, err := os.Stat(filepath); err == nil {
		return os.Remove(filepath)
	}
	return nil
}

func isValidImageType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
	}

	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}
	return false
}

func getFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return ".jpg" // дефолтное расширение
	}
	return strings.ToLower(ext)
}
