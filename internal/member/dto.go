package member

type FetchProfileResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
}

type UpdateProfileRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=10"`
	PhoneNumber *string `json:"phoneNumber" binding:"omitempty,phone"`
}
