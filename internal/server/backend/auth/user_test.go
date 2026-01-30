package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

func TestUserCRUD(t *testing.T) {
	m := newTestManager(t)

	// Create User (via Register or CreateUser)
	// CreateUser is lower level.
	user, err := m.CreateUser("user1", auth.WithGroup("group1"))
	assert.NoError(t, err)
	assert.Equal(t, "user1", user.Name)
	assert.Len(t, user.Groups, 1)
	assert.Equal(t, "group1", user.Groups[0].Name)

	// Get User
	gotUser, err := m.GetUser("user1")
	assert.NoError(t, err)
	assert.Equal(t, "user1", gotUser.Name)

	// List User
	users, err := m.ListUser()
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "user1", users[0].Name)

	// Update User
	gotUser.Labels = map[string]string{"key": "value"}
	err = m.UpdateUser(gotUser)
	assert.NoError(t, err)

	gotUser, err = m.GetUser("user1")
	assert.NoError(t, err)
	assert.Equal(t, "value", gotUser.Labels["key"])

	// Delete User
	err = m.DeleteUser("user1")
	assert.NoError(t, err)

	_, err = m.GetUser("user1")
	assert.Error(t, err)
}

func TestSubjects(t *testing.T) {
	m := newTestManager(t)
	user, _ := m.CreateUser("user1", auth.WithGroup("group1"))

	subjects := user.Subjects()
	// Should contain user subject and group subject
	// user:kind=user,name=user1
	// group:kind=group,name=group1

	hasUserSubject := false
	hasGroupSubject := false

	for _, s := range subjects {
		if s.Kind == auth.KindUser && s.Name == "user1" {
			hasUserSubject = true
		}
		if s.Kind == auth.KindGroup && s.Name == "group1" {
			hasGroupSubject = true
		}
	}
	assert.True(t, hasUserSubject)
	assert.True(t, hasGroupSubject)
}
