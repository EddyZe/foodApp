package auth

type LoginDto struct {
	Email    string `json:"email" required:"true"`
	Password string `json:"password" required:"true"`
}

type RegisterDto struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type TokensDto struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
