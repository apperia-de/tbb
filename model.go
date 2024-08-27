package tbb

import (
	"time"
)

type User struct {
	ID           uint64     `gorm:"primaryKey" json:"-"`
	Username     string     `json:"username"`
	Firstname    string     `json:"firstname"`
	Lastname     string     `json:"lastname"`
	IsBot        bool       `json:"isBot"`                // True if user is itself a Bot
	ChatID       int64      `gorm:"unique" json:"chatID"` // Telegram chatID of the user
	LanguageCode string     `json:"language_code"`        // Language code of the user
	IsPremium    bool       `json:"isPremium"`            // True, if this user is a Telegram Premium user
	UserInfo     *UserInfo  `json:"userInfo,omitempty"`
	UserPhoto    *UserPhoto `json:"-"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserInfo struct {
	UserID    uint64 `gorm:"primaryKey"`
	IsActive  bool
	Status    string  // Either "member" or "kicked"
	Latitude  float64 // Latitude the user sends for determining the users current Time zone
	Longitude float64 // Longitude the user sends for determining the users current Time zone
	Location  string  // Location of the user's timezone
	ZoneName  string  // Zone name of the user's timezone
	TZOffset  *int    // Time zone offset in seconds
	IsDST     bool    // Whether the TZOffset is in daylight saving time or normal time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserPhoto struct {
	UserID       uint64 `gorm:"primaryKey"` // ID from User table to whom the photo belongs to.
	FileID       string // Identifier for this file, which can be used to download or reuse the file
	FileUniqueID string // Unique identifier for this file, which is supposed to be the same over time and for different bots. Can't be used to download or reuse the file.
	FileSize     int    // Size in bytes of the user photo
	FileHash     string // The md5 file hash of the user photo
	FileData     []byte // The binary Data of the user photo
	Width        int    // Photo width
	Height       int    // Photo height
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
