package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/humper/tor_exit_nodes/models"
	"github.com/humper/tor_exit_nodes/pkg/auth"
	"github.com/humper/tor_exit_nodes/pkg/database"
	mock_database "github.com/humper/tor_exit_nodes/pkg/database/mock"
	"github.com/humper/tor_exit_nodes/pkg/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var passwords = map[string]string{
	"test@test.com":   "password",
	"admin@admin.com": "admin_password",
}

func testAccount() *models.User {
	password, _ := auth.HashPassword(passwords["test@test.com"])

	return &models.User{
		Model: gorm.Model{
			ID: 1,
		},
		Name:     "Testy McTestface",
		Email:    "test@test.com",
		Password: password,
		Role:     "user",
	}
}

func adminAccount() *models.User {
	password, _ := auth.HashPassword(passwords["admin@admin.com"])

	return &models.User{
		Model: gorm.Model{
			ID: 2,
		},
		Name:     "Admin McAdminface",
		Email:    "admin@admin.com",
		Password: password,
		Role:     "admin",
	}
}

func addAuth(req *http.Request, user *models.User) {
	jwt, _, _ := auth.CreateJWT(user)
	req.Header.Set("Cookie", "token="+jwt)
}

func TestHandleLoginHappy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	user := testAccount()

	users := mock_database.NewMockUsers(ctrl)
	users.EXPECT().GetByEmail(gomock.Any(), gomock.Eq(user.Email)).
		Times(1).
		Return(user, nil)

	db := &database.Database{
		Users: users,
	}

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})
	recorder := httptest.NewRecorder()

	body := map[string]string{
		"email":    user.Email,
		"password": passwords[user.Email],
	}
	jsonBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBytes))
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	require.Len(t, recorder.Result().Cookies(), 1)
	jwt := recorder.Result().Cookies()[0]
	assert.Equal(t, "token", jwt.Name)
	assert.NotEmpty(t, jwt.Value)
	assert.Equal(t, "/", jwt.Path)
	assert.Equal(t, "localhost", jwt.Domain)
	assert.True(t, jwt.HttpOnly)

	val := jwt.Value
	claims, err := auth.ParseJWT(val)
	require.NoError(t, err)
	assert.Equal(t, user.Email, claims.Subject)
	assert.Equal(t, "ten", claims.Issuer)
	assert.Equal(t, "1", claims.Id)
	assert.Equal(t, user.Role, claims.Role)
}

func TestHandleLoginBadBody(t *testing.T) {
	s := server.New(context.Background(), &server.NewServerParams{})
	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer([]byte("bad json")))
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandleLoginUnknownUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	users := mock_database.NewMockUsers(ctrl)
	users.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).
		Times(1).
		Return(nil, gorm.ErrRecordNotFound)

	db := &database.Database{
		Users: users,
	}

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})
	recorder := httptest.NewRecorder()

	body := map[string]string{
		"email":    "foo",
		"password": "bar",
	}
	jsonBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBytes))
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestHandleLoginWrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	user := testAccount()

	users := mock_database.NewMockUsers(ctrl)
	users.EXPECT().GetByEmail(gomock.Any(), gomock.Eq(user.Email)).
		Times(1).
		Return(user, nil)

	db := &database.Database{
		Users: users,
	}

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})
	recorder := httptest.NewRecorder()

	body := map[string]string{
		"email":    user.Email,
		"password": "wrong",
	}
	jsonBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBytes))
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestHandleRegisterHappy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	users := mock_database.NewMockUsers(ctrl)
	users.EXPECT().GetByEmail(gomock.Any(), gomock.Eq("test@test.com")).
		Times(1).
		Return(nil, gorm.ErrRecordNotFound)

	users.EXPECT().Create(gomock.Any(), gomock.Any()).
		Times(1).
		Return(nil)

	db := &database.Database{
		Users: users,
	}

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})
	recorder := httptest.NewRecorder()

	u := models.User{
		Email:    "test@test.com",
		Password: passwords["test@test.com"],
		Role:     "user",
	}
	jsonBytes, err := json.Marshal(u)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonBytes))
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Code)

	var user models.User
	err = json.NewDecoder(recorder.Body).Decode(&user)
	require.NoError(t, err)
	assert.Equal(t, u.Email, user.Email)
	assert.Equal(t, u.Role, user.Role)

	assert.True(t, auth.ComparePassword(passwords["test@test.com"], user.Password))
}

