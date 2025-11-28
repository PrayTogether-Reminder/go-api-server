package model

// PrayerCompletion represents a prayer completion record
// Records when a member completes prayer for a prayer title
type PrayerCompletion struct {
	ID            int64        `gorm:"column:id;primaryKey;autoIncrement"`
	PrayerTitleID *int64       `gorm:"column:prayer_title_id"` // nullable - SET NULL on delete
	PrayerTitle   *PrayerTitle `gorm:"foreignKey:PrayerTitleID;constraint:OnDelete:SET NULL"`
	PrayerID      int64        `gorm:"column:prayer_id;not null"` // member ID who completed prayer

	BaseEntity
}

// TableName specifies the table name for PrayerCompletion
func (*PrayerCompletion) TableName() string {
	return "prayer_completion"
}

// NewPrayerCompletion creates a new PrayerCompletion instance
func NewPrayerCompletion(prayerID int64, prayerTitle *PrayerTitle) *PrayerCompletion {
	return &PrayerCompletion{
		PrayerID:      prayerID,
		PrayerTitle:   prayerTitle,
		PrayerTitleID: &prayerTitle.ID,
	}
}
