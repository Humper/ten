package psql

import (
	"context"

	"github.com/humper/tor_exit_nodes/models"
	"gorm.io/gorm"
)

type torExitNodes struct {
	db *gorm.DB
}

func (t *torExitNodes) GetAll(ctx context.Context, country_codes []string, excludedIPs []string, pagination *models.Pagination) (*models.Pagination, error) {
	var exitNodes []*models.TorExitNode

	db := t.db

	if len(country_codes) > 0 {
		db = db.Where("country_code IN ?", country_codes)
	}

	if len(excludedIPs) > 0 {
		db = db.Where("ip NOT IN ?", excludedIPs)
	}

	db.Scopes(paginate(exitNodes, pagination, db)).Find(&exitNodes)
	pagination.Rows = exitNodes

	return pagination, nil
}

func (t *torExitNodes) GetUniqueCountryCodes(ctx context.Context, pagination *models.Pagination) (*models.Pagination, error) {
	var exitNodes []*models.TorExitNode

	db := t.db

	db = db.Select("DISTINCT country_code")
	db.Scopes(paginate(exitNodes, pagination, db)).Find(&exitNodes)
	pagination.Rows = exitNodes

	return pagination, nil
}

func (t *torExitNodes) DeleteAndAdd(ctx context.Context, nodes_to_delete []models.TorExitNode, nodes_to_add []*models.TorExitNode) error {
	return t.db.Transaction(func(tx *gorm.DB) error {
		if len(nodes_to_delete) > 0 {
			if err := tx.Unscoped().Delete(nodes_to_delete).Error; err != nil {
				return err
			}
		}

		if len(nodes_to_add) > 0 {
			if err := tx.CreateInBatches(nodes_to_add, 100).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (t *torExitNodes) GetMissingCountries(ctx context.Context, batchSize int) ([]*models.TorExitNode, error) {
	var nodes []*models.TorExitNode
	if err := t.db.Where("country_name = ?", "").Limit(batchSize).Find(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}

func (t *torExitNodes) Update(ctx context.Context, nodes []*models.TorExitNode) error {
	return t.db.Save(nodes).Error
}
