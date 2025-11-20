package model

// PrayerTitle represents a prayer title in the system
// Oracle IDENTITY (auto-increment) is used for ID generation
type PrayerTitle struct {
	ID     int64  `gorm:"column:id;primaryKey;autoIncrement"`
	RoomID int64  `gorm:"column:room_id;not null;index"`
	Room   *Room  `gorm:"foreignKey:RoomID"`
	Title  string `gorm:"column:title;type:VARCHAR2(255);not null"` // Java의 TITLE_ENTITY_MAX_LEN 값에 따라 조정 필요

	BaseEntity

	// OneToMany relationship - PrayerContent 모델 생성 후 활성화
	// PrayerContents []PrayerContent `gorm:"foreignKey:PrayerTitleID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for PrayerTitle
func (*PrayerTitle) TableName() string {
	return "prayer_title"
}

// NewPrayerTitle creates a new PrayerTitle instance
// Factory method pattern (Java의 static create 메서드와 동일)
func NewPrayerTitle(room *Room, title string) *PrayerTitle {
	return &PrayerTitle{
		Room:   room,
		RoomID: room.ID,
		Title:  title,
	}
}

// UpdateTitle updates the title
func (pt *PrayerTitle) UpdateTitle(title string) {
	pt.Title = title
}
