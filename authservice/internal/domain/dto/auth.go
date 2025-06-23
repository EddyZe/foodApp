package dto

type LoginDto struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterDto struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" biding:"min=2,max=35"`
	LastName  string `json:"last_name" binding:"max=35"`
}

type TokensDto struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type ConfirmEmail struct {
	Code string `json:"code" binding:"required"`
}

type BanUser struct {
	UserId      int64  `json:"user_id" binding:"required"`
	Cause       string `json:"cause" binding:"required"`
	Days        int32  `json:"days"`
	IsFavourite bool   `json:"is_favourite,omitempty"`
}
