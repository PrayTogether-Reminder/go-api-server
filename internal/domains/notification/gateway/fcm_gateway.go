package gateway

import (
	"context"
	"fmt"
	"log"
)

// FCMGateway represents FCM notification gateway
type FCMGateway struct {
	// In production, this would use firebase-admin-go SDK
	// For now, we'll just log the notifications
	projectID string
}

// NewFCMGateway creates a new FCM gateway
func NewFCMGateway(projectID string) *FCMGateway {
	return &FCMGateway{
		projectID: projectID,
	}
}

// SendNotification sends a push notification via FCM
func (g *FCMGateway) SendNotification(ctx context.Context, tokens []string, title, body string, data map[string]interface{}) error {
	// In production, this would use Firebase Admin SDK:
	//
	// import firebase "firebase.google.com/go/v4"
	// import "firebase.google.com/go/v4/messaging"
	//
	// app, err := firebase.NewApp(ctx, nil)
	// client, err := app.Messaging(ctx)
	//
	// message := &messaging.MulticastMessage{
	//     Notification: &messaging.Notification{
	//         Title: title,
	//         Body:  body,
	//     },
	//     Data:   data,
	//     Tokens: tokens,
	// }
	//
	// response, err := client.SendMulticast(ctx, message)

	// For now, just log the notification
	log.Printf("FCM Notification: title=%s, body=%s, tokens=%v, data=%v", title, body, tokens, data)

	// Simulate some tokens being invalid (for testing)
	if len(tokens) > 0 && tokens[0] == "invalid_token" {
		return fmt.Errorf("invalid FCM token")
	}

	return nil
}

// SendPrayerCompletionNotification sends a prayer completion notification
func (g *FCMGateway) SendPrayerCompletionNotification(ctx context.Context, tokens []string, message string, prayerTitleID uint64) error {
	data := map[string]interface{}{
		"type":          "PRAYER_COMPLETION",
		"prayerTitleId": fmt.Sprintf("%d", prayerTitleID),
	}

	return g.SendNotification(ctx, tokens, "기도 완료 알림", message, data)
}

// ValidateToken validates an FCM token
func (g *FCMGateway) ValidateToken(ctx context.Context, token string) error {
	// In production, this would validate the token with FCM
	// For now, just check if it's not empty
	if token == "" {
		return fmt.Errorf("empty FCM token")
	}

	if token == "invalid_token" {
		return fmt.Errorf("invalid FCM token")
	}

	return nil
}
