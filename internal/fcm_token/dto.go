package fcm_token

// RegisterFcmTokenRequest represents the request payload for registering an FCM token.
type RegisterFcmTokenRequest struct {
	FcmToken string `json:"fcmToken" binding:"required,notblank,max=512"`
}
