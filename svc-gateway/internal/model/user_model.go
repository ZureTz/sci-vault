package model

// Models for user authentication (login and registration)

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20,custom_username_validator"`
	Password string `json:"password" binding:"required,min=6,max=50,custom_password_validator"`
}

type LoginResponse struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	JWTToken string `json:"token"`
}

type RegisterRequest struct {
	Username          string `json:"username" binding:"required,min=3,max=20,custom_username_validator"`
	Password          string `json:"password" binding:"required,min=6,max=50,custom_password_validator"`
	ConfirmedPassword string `json:"confirmed_password" binding:"required,eqfield=Password"`
	// TODO: For future email verification feature
	// Email             string `json:"email" binding:"required,email"`
	// EmailCode         string `json:"email_code" binding:"required,len=6,numeric"`
}
