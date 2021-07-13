package twitch

import (
	"net/http"
	"time"
)

const (
	GetUsersUrl = "https://api.twitch.tv/helix/users"
)

type User struct {
	Id              string    `json:"id"`
	Login           string    `json:"login"`
	DisplayName     string    `json:"display_name"`
	Type            string    `json:"type"`
	BroadcasterType string    `json:"broadcaster_type"`
	Description     string    `json:"description"`
	ProfileImageUrl string    `json:"profile_image_url"`
	OfflineImageUrl string    `json:"offline_image_url"`
	ViewCount       int       `json:"view_count"`
	Email           string    `json:"email"`
	CreatedAt       time.Time `json:"created_at"`
}

// GetMe will return the user indicated by the accessToken.
func GetMe(clientId, accessToken string, timeout time.Duration) (*User, error) {
	res, err := Run(&Request{
		Method: http.MethodGet,
		Url: GetUsersUrl,
		ClientId: clientId,
		AccessToken: accessToken,
		Timeout: timeout,
	})
	if err != nil {
		return nil, err
	}

	var x struct {
		Data []*User `json:"data"`
	}
	if err := DecodeResponse(res, &x); err != nil {
		return nil, err
	}

	if len(x.Data) == 0 {
		return nil, ErrNotFound
	}
	return x.Data[0], nil
}
