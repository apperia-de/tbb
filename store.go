package tbb

import (
	"errors"
	"sync"
)

// ErrUserNotFound is returned when a user cannot be found in the store.
var ErrUserNotFound = errors.New("user not found")

// UserStore defines the storage interface for managing Telegram users.
type UserStore interface {
	FindUserByChatID(chatID int64) (*User, error)
	Save(user *User) error
}

// InMemoryStore is an in-memory implementation of UserStore, useful for
// development, testing, or simple deployments that don't need persistence.
type InMemoryStore struct {
	mu    sync.RWMutex
	users map[int64]*User
}

// NewInMemoryStore creates and returns a new InMemoryStore.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		users: make(map[int64]*User),
	}
}

// FindUserByChatID retrieves a user from the in-memory store by their Telegram ChatID.
func (s *InMemoryStore) FindUserByChatID(chatID int64) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[chatID]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// Save inserts or updates a user in the in-memory store.
func (s *InMemoryStore) Save(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[user.ChatID] = user
	return nil
}
