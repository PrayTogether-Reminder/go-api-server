package fcm_token

import "strings"

// RegisterFcmTokenRequest represents the request payload for registering an FCM token.
type RegisterFcmTokenRequest struct {
	FcmToken string `json:"fcmToken" binding:"required,notblank,max=512"`
}

// TrimmedToken removes leading/trailing spaces for safe persistence.
func (r *RegisterFcmTokenRequest) TrimmedToken() string {
	return strings.TrimSpace(r.FcmToken)
}

// DeleteFcmTokenRequest represents the request payload for deleting an FCM token.
type DeleteFcmTokenRequest struct {
	FcmToken string `json:"fcmToken" binding:"required,notblank,max=512"`
}

// TrimmedToken removes leading/trailing spaces for safe persistence.
func (r *DeleteFcmTokenRequest) TrimmedToken() string {
	return strings.TrimSpace(r.FcmToken)
}
