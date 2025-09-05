package captcha

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const TurnstileVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

type TurnstileService struct {
	secretKey string
	client    *http.Client
}

type TurnstileResponse struct {
	Success     bool     `json:"success"`
	ErrorCodes  []string `json:"error-codes,omitempty"`
	ChallengeTS string   `json:"challenge_ts,omitempty"`
	Hostname    string   `json:"hostname,omitempty"`
}

func NewTurnstileService(secretKey string) *TurnstileService {
	return &TurnstileService{
		secretKey: secretKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (t *TurnstileService) Verify(token, remoteIP string) (bool, error) {
	if t.secretKey == "" {
		// В разработке без секрета - пропускаем проверку
		return true, nil
	}

	data := url.Values{
		"secret":   {t.secretKey},
		"response": {token},
		"remoteip": {remoteIP},
	}

	resp, err := t.client.PostForm(TurnstileVerifyURL, data)
	if err != nil {
		return false, fmt.Errorf("ошибка запроса к Turnstile: %v", err)
	}
	defer resp.Body.Close()

	var result TurnstileResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("ошибка парсинга ответа: %v", err)
	}

	return result.Success, nil
}