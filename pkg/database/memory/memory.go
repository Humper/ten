package memory

import (
	"context"

	"github.com/humper/tor_exit_nodes/models"
	"github.com/humper/tor_exit_nodes/pkg/database"
)

func New(ctx context.Context) (*database.Database, error) {

	return &database.Database{
		Users: &users{
			byEmail: make(map[string]*models.User),
			byId:    make(map[uint]*models.User),
		},
		TorExitNodes: &torExitNodes{
			nodes: make(map[string]*models.TorExitNode),
		},
	}, nil
}
