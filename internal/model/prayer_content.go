package model

// PrayerContent represents a prayer content (like a comment) for a prayer title
// Oracle IDENTITY (auto-increment) is used for ID generation
type PrayerContent struct {
	ID            int64        `gorm:"column:id;primaryKey;autoIncrement"`
	PrayerTitleID int64        `gorm:"column:prayer_title_id;not null;index"`
	PrayerTitle   *PrayerTitle `gorm:"foreignKey:PrayerTitleID;constraint:OnDelete:CASCADE"`
	WriterID      int64        `gorm:"column:writer_id;not null;index"`
	Writer        *Member      `gorm:"foreignKey:WriterID"`
	MemberID      *int64       `gorm:"column:member_id;index"` // nullable - 기도를 요청한 멤버
	MemberName    string       `gorm:"column:member_name;type:VARCHAR2(255);not null"`
	Content       string       `gorm:"column:content;type:CLOB;not null"`

	BaseEntity
}

// TableName specifies the table name for PrayerContent
func (*PrayerContent) TableName() string {
	return "prayer_content"
}

// NewPrayerContent creates a new PrayerContent instance
// Factory method pattern
func NewPrayerContent(
	prayerTitle *PrayerTitle,
	writer *Member,
	memberID *int64,
	memberName string,
	content string,
) *PrayerContent {
	return &PrayerContent{
		PrayerTitle:   prayerTitle,
		PrayerTitleID: prayerTitle.ID,
		Writer:        writer,
		WriterID:      writer.ID,
		MemberID:      memberID,
		MemberName:    memberName,
		Content:       content,
	}
}

// UpdateContent updates the content
func (pc *PrayerContent) UpdateContent(content string) {
	pc.Content = content
}
