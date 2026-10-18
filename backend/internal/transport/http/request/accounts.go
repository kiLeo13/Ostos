package request

type UserCreateRequest struct {
	Email       string `json:"email" validate:"required,email,max=254"`
	Username    string `json:"username" validate:"required,username"`
	DisplayName string `json:"display_name" validate:"required,notblank,min=1,max=100"`
	Password    string `json:"password" validate:"required,min=8,max=128"`
}

type UserUpdateRequest struct {
	Username    *string `json:"username" validate:"omitempty,username"`
	DisplayName *string `json:"display_name" validate:"omitempty,notblank,min=1,max=100"`
}

type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,max=128"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=128"`
}

type AccountDeleteRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,max=128"`
}

type LoginRequest struct {
	Identifier string  `json:"identifier" validate:"required,max=254"`
	Password   string  `json:"password" validate:"required,max=128"`
	DeviceName *string `json:"device_name" validate:"omitempty,max=100"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,max=200"`
}

type AccountTokenRequest struct {
	Token string `json:"token" validate:"required,max=200"`
}

type EmailRequest struct {
	Email string `json:"email" validate:"required,email,max=254"`
}

type PasswordResetRequest struct {
	Token       string `json:"token" validate:"required,max=200"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
}
