package memory_test

import (
	"context"
	"os"
	"testing"

	"github.com/humper/tor_exit_nodes/models"
	"github.com/humper/tor_exit_nodes/pkg/database"
	"github.com/humper/tor_exit_nodes/pkg/database/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var db *database.Database
var ctx context.Context

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error

	db, err = memory.New(ctx)
	if err != nil {
		panic(err)
	}

	code := m.Run()
	os.Exit(code)
}

func makeTestUser() *models.User {
	return &models.User{
		Name:       "Test Person",
		Email:      "test@example.com",
		Password:   "password",
		Role:       "admin",
		AllowedIPs: []string{},
	}
}

func compareUsers(t *testing.T, expected, actual *models.User) {
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Email, actual.Email)
	assert.Equal(t, expected.Password, actual.Password)
	assert.Equal(t, expected.Role, actual.Role)
	assert.EqualValues(t, expected.AllowedIPs, actual.AllowedIPs)
}

func TestCreateUser(t *testing.T) {
	user := makeTestUser()

	err := db.Users.Create(ctx, user)
	require.NoError(t, err, "Failed to create user")

	// Verify that the user was created successfully
	createdUser, err := db.Users.GetByEmail(ctx, user.Email)
	require.NoError(t, err, "Failed to get user by Email")

	compareUsers(t, user, createdUser)
}

func TestGetAllUsers(t *testing.T) {
	user := makeTestUser()
	user2 := makeTestUser()
	user2.Email = "foo@bar.com"
	user2.Name = "Foo Bar"

	err := db.Users.Create(ctx, user)
	require.NoError(t, err, "Failed to create user")
	err = db.Users.Create(ctx, user2)
	require.NoError(t, err, "Failed to create user 2")

	pagination := &models.Pagination{
		Page:  1,
		Limit: 10,
	}
	pagination, err = db.Users.GetAll(ctx, pagination)
	require.NoError(t, err, "Failed to get all users")

	require.Len(t, pagination.Rows, 2, "Unexpected number of users")
}

func TestGetUserByIDNotFound(t *testing.T) {
	_, err := db.Users.GetByID(ctx, 999)
	require.Error(t, err, "Expected error when user is not found")
}

func TestGetUserByEmailNotFound(t *testing.T) {
	_, err := db.Users.GetByEmail(ctx, "bogus")
	require.Error(t, err, "Expected error when user is not found")
}

func TestUpdateUserNotFound(t *testing.T) {
	user := makeTestUser()
	user.ID = 999
	err := db.Users.Update(ctx, user)
	require.Error(t, err, "Expected error when user is not found")
}

func TestDeleteUserNotFound(t *testing.T) {
	user, err := db.Users.Delete(ctx, uint(999999))
	require.Error(t, err, "Expected error when user is not found")
	assert.Nil(t, user, "Expected nil user when user is not found")
}

func TestGetUserByID(t *testing.T) {
	user := makeTestUser()

	err := db.Users.Create(ctx, user)
	require.NoError(t, err, "Failed to create user")

	createdUser, err := db.Users.GetByEmail(ctx, user.Email)
	require.NoError(t, err, "Failed to get user by Email")

	userByID, err := db.Users.GetByID(ctx, createdUser.ID)
	require.NoError(t, err, "Failed to get user by ID")

	compareUsers(t, user, userByID)
}

func TestUpdateUser(t *testing.T) {
	user := makeTestUser()

	err := db.Users.Create(ctx, user)
	require.NoError(t, err, "Failed to create user")

	createdUser, err := db.Users.GetByEmail(ctx, user.Email)
	require.NoError(t, err, "Failed to get user by Email")

	// Update the user's email
	newEmail := "newemail@example.com"
	createdUser.Email = newEmail

	err = db.Users.Update(ctx, createdUser)
	require.NoError(t, err, "Failed to update user")

	// Get the user by ID
	retrievedUser, err := db.Users.GetByID(ctx, createdUser.ID)
	require.NoError(t, err, "Failed to get user by ID")

	// Verify that the user's email was updated
	compareUsers(t, createdUser, retrievedUser)
}

func TestDeleteUser(t *testing.T) {
	user := makeTestUser()
	err := db.Users.Create(ctx, user)
	require.NoError(t, err, "Failed to create user")

	// Verify that the user was created successfully
	createdUser, err := db.Users.GetByEmail(ctx, user.Email)
	require.NoError(t, err, "Failed to get user by Email")

	// Delete the user
	user, err = db.Users.Delete(ctx, createdUser.ID)
	require.NoError(t, err, "Failed to delete user")
	assert.Equal(t, user.ID, createdUser.ID, "Unexpected user ID")

	// Verify that the user was deleted
	_, err = db.Users.GetByEmail(ctx, user.Email)
	require.Error(t, err, "User was not deleted")
}
