package domain

import "time"

type User struct {
	ID      int64     `json:"id"`
	Email   string    `json:"email"`
	Nick    string    `json:"nick"`
	Role    string    `json:"role"`
	Avatar  *string   `json:"avatarUrl,omitempty"`
	Bio     *string   `json:"bio,omitempty"`
	Links   []string  `json:"links,omitempty"`
	Banned  bool      `json:"is_banned"`
	Created time.Time `json:"created_at"`
}

type ErrorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func NewError(code, msg string) ErrorEnvelope {
	var e ErrorEnvelope
	e.Error.Code = code
	e.Error.Message = msg
	return e
}
