package database

import (
	"context"

	"github.com/humper/tor_exit_nodes/models"
)

type Users interface {
	GetAll(ctx context.Context, pagination *models.Pagination) (*models.Pagination, error)
	GetByID(ctx context.Context, id uint) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uint) (*models.User, error)
}
