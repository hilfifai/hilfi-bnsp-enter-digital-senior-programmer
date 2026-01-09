package dto

type LoginRequest struct {
	Username string `json:"username" binding:"required" validate:"required,min=3,max=100"`
	Password string `json:"password" binding:"required" validate:"required,min=6,max=100"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" validate:"required"`
}
