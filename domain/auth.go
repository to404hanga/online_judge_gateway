package domain

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type InfoResponse struct {
	Username string `json:"username"`
	Realname string `json:"realname"`
	Role     int8   `json:"role"`
	Status   int8   `json:"status"`
}
