package fcm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// FCMService handles Firebase Cloud Messaging operations
type FCMService struct {
	client *messaging.Client
}

// NewFCMService creates a new FCM service
func NewFCMService(credentialsPath string) (*FCMService, error) {
	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %v", err)
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting messaging client: %v", err)
	}

	return &FCMService{
		client: client,
	}, nil
}

// SendNotification sends a push notification to a single device
func (s *FCMService) SendNotification(ctx context.Context, token string, title string, body string, data map[string]string) error {
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data:  data,
		Token: token,
	}

	response, err := s.client.Send(ctx, message)
	if err != nil {
		log.Printf("Error sending FCM message: %v", err)
		return err
	}

	log.Printf("Successfully sent FCM message: %s", response)
	return nil
}

// SendMulticastNotification sends a push notification to multiple devices
func (s *FCMService) SendMulticastNotification(ctx context.Context, tokens []string, title string, body string, data map[string]string) (*messaging.BatchResponse, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("no tokens provided")
	}

	message := &messaging.MulticastMessage{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data:   data,
		Tokens: tokens,
	}

	batchResponse, err := s.client.SendMulticast(ctx, message)
	if err != nil {
		log.Printf("Error sending multicast FCM message: %v", err)
		return nil, err
	}

	if batchResponse.FailureCount > 0 {
		for idx, resp := range batchResponse.Responses {
			if !resp.Success {
				log.Printf("Failed to send to token %s: %v", tokens[idx], resp.Error)
			}
		}
	}

	log.Printf("Successfully sent %d/%d FCM messages", batchResponse.SuccessCount, len(tokens))
	return batchResponse, nil
}

// SendTopicNotification sends a notification to all devices subscribed to a topic
func (s *FCMService) SendTopicNotification(ctx context.Context, topic string, title string, body string, data map[string]string) error {
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data:  data,
		Topic: topic,
	}

	response, err := s.client.Send(ctx, message)
	if err != nil {
		log.Printf("Error sending FCM topic message: %v", err)
		return err
	}

	log.Printf("Successfully sent FCM topic message: %s", response)
	return nil
}

// SubscribeToTopic subscribes tokens to a topic
func (s *FCMService) SubscribeToTopic(ctx context.Context, tokens []string, topic string) error {
	response, err := s.client.SubscribeToTopic(ctx, tokens, topic)
	if err != nil {
		log.Printf("Error subscribing to topic: %v", err)
		return err
	}

	if response.FailureCount > 0 {
		log.Printf("Failed to subscribe %d tokens to topic %s", response.FailureCount, topic)
	}

	return nil
}

// UnsubscribeFromTopic unsubscribes tokens from a topic
func (s *FCMService) UnsubscribeFromTopic(ctx context.Context, tokens []string, topic string) error {
	response, err := s.client.UnsubscribeFromTopic(ctx, tokens, topic)
	if err != nil {
		log.Printf("Error unsubscribing from topic: %v", err)
		return err
	}

	if response.FailureCount > 0 {
		log.Printf("Failed to unsubscribe %d tokens from topic %s", response.FailureCount, topic)
	}

	return nil
}

// PrayerCompletionNotification sends a notification when a prayer is completed
func (s *FCMService) PrayerCompletionNotification(ctx context.Context, tokens []string, prayerTitle string, memberName string, roomName string) error {
	title := "기도가 완료되었습니다 🙏"
	body := fmt.Sprintf("%s님이 '%s'에서 '%s' 기도를 완료했습니다", memberName, roomName, prayerTitle)

	data := map[string]string{
		"type":        "PRAYER_COMPLETION",
		"prayerTitle": prayerTitle,
		"memberName":  memberName,
		"roomName":    roomName,
	}

	_, err := s.SendMulticastNotification(ctx, tokens, title, body, data)
	return err
}

// RoomInvitationNotification sends a notification for room invitation
func (s *FCMService) RoomInvitationNotification(ctx context.Context, token string, inviterName string, roomName string) error {
	title := "새로운 초대가 있습니다 💌"
	body := fmt.Sprintf("%s님이 '%s' 방으로 초대했습니다", inviterName, roomName)

	data := map[string]string{
		"type":        "ROOM_INVITATION",
		"inviterName": inviterName,
		"roomName":    roomName,
	}

	return s.SendNotification(ctx, token, title, body, data)
}

// DailyReminderNotification sends a daily prayer reminder
func (s *FCMService) DailyReminderNotification(ctx context.Context, token string, roomName string) error {
	title := "오늘의 기도 시간입니다 ⏰"
	body := fmt.Sprintf("'%s' 방에서 함께 기도해요", roomName)

	data := map[string]string{
		"type":     "DAILY_REMINDER",
		"roomName": roomName,
	}

	return s.SendNotification(ctx, token, title, body, data)
}

// NotificationData represents the data structure for FCM notifications
type NotificationData struct {
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Body      string                 `json:"body"`
	ExtraData map[string]interface{} `json:"extraData,omitempty"`
}

// SendCustomNotification sends a custom notification with structured data
func (s *FCMService) SendCustomNotification(ctx context.Context, token string, notifData NotificationData) error {
	// Convert notification data to string map for FCM
	data := make(map[string]string)
	data["type"] = notifData.Type
	data["title"] = notifData.Title
	data["body"] = notifData.Body

	if notifData.ExtraData != nil {
		extraJSON, err := json.Marshal(notifData.ExtraData)
		if err == nil {
			data["extraData"] = string(extraJSON)
		}
	}

	return s.SendNotification(ctx, token, notifData.Title, notifData.Body, data)
}
