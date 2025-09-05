package validation

import "regexp"

// Pre-compiled regex patterns for optimal performance
// All patterns are compiled once at startup to avoid CPU bottlenecks
var (
	// Email validation pattern
	EmailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	
	// Password validation: 0-9, A-Z, a-z, special chars
	PasswordPattern = regexp.MustCompile(`^[0-9A-Za-z!"#$%&'()*+,./:;<=>?@\[\\\]^_{}-]+$`)
	
	// Nick validation: A-Z, a-z, А-Я, а-я
	NickPattern = regexp.MustCompile(`^[A-Za-zА-Яа-я]+$`)
	
	// URL validation: http/https URLs
	URLPattern = regexp.MustCompile(`^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(?:/[^\s]*)?$`)
	
	// Log sanitization: remove control characters
	LogSanitizePattern = regexp.MustCompile(`[\r\n\t\x00-\x1f\x7f-\x9f]`)
	
	// Authorization header pattern: "Bearer <token>"
	AuthHeaderPattern = regexp.MustCompile(`^Bearer\s+([A-Za-z0-9\-._~+/]+=*)$`)
)

// Fast validation functions using pre-compiled patterns
func IsValidEmail(email string) bool {
	return EmailPattern.MatchString(email)
}

func IsValidPassword(password string) bool {
	return PasswordPattern.MatchString(password)
}

func IsValidNick(nick string) bool {
	return NickPattern.MatchString(nick)
}

func IsValidURL(url string) bool {
	return URLPattern.MatchString(url)
}

func SanitizeForLog(input string) string {
	return LogSanitizePattern.ReplaceAllString(input, "")
}

func ExtractBearerToken(authHeader string) string {
	matches := AuthHeaderPattern.FindStringSubmatch(authHeader)
	if len(matches) == 2 {
		return matches[1]
	}
	return ""
}