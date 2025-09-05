package upload

import (
	"bytes"
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
	// Создаем директорию с правами 0755 для веб-сервера
	os.MkdirAll(UploadDir, 0755)
	return &AvatarUploader{baseURL: baseURL}
}

// UploadAvatar загружает аватар согласно UC-1.2.1 (jpeg/png, до 2 МБ)
func (u *AvatarUploader) UploadAvatar(userID int64, file *multipart.FileHeader) (string, error) {
	if file.Size > MaxAvatarSize {
		return "", fmt.Errorf("file too large")
	}

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("file open failed")
	}
	defer src.Close()

	buffer := make([]byte, 512)
	n, err := src.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("file read failed")
	}
	buffer = buffer[:n]

	if !isValidImageByMagicBytes(buffer) {
		return "", fmt.Errorf("invalid image format")
	}

	if containsMaliciousContent(buffer) {
		return "", fmt.Errorf("invalid file content")
	}

	src.Seek(0, 0)

	randomID, err := generateSecureFilename()
	if err != nil {
		return "", fmt.Errorf("filename generation failed")
	}
	
	filename := fmt.Sprintf("%d_%s.jpg", userID, randomID)
	
	// Создаем абсолютный путь к upload директории
	absUploadDir, err := filepath.Abs(UploadDir)
	if err != nil {
		return "", fmt.Errorf("path resolution failed")
	}
	
	// Создаем полный путь к файлу БЕЗ использования пользовательского ввода
	filePath := filepath.Join(absUploadDir, filename)
	
	// Проверяем что результирующий путь находится в пределах upload директории
	if !strings.HasPrefix(filePath, absUploadDir+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid path")
	}

	dst, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return "", fmt.Errorf("file creation failed")
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("file write failed")
	}

	avatarURL := fmt.Sprintf("%s/uploads/avatars/%s", u.baseURL, filename)
	return avatarURL, nil
}

// DeleteAvatar удаляет старый аватар
func (u *AvatarUploader) DeleteAvatar(avatarURL string) error {
	if avatarURL == "" {
		return nil
	}

	parts := strings.Split(avatarURL, "/")
	if len(parts) == 0 {
		return nil
	}
	filename := filepath.Base(parts[len(parts)-1])
	
	// Проверяем что имя файла соответствует нашему формату
	for _, r := range filename {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '.') {
			return fmt.Errorf("invalid filename")
		}
	}
	
	absUploadDir, err := filepath.Abs(UploadDir)
	if err != nil {
		return fmt.Errorf("path resolution failed")
	}
	
	filePath := filepath.Join(absUploadDir, filename)
	
	if !strings.HasPrefix(filePath, absUploadDir+string(filepath.Separator)) {
		return fmt.Errorf("invalid path")
	}

	if _, err := os.Stat(filePath); err == nil {
		return os.Remove(filePath)
	}
	return nil
}

// Магические байты для проверки типов файлов
var (
	jpegMagic1 = []byte{0xFF, 0xD8, 0xFF}
	jpegMagic2 = []byte{0xFF, 0xD8, 0xFF, 0xE0}
	jpegMagic3 = []byte{0xFF, 0xD8, 0xFF, 0xE1}
	pngMagic   = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
)

// isValidImageByMagicBytes проверяет файл по магическим байтам
func isValidImageByMagicBytes(data []byte) bool {
	if len(data) < 8 {
		return false
	}

	// Проверка PNG
	if bytes.HasPrefix(data, pngMagic) {
		return true
	}

	// Проверка JPEG (несколько вариантов)
	if bytes.HasPrefix(data, jpegMagic1) || 
	   bytes.HasPrefix(data, jpegMagic2) || 
	   bytes.HasPrefix(data, jpegMagic3) {
		return true
	}

	return false
}

// containsMaliciousContent проверяет на подозрительное содержимое
func containsMaliciousContent(data []byte) bool {
	// Список подозрительных строк
	maliciousPatterns := []string{
		"<?php",
		"<script",
		"javascript:",
		"eval(",
		"exec(",
		"system(",
		"shell_exec(",
		"passthru(",
		"base64_decode(",
	}

	dataStr := strings.ToLower(string(data))
	for _, pattern := range maliciousPatterns {
		if strings.Contains(dataStr, pattern) {
			return true
		}
	}
	return false
}

// generateSecureFilename - генерация безопасного имени файла
func generateSecureFilename() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}


