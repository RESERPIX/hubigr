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

	// Открытие файла для проверки
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Проверка магических байтов (первые 512 байт)
	buffer := make([]byte, 512)
	n, err := src.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("ошибка чтения файла: %v", err)
	}
	buffer = buffer[:n]

	// Проверка на реальный тип файла
	if !isValidImageByMagicBytes(buffer) {
		return "", fmt.Errorf("файл не является изображением JPEG/PNG")
	}

	// Проверка на вредоносное содержимое
	if containsMaliciousContent(buffer) {
		return "", fmt.Errorf("файл содержит подозрительное содержимое")
	}

	// Возврат к началу файла
	src.Seek(0, 0)

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

func getFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return ".jpg" // дефолтное расширение
	}
	return strings.ToLower(ext)
}
