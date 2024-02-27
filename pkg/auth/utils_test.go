package auth_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/humper/tor_exit_nodes/models"
	"github.com/humper/tor_exit_nodes/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestComparePassword(t *testing.T) {
	password := "password"
	hash, _ := auth.HashPassword(password)

	assert.True(t, auth.ComparePassword(password, hash))
	assert.False(t, auth.ComparePassword("wrong_password", hash))
}

func TestHashPassword(t *testing.T) {
	password := "password"
	hash, err := auth.HashPassword(password)

	require.NoError(t, err, "Error hashing password")
	assert.NotEqual(t, password, hash, "HashPassword did not hash the password")
}

func TestCreateJWT(t *testing.T) {
	user := &models.User{
		Model: gorm.Model{
			ID: 1,
		},
		Email: "test@test.com",
	}

	tokenString, _, err := auth.CreateJWT(user)
	require.NoError(t, err, "Error creating JWT")
	assert.NotEmpty(t, tokenString, "CreateJWT did not generate a token")
}

func TestParseJWT(t *testing.T) {
	user := &models.User{
		Model: gorm.Model{
			ID: 1,
		},
		Email: "test@test.com",
	}

	tokenString, _, _ := auth.CreateJWT(user)
	claims, err := auth.ParseJWT(tokenString)

	require.NoError(t, err, "Error parsing JWT")
	assert.Equal(t, user.Email, claims.StandardClaims.Subject, "ParseJWT did not parse the correct username")
	assert.Equal(t, fmt.Sprintf("%d", user.ID), claims.StandardClaims.Id, "ParseJWT did not parse the correct user ID")
}

func TestParseJWTBadToken(t *testing.T) {
	_, err := auth.ParseJWT("bad_token")
	require.Error(t, err, "ParseJWT did not return an error for a bad token")
}

func TestNewContext(t *testing.T) {
	user := &models.User{
		Model: gorm.Model{
			ID: 1,
		},
		Email: "test@test.com",
	}

	ctx := context.Background()
	ctx = auth.NewContext(ctx, user)

	result := auth.GetUser(ctx)
	require.NotNil(t, result, "NewContext did not set the user in the context")
	assert.Equal(t, user.ID, result.ID, "NewContext did not set the correct user in the context")
}