func TestHandleRegisterBadBody(t *testing.T) {
	s := server.New(context.Background(), &server.NewServerParams{})
	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("POST", "/users", bytes.NewBuffer([]byte("bad json")))
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandleRegisterExistingUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	user := testAccount()

	users := mock_database.NewMockUsers(ctrl)
	users.EXPECT().GetByEmail(gomock.Any(), gomock.Eq(user.Email)).
		Times(1).
		Return(user, nil)

	db := &database.Database{
		Users: users,
	}

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})
	recorder := httptest.NewRecorder()

	u := models.User{
		Email:    user.Email,
		Password: passwords[user.Email],
		Role:     "user",
	}
	jsonBytes, err := json.Marshal(u)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonBytes))
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusConflict, recorder.Code)
}

func TestHandleRegisterPasswordTooLong(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	users := mock_database.NewMockUsers(ctrl)
	users.EXPECT().GetByEmail(gomock.Any(), gomock.Eq("test@test.com")).
		Times(1).
		Return(nil, gorm.ErrRecordNotFound)

	db := &database.Database{
		Users: users,
	}

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})
	recorder := httptest.NewRecorder()

	u := models.User{
		Email:    "test@test.com",
		Password: "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
		Role:     "user",
	}
	jsonBytes, err := json.Marshal(u)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonBytes))
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandleRegisterUserCreateFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	users := mock_database.NewMockUsers(ctrl)
	users.EXPECT().GetByEmail(gomock.Any(), gomock.Eq("test@test.com")).
		Times(1).
		Return(nil, gorm.ErrRecordNotFound)

	users.EXPECT().Create(gomock.Any(), gomock.Any()).
		Times(1).
		Return(gorm.ErrInvalidDB)

	db := &database.Database{
		Users: users,
	}

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})
	recorder := httptest.NewRecorder()

	u := models.User{
		Email:    "test@test.com",
		Password: passwords["test@test.com"],
		Role:     "user",
	}
	jsonBytes, err := json.Marshal(u)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonBytes))
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestHandleLogout(t *testing.T) {
	s := server.New(context.Background(), &server.NewServerParams{})
	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("POST", "/logout", nil)
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusNoContent, recorder.Code)
	require.Len(t, recorder.Result().Cookies(), 1)
	jwt := recorder.Result().Cookies()[0]
	assert.Equal(t, "token", jwt.Name)
	assert.Empty(t, jwt.Value)
	assert.Equal(t, 0, jwt.MaxAge)
	assert.True(t, jwt.HttpOnly)
}

func TestHandleGetUsersHappy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	user := testAccount()

	users := mock_database.NewMockUsers(ctrl)
	users.EXPECT().GetAll(gomock.Any(), gomock.Eq(&models.Pagination{
		Page:  1,
		Limit: 10,
	})).
		Times(1).
		Return(&models.Pagination{
			Page:       1,
			Limit:      10,
			TotalRows:  1,
			TotalPages: 1,
			Rows:       []models.User{*user},
		}, nil)

	db := &database.Database{
		Users: users,
	}

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})
	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/users", nil)
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestHandleGetUsersBadPage(t *testing.T) {
	s := server.New(context.Background(), &server.NewServerParams{})
	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/users?page=foo", nil)
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandleGetUsersBadLimit(t *testing.T) {
	s := server.New(context.Background(), &server.NewServerParams{})
	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/users?limit=foo", nil)
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandleGetUsersBadFilter(t *testing.T) {
	s := server.New(context.Background(), &server.NewServerParams{})
	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/users?filter=foo", nil)
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestGetUsersDatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	users := mock_database.NewMockUsers(ctrl)
	users.EXPECT().GetAll(gomock.Any(), gomock.Any()).
		Times(1).
		Return(nil, gorm.ErrInvalidDB)

	db := &database.Database{
		Users: users,
	}

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})
	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/users", nil)
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestHandleGetUserHappy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	user := testAccount()

	users := mock_database.NewMockUsers(ctrl)
	users.EXPECT().GetByID(gomock.Any(), gomock.Eq(uint(1))).
		Times(1).
		Return(user, nil)

	db := &database.Database{
		Users: users,
	}

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})
	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/users/1", nil)
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var u models.User
	err = json.NewDecoder(recorder.Body).Decode(&u)
	require.NoError(t, err)
	assert.Equal(t, user.Email, u.Email)
	assert.Equal(t, user.Role, u.Role)
	assert.Equal(t, user.ID, u.ID)
	assert.True(t, auth.ComparePassword(passwords["test@test.com"], u.Password))
}

