package psql

import (
	"context"

	"github.com/humper/tor_exit_nodes/models"
	"gorm.io/gorm"
)

type users struct {
	db *gorm.DB
}

func (u *users) GetAll(ctx context.Context, pagination *models.Pagination) (*models.Pagination, error) {
	var users []*models.User
	if err := u.db.Scopes(paginate(users, pagination, u.db)).Find(&users).Error; err != nil {
		return nil, err
	}

	// hack to bootstrap the database

	if pagination.TotalRows == 0 {
		err := u.Create(ctx, &models.User{
			Role:       "admin",
			Name:       "Admin",
			Email:      "admin@admin.com",
			Password:   "password",
			AllowedIPs: []string{},
		})
		if err != nil {
			return nil, err
		}
	}

	pagination.Rows = users
	return pagination, nil
}

func (u *users) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := u.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *users) GetByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := u.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *users) Create(ctx context.Context, user *models.User) error {
	if err := u.db.Create(user).Error; err != nil {
		return err
	}
	return nil
}

func (u *users) Update(ctx context.Context, user *models.User) error {
	if err := u.db.Save(user).Error; err != nil {
		return err
	}
	return nil
}

func (u *users) Delete(ctx context.Context, id uint) (*models.User, error) {
	user, err := u.GetByID(ctx, id)
	if err != nil {
		return user, err
	}
	if err := u.db.Where("id = ?", id).Delete(&models.User{}).Error; err != nil {
		return nil, err
	}
	return user, nil
}
