package utils

import (
	"regexp"
	"strings"
)

var (
	logSanitizeRegex = regexp.MustCompile(`[\r\n\t\x00-\x1f\x7f-\x9f]`)
)

// SanitizeForLog очищает строку от символов log injection
func SanitizeForLog(input string) string {
	if input == "" {
		return ""
	}
	
	sanitized := logSanitizeRegex.ReplaceAllString(input, "")
	
	// Корректная обработка UTF-8 - считаем руны, а не байты
	runes := []rune(sanitized)
	if len(runes) > 200 {
		sanitized = string(runes[:200]) + "..."
	}
	
	return sanitized
}

// SanitizeEmail маскирует email для логирования
func SanitizeEmail(email string) string {
	if email == "" {
		return ""
	}
	
	email = SanitizeForLog(email)
	
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***@***"
	}
	
	local := parts[0]
	domain := parts[1]
	
	// Корректная обработка UTF-8 для local части
	localRunes := []rune(local)
	if len(localRunes) <= 2 {
		return "***@" + domain
	}
	return string(localRunes[:2]) + "***@" + domain
}