func TestHandleGetUserBadID(t *testing.T) {
	s := server.New(context.Background(), &server.NewServerParams{})
	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/users/foo", nil)
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandleGetUserDatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	users := mock_database.NewMockUsers(ctrl)
	users.EXPECT().GetByID(gomock.Any(), gomock.Eq(uint(1))).
		Times(1).
		Return(nil, gorm.ErrInvalidDB)

	db := &database.Database{
		Users: users,
	}

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})
	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/users/1", nil)
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestHandleUpdateUserHappy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	user := testAccount()

	users := mock_database.NewMockUsers(ctrl)
	users.EXPECT().GetByID(gomock.Any(), gomock.Eq(uint(1))).
		Times(2).
		Return(user, nil)

	users.EXPECT().Update(gomock.Any(), gomock.Any()).
		Times(1).
		Return(nil)

	db := &database.Database{
		Users: users,
	}

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})
	recorder := httptest.NewRecorder()

	user.ID = 1
	user.Name = "New Name"
	jsonBytes, err := json.Marshal(user)
	require.NoError(t, err)

	req, err := http.NewRequest("PUT", "/users/1", bytes.NewBuffer(jsonBytes))
	require.NoError(t, err)

	addAuth(req, user)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Code)

	var u models.User
	err = json.NewDecoder(recorder.Body).Decode(&u)
	require.NoError(t, err)
	assert.Equal(t, u.Email, user.Email)
	assert.Equal(t, u.Role, user.Role)

	assert.True(t, auth.ComparePassword(passwords["test@test.com"], u.Password))
}

func TestHandleUpdateUserLoggedOut(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	user := testAccount()

	users := mock_database.NewMockUsers(ctrl)
	users.EXPECT().GetByID(gomock.Any(), gomock.Eq(uint(1))).
		Times(1).
		Return(user, nil)

	db := &database.Database{
		Users: users,
	}

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})
	recorder := httptest.NewRecorder()

	user.ID = 1
	user.Name = "New Name"
	jsonBytes, err := json.Marshal(user)
	require.NoError(t, err)

	req, err := http.NewRequest("PUT", "/users/1", bytes.NewBuffer(jsonBytes))
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestHandleUpdateUserBadUserID(t *testing.T) {
	s := server.New(context.Background(), &server.NewServerParams{})
	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("PUT", "/users/foo", nil)
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandleUpdateUserUnknownUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	users := mock_database.NewMockUsers(ctrl)
	users.EXPECT().GetByID(gomock.Any(), gomock.Eq(uint(1))).
		Times(1).
		Return(nil, gorm.ErrRecordNotFound)

	db := &database.Database{
		Users: users,
	}

	user := testAccount()
	jsonBytes, err := json.Marshal(user)
	require.NoError(t, err)

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})
	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("PUT", "/users/1", bytes.NewBuffer(jsonBytes))
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestHandleUpdateUserDatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	user := testAccount()

	users := mock_database.NewMockUsers(ctrl)
	users.EXPECT().GetByID(gomock.Any(), gomock.Eq(uint(1))).
		Times(2).
		Return(user, nil)

	users.EXPECT().Update(gomock.Any(), gomock.Any()).
		Times(1).
		Return(gorm.ErrInvalidDB)

	db := &database.Database{
		Users: users,
	}

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})
	recorder := httptest.NewRecorder()

	user.ID = 1
	user.Name = "New Name"
	jsonBytes, err := json.Marshal(user)
	require.NoError(t, err)

	req, err := http.NewRequest("PUT", "/users/1", bytes.NewBuffer(jsonBytes))
	require.NoError(t, err)

	addAuth(req, user)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestHandleUpdateUserBadBody(t *testing.T) {
	s := server.New(context.Background(), &server.NewServerParams{})
	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("PUT", "/users/1", bytes.NewBuffer([]byte("bad json")))
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandleUpdateUserAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	adminUser := adminAccount()
	user := testAccount()

	users := mock_database.NewMockUsers(ctrl)
	users.EXPECT().GetByID(gomock.Any(), gomock.Eq(uint(1))).
		Times(1).
		Return(user, nil)
	users.EXPECT().GetByID(gomock.Any(), gomock.Eq(uint(2))).
		Times(1).
		Return(adminUser, nil)

	users.EXPECT().Update(gomock.Any(), gomock.Any()).
		Times(1).
		Return(nil)

	db := &database.Database{
		Users: users,
	}

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})
	recorder := httptest.NewRecorder()

	user.ID = 1
	user.Name = "New Name"
	jsonBytes, err := json.Marshal(user)
	require.NoError(t, err)

	req, err := http.NewRequest("PUT", "/users/1", bytes.NewBuffer(jsonBytes))
	require.NoError(t, err)

	addAuth(req, adminUser)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Code)

	var u models.User
	err = json.NewDecoder(recorder.Body).Decode(&u)
	require.NoError(t, err)
	assert.Equal(t, u.Email, user.Email)
	assert.Equal(t, u.Role, user.Role)
}

func TestHandleUpdateUserRegularUserUnauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	adminUser := adminAccount()
	user := testAccount()

	users := mock_database.NewMockUsers(ctrl)
	users.EXPECT().GetByID(gomock.Any(), gomock.Eq(uint(1))).
		Times(1).
		Return(user, nil)
	users.EXPECT().GetByID(gomock.Any(), gomock.Eq(uint(2))).
		Times(1).
		Return(adminUser, nil)

	db := &database.Database{
		Users: users,
	}

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})
	recorder := httptest.NewRecorder()

	adminUser.ID = 2
	adminUser.Name = "New Name"
	jsonBytes, err := json.Marshal(adminUser)
	require.NoError(t, err)

	req, err := http.NewRequest("PUT", "/users/2", bytes.NewBuffer(jsonBytes))
	require.NoError(t, err)

	addAuth(req, user)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestHandleDeleteUserHappy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	adminUser := adminAccount()
	testUser := testAccount()

	users := mock_database.NewMockUsers(ctrl)
	users.EXPECT().Delete(gomock.Any(), gomock.Eq(uint(1))).
		Times(1).
		Return(testUser, nil)
	users.EXPECT().GetByID(gomock.Any(), gomock.Eq(uint(2))).
		Times(1).
		Return(adminUser, nil)

	db := &database.Database{
		Users: users,
	}

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})
	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("DELETE", "/users/1", nil)
	require.NoError(t, err)

	addAuth(req, adminUser)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestHandleDeleteUserLoggedOut(t *testing.T) {
	s := server.New(context.Background(), &server.NewServerParams{})
	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("DELETE", "/users/1", nil)
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestHandleDeleteUserNonAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testUser := testAccount()

	users := mock_database.NewMockUsers(ctrl)
	users.EXPECT().GetByID(gomock.Any(), gomock.Eq(uint(1))).
		Times(1).
		Return(testUser, nil)

	db := &database.Database{
		Users: users,
	}

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})
	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("DELETE", "/users/1", nil)
	require.NoError(t, err)

	addAuth(req, testUser)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestHandleDeleteUserBadID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	adminUser := adminAccount()

	users := mock_database.NewMockUsers(ctrl)
	users.EXPECT().GetByID(gomock.Any(), gomock.Eq(uint(2))).
		Times(1).
		Return(adminUser, nil)

	db := &database.Database{
		Users: users,
	}

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})
	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("DELETE", "/users/foo", nil)
	require.NoError(t, err)

	addAuth(req, adminUser)
	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandleDeleteUserDatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	adminUser := adminAccount()

	users := mock_database.NewMockUsers(ctrl)
	users.EXPECT().Delete(gomock.Any(), gomock.Eq(uint(1))).
		Times(1).
		Return(nil, gorm.ErrInvalidDB)
	users.EXPECT().GetByID(gomock.Any(), gomock.Eq(uint(2))).
		Times(1).
		Return(adminUser, nil)

	db := &database.Database{
		Users: users,
	}

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})
	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("DELETE", "/users/1", nil)
	require.NoError(t, err)

	addAuth(req, adminUser)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}
