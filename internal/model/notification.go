package model

// Notification represents the base notification structure (embedded in specific notification types)
// This is NOT a separate table - fields are embedded into child structs
type Notification struct {
	ID               int64  `gorm:"column:id;primaryKey;autoIncrement"`
	SenderID         int64  `gorm:"column:sender_id;not null"`                  // member who sent the notification
	RecipientID      int64  `gorm:"column:recipient_id;not null"`               // member who receives notification
	Message          string `gorm:"column:message;type:VARCHAR2(500);not null"` // notification message
	NotificationType string `gorm:"column:notification_type;type:VARCHAR2(50);not null;default:'PRAYER_COMPLETION'"`

	BaseEntity
}

// PrayerCompletionNotification represents an in-app notification for prayer completion
// Stores notification history that can be viewed in the app's notification center
type PrayerCompletionNotification struct {
	Notification  `gorm:"embedded"` // Embed base notification fields
	PrayerTitleID int64             `gorm:"column:prayer_title_id;not null"` // NOT a foreign key - preserves history even if title is deleted
}

// TableName specifies the table name for PrayerCompletionNotification
func (*PrayerCompletionNotification) TableName() string {
	return "prayer_completion_notification"
}

// NewPrayerCompletionNotification creates a new PrayerCompletionNotification instance
func NewPrayerCompletionNotification(senderID, recipientID, prayerTitleID int64, message string) *PrayerCompletionNotification {
	return &PrayerCompletionNotification{
		Notification: Notification{
			SenderID:         senderID,
			RecipientID:      recipientID,
			Message:          message,
			NotificationType: "PRAYER_COMPLETION",
		},
		PrayerTitleID: prayerTitleID,
	}
}
