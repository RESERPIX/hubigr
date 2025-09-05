package upload

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

const (
	MaxAvatarSize = 2 << 20 // 2 МБ согласно ТЗ
	UploadDir     = "./uploads/avatars"
)

type AvatarUploader struct {
	baseURL string
}

func NewAvatarUploader(baseURL string) *AvatarUploader {
	// Создаем директорию с безопасными правами
	os.MkdirAll(UploadDir, 0700)
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

	// Генерация безопасного имени файла
	randomID, err := generateSecureFilename()
	if err != nil {
		return "", fmt.Errorf("ошибка генерации имени: %v", err)
	}
	
	ext := getFileExtension(file.Filename)
	filename := fmt.Sprintf("%d_%s%s", userID, randomID, ext)
	
	// Проверка безопасности пути
	filePath := filepath.Join(UploadDir, filepath.Base(filename))
	cleanPath := filepath.Clean(filePath)
	if !strings.HasPrefix(cleanPath, UploadDir) {
		return "", fmt.Errorf("небезопасный путь")
	}

	// Создание файла с безопасными правами
	dst, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Копирование данных
	_, err = io.Copy(dst, src)
	if err != nil {
		os.Remove(filePath) // Удаляем файл при ошибке
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

	// Извлекаем имя файла из URL безопасно
	parts := strings.Split(avatarURL, "/")
	if len(parts) == 0 {
		return nil
	}
	filename := filepath.Base(parts[len(parts)-1])
	filePath := filepath.Join(UploadDir, filename)
	
	// Проверка безопасности пути
	cleanPath := filepath.Clean(filePath)
	if !strings.HasPrefix(cleanPath, UploadDir) {
		return fmt.Errorf("небезопасный путь")
	}

	// Удаляем файл если существует
	if _, err := os.Stat(filePath); err == nil {
		return os.Remove(filePath)
	}
	return nil
}

// Оптимизированная проверка типа файла
var validImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
}

func isValidImageType(contentType string) bool {
	return validImageTypes[contentType]
}

// generateSecureFilename - генерация безопасного имени файла
func generateSecureFilename() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func getFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return ".jpg" // дефолтное расширение
	}
	return strings.ToLower(ext)
}
