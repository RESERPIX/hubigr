package domain

import "time"

// User - DM-3.1 из ТЗ
type User struct {
	ID              int64                  `json:"id"`
	Email           string                 `json:"email"`
	Hash            string                 `json:"-"`
	Role            Role                   `json:"role"`
	Nick            string                 `json:"nick"`
	Avatar          *string                `json:"avatar,omitempty"`
	Bio             *string                `json:"bio,omitempty"`
	Links           []Link                 `json:"links,omitempty"`
	IsBanned        bool                   `json:"is_banned"`
	EmailVerified   bool                   `json:"email_verified"`
	CreatedAt       time.Time              `json:"created_at"`
	PrivacySettings PrivacySettings        `json:"privacy_settings"`
}

type Role string

const (
	RoleParticipant Role = "participant"
	RoleJury        Role = "jury"
	RoleModerator   Role = "moderator"
	RoleAdmin       Role = "admin"
	RoleOrganizer   Role = "organizer"
)

type Link struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type PrivacySettings struct {
	ShowFollowers *bool `json:"show_followers,omitempty"`
	ShowFollowing *bool `json:"show_following,omitempty"`
}

// NotificationSettings - DM-3.15 из ТЗ
type NotificationSettings struct {
	UserID   int64    `json:"user_id"`
	NewGame  bool     `json:"new_game"`
	NewBuild bool     `json:"new_build"`
	NewPost  bool     `json:"new_post"`
	Channels Channels `json:"channels"`
}

type Channels struct {
	InApp bool  `json:"in_app"`
	Email *bool `json:"email,omitempty"`
}

// DefaultChannels возвращает дефолтные настройки каналов
func DefaultChannels() Channels {
	return Channels{
		InApp: true,
	}
}

// DefaultNotificationSettings возвращает дефолтные настройки уведомлений
func DefaultNotificationSettings(userID int64) NotificationSettings {
	return NotificationSettings{
		UserID:   userID,
		NewGame:  true,
		NewBuild: true,
		NewPost:  true,
		Channels: DefaultChannels(),
	}
}

// API Request/Response types
type SignUpRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	Nick            string `json:"nick"`
	AgreeTerms      bool   `json:"agree_terms"`
	CaptchaToken    string `json:"captcha_token"`
}

type LoginRequest struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	CaptchaToken string `json:"captcha_token"`
}

type AuthResponse struct {
	User         User   `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UpdateProfileRequest struct {
	Nick            string          `json:"nick"`
	Avatar          *string         `json:"avatar"`
	Bio             *string         `json:"bio"`
	Links           []Link          `json:"links"`
	PrivacySettings PrivacySettings `json:"privacy_settings"`
}

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// RefreshToken модель для refresh токенов
type RefreshToken struct {
	ID         int64      `json:"id"`
	UserID     int64      `json:"user_id"`
	TokenHash  string     `json:"-"`
	ExpiresAt  time.Time  `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	DeviceInfo *string    `json:"device_info,omitempty"`
	IPAddress  *string    `json:"ip_address,omitempty"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func NewError(code, message string) ErrorResponse {
	var e ErrorResponse
	e.Error.Code = code
	e.Error.Message = message
	return e
}