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
	// Создаем директорию с безопасными правами (0700 - только владелец)
	os.MkdirAll(UploadDir, 0700)
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
	
	// Безопасное построение пути - используем только базовое имя файла
	safeFilename := filepath.Base(filename)
	absUploadDir, err := filepath.Abs(UploadDir)
	if err != nil {
		return "", fmt.Errorf("path resolution failed")
	}
	
	// Строим путь только из безопасных компонентов
	filePath := filepath.Join(absUploadDir, safeFilename)
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("path resolution failed")
	}
	
	// Критическая проверка - файл должен быть строго внутри директории
	if !strings.HasPrefix(absFilePath, absUploadDir+string(filepath.Separator)) {
		return "", fmt.Errorf("path traversal detected")
	}

	dst, err := os.OpenFile(absFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return "", fmt.Errorf("file creation failed")
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		os.Remove(absFilePath)
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
	// Используем только базовое имя файла для безопасности
	safeFilename := filepath.Base(parts[len(parts)-1])
	
	// Проверяем что имя файла соответствует нашему формату
	for _, r := range safeFilename {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '.') {
			return fmt.Errorf("invalid filename")
		}
	}
	
	absUploadDir, err := filepath.Abs(UploadDir)
	if err != nil {
		return fmt.Errorf("path resolution failed")
	}
	
	filePath := filepath.Join(absUploadDir, safeFilename)
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("path resolution failed")
	}
	
	if !strings.HasPrefix(absFilePath, absUploadDir+string(filepath.Separator)) {
		return fmt.Errorf("path traversal detected")
	}

	if _, err := os.Stat(absFilePath); err == nil {
		return os.Remove(absFilePath)
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

// Прекомпилированные паттерны для эффективности
var maliciousPatterns = [][]byte{
	[]byte("<?php"),
	[]byte("<script"),
	[]byte("javascript:"),
	[]byte("eval("),
	[]byte("exec("),
	[]byte("system("),
	[]byte("shell_exec("),
	[]byte("passthru("),
	[]byte("base64_decode("),
}

// containsMaliciousContent проверяет на подозрительное содержимое
func containsMaliciousContent(data []byte) bool {
	// Преобразуем в lowercase только однажды
	lowerData := bytes.ToLower(data)
	
	// Используем bytes.Contains вместо string conversion
	for _, pattern := range maliciousPatterns {
		if bytes.Contains(lowerData, pattern) {
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


