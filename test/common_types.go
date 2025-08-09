package test

// MessageResponse represents a common message response (matching Java MessageResponse)
type MessageResponse struct {
	Message string `json:"message"`
}

// PrayerContentWithMemberInfo represents prayer content with member info for tests
// This is used to match Java's test structure which includes member info
type PrayerContentWithMemberInfo struct {
	ID            uint64 `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	PrayerTitleID uint64 `gorm:"column:prayer_title_id;not null;index" json:"prayerTitleId"`
	Content       string `gorm:"column:content;type:text;not null" json:"content"`
	MemberID      uint64 `gorm:"column:member_id;not null;index" json:"memberId"`
	MemberName    string `gorm:"column:member_name;not null" json:"memberName"`
}

// TableName specifies the table name for PrayerContentWithMemberInfo
func (PrayerContentWithMemberInfo) TableName() string {
	return "prayer_content"
}
