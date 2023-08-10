package res

import "time"

type TokenResponse struct {
	Jwt       string    `json:"jwt"`
	ExpiresAt time.Time `json:"expires_at"`
}
