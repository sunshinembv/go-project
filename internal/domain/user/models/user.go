package models

type User struct {
	UID      string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"    validate:"email"`
	Password string `json:"password" validate:"min=8"`
}

type UserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserUpdateRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
