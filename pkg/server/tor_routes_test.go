package server_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/humper/tor_exit_nodes/models"
	"github.com/humper/tor_exit_nodes/pkg/database"
	mock_database "github.com/humper/tor_exit_nodes/pkg/database/mock"
	"github.com/humper/tor_exit_nodes/pkg/server"
	"github.com/humper/tor_exit_nodes/testing/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAllTorExitNodesHappy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	users := mock_database.NewMockUsers(ctrl)
	torExitNodes := mock_database.NewMockTorExitNodes(ctrl)

	torExitNodes.EXPECT().GetAll(gomock.Any(), gomock.Eq([]string{}), gomock.Eq(&models.Pagination{
		Page:  1,
		Limit: 10,
	})).Return(&models.Pagination{
		Page:       1,
		Limit:      10,
		TotalRows:  int64(len(fixtures.TestRows)),
		TotalPages: (len(fixtures.TestRows) + 9) / 10,
		Rows:       fixtures.TestRows[:10],
	}, nil)

	db := &database.Database{
		Users:        users,
		TorExitNodes: torExitNodes,
	}

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})

	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/tor", nil)
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var response models.TENPagination
	err = json.NewDecoder(recorder.Body).Decode(&response)
	require.NoError(t, err)

	require.Equal(t, 1, response.Page)
	assert.Equal(t, 10, response.Limit)
	assert.Equal(t, int64(len(fixtures.TestRows)), response.TotalRows)
	assert.Equal(t, (len(fixtures.TestRows)+9)/10, response.TotalPages)
	assert.EqualValues(t, fixtures.TestRows[:10], response.Rows)
}

func TestGetAllTorExitNodesHappyWithUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	users := mock_database.NewMockUsers(ctrl)
	torExitNodes := mock_database.NewMockTorExitNodes(ctrl)

	user := testAccount()
	user.AllowedIPs = []string{
		"101.99.84.87",
		"101.99.92.179",
		"101.99.92.182",
		"101.99.92.194",
		"101.99.92.198",
		"102.130.113.9",
	}

	users.EXPECT().GetByID(gomock.Any(), gomock.Eq(uint(1))).
		Times(1).
		Return(user, nil)

	filteredRows := fixtures.FilterRows(user.AllowedIPs)

	torExitNodes.EXPECT().GetAll(gomock.Any(),
		gomock.InAnyOrder(user.AllowedIPs),
		gomock.Eq(&models.Pagination{
			Page:  1,
			Limit: 10,
		})).Return(&models.Pagination{
		Page:       1,
		Limit:      10,
		TotalRows:  int64(len(filteredRows)),
		TotalPages: (len(filteredRows) + 9) / 10,
		Rows:       filteredRows,
	}, nil)

	db := &database.Database{
		Users:        users,
		TorExitNodes: torExitNodes,
	}

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})

	recorder := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/tor", nil)
	require.NoError(t, err)

	addAuth(req, user)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Code)

	require.NoError(t, err)
	var response models.TENPagination
	err = json.NewDecoder(recorder.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	assert.Equal(t, 10, response.Limit)
	assert.Equal(t, int64(len(filteredRows)), response.TotalRows)
	assert.Equal(t, (len(filteredRows)+9)/10, response.TotalPages)
	assert.EqualValues(t, filteredRows, response.Rows)
}

func TestGetAllTorExitNodesBadPagination(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	users := mock_database.NewMockUsers(ctrl)
	torExitNodes := mock_database.NewMockTorExitNodes(ctrl)

	db := &database.Database{
		Users:        users,
		TorExitNodes: torExitNodes,
	}

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})

	recorder := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/tor?page=bad", nil)
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestGetAllTorNodesDatabaseFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	users := mock_database.NewMockUsers(ctrl)
	torExitNodes := mock_database.NewMockTorExitNodes(ctrl)

	torExitNodes.EXPECT().GetAll(gomock.Any(), gomock.Eq([]string{}), gomock.Eq(&models.Pagination{
		Page:  1,
		Limit: 10,
	})).Return(nil, assert.AnError)

	db := &database.Database{
		Users:        users,
		TorExitNodes: torExitNodes,
	}

	s := server.New(context.Background(), &server.NewServerParams{
		DB: db,
	})

	recorder := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/tor", nil)
	require.NoError(t, err)

	s.GetHandler().ServeHTTP(recorder, req)
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}
