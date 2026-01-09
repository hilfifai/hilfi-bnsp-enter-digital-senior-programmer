package dto

type TokenResponse struct {
	Data         any    `json:"data"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
