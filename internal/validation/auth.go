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
	urlRegex      = regexp.MustCompile(`^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(?:/[^\s]*)?$`)
)

func ValidateSignUp(req domain.SignUpRequest) []string {
	var errors []string

	if !emailRegex.MatchString(req.Email) {
		errors = append(errors, "Email должен соответствовать формату example@example.ru")
	}

	if l := utf8.RuneCountInString(req.Password); l < 6 || l > 20 {
		errors = append(errors, "Пароль должен содержать от 6 до 20 символов")
	}
	if !passwordRegex.MatchString(req.Password) {
		errors = append(errors, "Пароль содержит недопустимые символы")
	}

	if req.Password != req.ConfirmPassword {
		errors = append(errors, "Пароли должны совпадать")
	}

	nick := strings.TrimSpace(req.Nick)
	if l := utf8.RuneCountInString(nick); l < 2 || l > 50 {
		errors = append(errors, "Ник должен содержать от 2 до 50 символов")
	}
	if !nickRegex.MatchString(nick) {
		errors = append(errors, "Ник может содержать только буквы")
	}

	if !req.AgreeTerms {
		errors = append(errors, "Необходимо согласиться с Условиями и Политикой конфиденциальности")
	}

	return errors
}

func ValidateProfile(req domain.UpdateProfileRequest) []string {
	var errors []string

	nick := strings.TrimSpace(req.Nick)
	if l := utf8.RuneCountInString(nick); l < 2 || l > 50 {
		errors = append(errors, "Ник должен содержать от 2 до 50 символов")
	}
	if !nickRegex.MatchString(nick) {
		errors = append(errors, "Ник может содержать только буквы")
	}

	if req.Bio != nil && utf8.RuneCountInString(*req.Bio) > 200 {
		errors = append(errors, "Био не должно превышать 200 символов")
	}

	if len(req.Links) > 5 {
		errors = append(errors, "Максимум 5 ссылок")
	}

	for _, link := range req.Links {
		if !urlRegex.MatchString(link.URL) {
			errors = append(errors, "Неверный формат ссылки: "+link.URL)
		}
	}

	return errors
}