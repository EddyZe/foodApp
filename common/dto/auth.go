package dto

type LoginDto struct {
	Email    string `json:"email" required:"true"`
	Password string `json:"password" required:"true"`
}

type RegisterDto struct {
	Email     string `json:"email" required:"true"`
	Password  string `json:"password" required:"true"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
