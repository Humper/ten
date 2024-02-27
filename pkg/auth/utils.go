package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/humper/tor_exit_nodes/models"
	"golang.org/x/crypto/bcrypt"
)

// LOL don't put secrets in code
var jwtKey = []byte("super_secret_key")

func ComparePassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CreateJWT(user *models.User) (string, time.Time, error) {

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &models.Claims{
		Role: user.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Issuer:    "ten",
			Id:        fmt.Sprintf("%v", user.ID),
			Subject:   user.Email,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(jwtKey)
	return ss, expirationTime, err
}

func ParseJWT(tokenString string) (claims *models.Claims, err error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*models.Claims)

	if !ok {
		return nil, err
	}

	return claims, nil
}

func NewContext(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, "user", user)
}

func GetUser(ctx context.Context) *models.User {
	user, ok := ctx.Value("user").(*models.User)
	if !ok {
		return nil
	}
	return user
}
