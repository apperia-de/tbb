package tbb_test

import (
	"testing"

	"github.com/apperia-de/tbb"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryStore(t *testing.T) {
	store := tbb.NewInMemoryStore()

	// Find non-existent user
	user, err := store.FindUserByChatID(12345)
	assert.ErrorIs(t, err, tbb.ErrUserNotFound)
	assert.Nil(t, user)

	// Save user
	expectedUser := &tbb.User{
		ChatID:    12345,
		Username:  "testuser",
		Firstname: "Test",
		UserInfo: &tbb.UserInfo{
			IsActive: true,
		},
	}
	err = store.Save(expectedUser)
	assert.NoError(t, err)

	// Find saved user
	user, err = store.FindUserByChatID(12345)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
}
