package http

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/RESERPIX/hubigr/internal/domain"
	"github.com/gofiber/fiber/v2"
)

const (
	UploadsDir = "./uploads"
	MaxFileSize = 10 << 20 // 10 MB
)

// SecureStaticHandler безопасно обслуживает статические файлы
func SecureStaticHandler(c *fiber.Ctx) error {
	requestedPath := c.Params("*")
	if requestedPath == "" {
		return c.Status(404).JSON(domain.NewError("not_found", "File not found"))
	}

	// Строгая валидация - только алфавитно-цифровые символы и слеши
	for _, r := range requestedPath {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '.' || r == '/' || r == '-') {
			return c.Status(403).JSON(domain.NewError("forbidden", "Invalid path"))
		}
	}

	// Проверка на path traversal
	if strings.Contains(requestedPath, "..") {
		return c.Status(403).JSON(domain.NewError("forbidden", "Path traversal detected"))
	}

	// Получаем абсолютный путь к uploads директории
	absUploadsDir, err := filepath.Abs(UploadsDir)
	if err != nil {
		return c.Status(500).JSON(domain.NewError("internal_error", "Server error"))
	}
	
	// Создаем полный путь
	fullPath := filepath.Join(absUploadsDir, requestedPath)
	
	// Критическая проверка - файл должен быть строго внутри uploads
	if !strings.HasPrefix(fullPath, absUploadsDir+string(filepath.Separator)) {
		return c.Status(403).JSON(domain.NewError("forbidden", "Access denied"))
	}

	fileInfo, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return c.Status(404).JSON(domain.NewError("not_found", "File not found"))
	}
	if err != nil {
		return c.Status(500).JSON(domain.NewError("internal_error", "File access error"))
	}
	
	if fileInfo.Size() > MaxFileSize {
		return c.Status(413).JSON(domain.NewError("file_too_large", "File too large"))
	}

	// Установка безопасных заголовков
	c.Set("X-Content-Type-Options", "nosniff")
	c.Set("X-Frame-Options", "DENY")
	
	// Определение Content-Type по расширению
	ext := strings.ToLower(filepath.Ext(fullPath))
	switch ext {
	case ".jpg", ".jpeg":
		c.Set("Content-Type", "image/jpeg")
	case ".png":
		c.Set("Content-Type", "image/png")
	case ".gif":
		c.Set("Content-Type", "image/gif")
	default:
		// Для неизвестных типов принудительно скачивание
		c.Set("Content-Type", "application/octet-stream")
		c.Set("Content-Disposition", "attachment")
	}

	// Отправка файла
	return c.SendFile(fullPath)
}