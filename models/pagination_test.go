package models_test

import (
	"testing"

	"github.com/humper/tor_exit_nodes/models"
	"github.com/stretchr/testify/assert"
)

func TestGetOffset(t *testing.T) {
	p := &models.Pagination{
		Limit: 10,
		Page:  3,
	}
	assert.Equal(t, 20, p.GetOffset())
}

func TestGetLimit(t *testing.T) {
	p := &models.Pagination{
		Limit: 50,
	}
	assert.Equal(t, 50, p.GetLimit())

	pDefaults := &models.Pagination{}
	assert.Equal(t, 10, pDefaults.GetLimit())
}

func TestGetPage(t *testing.T) {
	p := &models.Pagination{
		Page: 5,
	}
	assert.Equal(t, 5, p.GetPage())

	pDefaults := &models.Pagination{}
	assert.Equal(t, 1, pDefaults.GetPage())
}

func TestGetSort(t *testing.T) {
	p := &models.Pagination{
		Sort: "Name asc",
	}
	assert.Equal(t, "Name asc", p.GetSort())

	pDefaults := &models.Pagination{}
	assert.Equal(t, "Id desc", pDefaults.GetSort())
}
