package psql

import (
	"math"

	"github.com/humper/tor_exit_nodes/models"
	"gorm.io/gorm"
)

func paginate(value interface{}, pagination *models.Pagination, db *gorm.DB) func(db *gorm.DB) *gorm.DB {
	var totalRows int64

	if pagination.Filter != nil {
		for key, value := range pagination.Filter {
			db = db.Where(key+" IN ?", value)
		}
	}

	db.Model(value).Count(&totalRows)

	pagination.TotalRows = totalRows
	totalPages := int(math.Ceil(float64(totalRows) / float64(pagination.GetLimit())))
	pagination.TotalPages = totalPages

	return func(db *gorm.DB) *gorm.DB {
		if pagination.Filter != nil {
			for key, value := range pagination.Filter {
				db = db.Where(key+" IN ?", value)
			}
		}
		return db.Offset(pagination.GetOffset()).Limit(pagination.GetLimit()).Order(pagination.GetSort())
	}
}
