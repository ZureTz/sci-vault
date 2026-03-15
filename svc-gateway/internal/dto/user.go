package dto

// DTOs for user authentication (login and registration)

type SendEmailCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type LoginRequest struct {
	// Mutex validation: either username or email must be provided, but not both
	// If both are provided, the username will be used for lookup. If only email is provided, it will be used for lookup.
	Username string `json:"username" binding:"required_without=Email,omitempty,min=3,max=20,custom_username_validator"`
	Email    string `json:"email" binding:"required_without=Username,omitempty,email"`
	Password string `json:"password" binding:"required,min=6,max=50,custom_password_validator"`
}

type LoginResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	JWTToken string `json:"token"`
}

type RegisterRequest struct {
	Username          string `json:"username" binding:"required,min=3,max=20,custom_username_validator"`
	Email             string `json:"email" binding:"required,email"`
	Password          string `json:"password" binding:"required,min=6,max=50,custom_password_validator"`
	ConfirmedPassword string `json:"confirmed_password" binding:"required,eqfield=Password"`
	EmailCode         string `json:"email_code" binding:"required,len=6,numeric"`
}

type ResetPasswordRequest struct {
	Email             string `json:"email" binding:"required,email"`
	EmailCode         string `json:"email_code" binding:"required,len=6,numeric"`
	Password          string `json:"password" binding:"required,min=6,max=50,custom_password_validator"`
	ConfirmedPassword string `json:"confirmed_password" binding:"required,eqfield=Password"`
}
