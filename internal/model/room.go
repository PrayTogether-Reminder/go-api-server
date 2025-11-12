package model

// Room represents a prayer room in the system
// Oracle IDENTITY (auto-increment) is used for ID generation
type Room struct {
	ID uint32 `gorm:"column:id;primaryKey;autoIncrement"`

	// Core fields
	Name        string `gorm:"column:name;type:VARCHAR2(50);not null"`         // 방 이름
	Description string `gorm:"column:description;type:VARCHAR2(255);not null"` // 방 설명

	BaseEntity
}

// TableName specifies the table name for Room
func (*Room) TableName() string {
	return "room"
}

// NewRoom creates a new Room instance
// Factory method pattern (Java의 static create 메서드와 동일)
func NewRoom(name, description string) *Room {
	return &Room{
		Name:        name,
		Description: description,
	}
}
