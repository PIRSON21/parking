package request

// UserLogin используется для валидации пользователя при авторизации менеджера
type UserLogin struct {
	Login    string `json:"login" validate:"required,min=4,max=8"`
	Password string `json:"password" validate:"required,min=4,max=10"`
}

// UserCreate используется для валидации тела при создании менеджера
type UserCreate struct {
	Login    string `json:"login" validate:"required,min=4,max=8"`
	Password string `json:"password" validate:"required,min=4,max=10"`
	Email    string `json:"email" validate:"required,email,min=8,max=15"`
}
