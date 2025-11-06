package member

type SignupRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=20"`
	Email       string `json:"email" binding:"required,email,max=50"`
	PhoneNumber string `json:"phoneNumber" binding:"required,phone"`
	Password    string `json:"password" binding:"required,min=8,max=15"`
}
