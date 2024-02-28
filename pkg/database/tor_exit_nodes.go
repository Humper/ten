package database

import (
	"context"

	"github.com/humper/tor_exit_nodes/models"
)

type TorExitNodes interface {
	GetAll(ctx context.Context, excludedIPs []string, pagination *models.Pagination) (*models.Pagination, error)
	DeleteAndAdd(ctx context.Context, nodes_to_delete []models.TorExitNode, nodes_to_add []*models.TorExitNode) error
	Update(ctx context.Context, nodes []*models.TorExitNode) error
	GetMissingCountries(ctx context.Context, batchSize int) ([]*models.TorExitNode, error)
}
