package tbb

import (
	"time"
)

type User struct {
	ID                      uint64     `gorm:"primaryKey" json:"-"`
	Username                string     `json:"username"`
	Firstname               string     `json:"firstname"`
	Lastname                string     `json:"lastname"`
	ChatID                  int64      `gorm:"uniqueIndex" json:"chatID"` // Telegram chatID of the user
	LanguageCode            string     `json:"language_code,omitempty"`   // Language code of the user
	IsBot                   bool       `json:"isBot"`                     // True if the user is itself a Bot
	IsPremium               bool       `json:"isPremium,omitempty"`       // True, if this user is a Telegram Premium user
	AddedToAttachmentMenu   bool       `json:"addedToAttachmentMenu,omitempty"`
	CanJoinGroups           bool       `json:"canJoinGroups,omitempty"`
	CanReadAllGroupMessages bool       `json:"canReadAllGroupMessages,omitempty"`
	SupportsInlineQueries   bool       `json:"supportsInlineQueries,omitempty"`
	CanConnectToBusiness    bool       `json:"canConnectToBusiness,omitempty"`
	HasMainWebApp           bool       `json:"hasMainWebApp,omitempty"`
	UserInfo                *UserInfo  `json:"userInfo,omitempty"`
	UserPhoto               *UserPhoto `json:"-"`
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

type UserInfo struct {
	UserID   uint64 `gorm:"primaryKey"`
	IsActive bool   `json:"isActive"`
	Status   string `json:"status,omitempty"` // Either "member" or "kicked"
	TimeZoneInfo
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

type File struct {
	ID        int64  `gorm:"primaryKey" json:"id"`
	Name      string // Filename
	UniqueID  string // Unique file ID from Telegram
	Extension string // File extension with the preceding "." if exists or empty otherwise
	MimeType  string // Mime type of the stored file
	Hash      string // MD5 hash of the file binary data
	Size      int64  // File size of the binary data
	Data      []byte // Binary file data
	CreatedAt time.Time
	UpdatedAt time.Time
}
