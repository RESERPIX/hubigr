package validation

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/RESERPIX/hubigr/internal/domain"
)

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	passwordRegex = regexp.MustCompile(`^[0-9A-Za-z!"#$%&'()*+,./:;<=>?@\[\\\]^_{}-]+$`)
	nickRegex     = regexp.MustCompile(`^[A-Za-zА-Яа-я]+$`)
)

// ValidateSignUp - UC-1.1.1 из ТЗ
func ValidateSignUp(req domain.SignUpRequest) []string {
	var errors []string

	// Email validation
	if !emailRegex.MatchString(req.Email) {
		errors = append(errors, "Email должен соответствовать формату example@example.ru")
	}

	// Password validation (6-20 символов, 0-9, A-Z, a-z, допустимые символы)
	if l := utf8.RuneCountInString(req.Password); l < 6 || l > 20 {
		errors = append(errors, "Пароль должен содержать от 6 до 20 символов")
	}
	if !passwordRegex.MatchString(req.Password) {
		errors = append(errors, "Пароль содержит недопустимые символы")
	}

	// Password confirmation
	if req.Password != req.ConfirmPassword {
		errors = append(errors, "Пароли должны совпадать")
	}

	// Nick validation (2-50 символов, A-Z, a-z, А-Я, а-я)
	nick := strings.TrimSpace(req.Nick)
	if l := utf8.RuneCountInString(nick); l < 2 || l > 50 {
		errors = append(errors, "Ник должен содержать от 2 до 50 символов")
	}
	if !nickRegex.MatchString(nick) {
		errors = append(errors, "Ник может содержать только буквы")
	}

	// Terms agreement
	if !req.AgreeTerms {
		errors = append(errors, "Необходимо согласиться с Условиями и Политикой конфиденциальности")
	}

	return errors
}

// ValidateProfile - UC-1.2.1 из ТЗ
func ValidateProfile(req domain.UpdateProfileRequest) []string {
	var errors []string

	// Nick validation
	nick := strings.TrimSpace(req.Nick)
	if l := utf8.RuneCountInString(nick); l < 2 || l > 50 {
		errors = append(errors, "Ник должен содержать от 2 до 50 символов")
	}
	if !nickRegex.MatchString(nick) {
		errors = append(errors, "Ник может содержать только буквы")
	}

	// Bio validation (до 200 символов)
	if req.Bio != nil && utf8.RuneCountInString(*req.Bio) > 200 {
		errors = append(errors, "Био не должно превышать 200 символов")
	}

	// Links validation (до 5 ссылок)
	if len(req.Links) > 5 {
		errors = append(errors, "Максимум 5 ссылок")
	}

	for _, link := range req.Links {
		if !strings.HasPrefix(link.URL, "http://") && !strings.HasPrefix(link.URL, "https://") {
			errors = append(errors, "Неверный формат ссылки: "+link.URL)
		}
	}

	return errors
}