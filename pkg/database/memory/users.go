package memory

import (
	"context"
	"math"
	"slices"
	"sync"

	"github.com/humper/tor_exit_nodes/models"
	"gorm.io/gorm"
)

func userCopy(user *models.User) *models.User {
	return &models.User{
		Model:      gorm.Model{ID: user.ID},
		Role:       user.Role,
		Name:       user.Name,
		Email:      user.Email,
		Password:   user.Password,
		AllowedIPs: slices.Clone(user.AllowedIPs),
	}
}

type users struct {
	byEmail          map[string]*models.User
	byId             map[uint]*models.User
	mutex            sync.Mutex
	allowlistCounter uint
}

func (u *users) GetAll(ctx context.Context, pagination *models.Pagination) (*models.Pagination, error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	allUsers := []*models.User{}
	for _, user := range u.byEmail {
		allUsers = append(allUsers, userCopy(user))
	}

	users := make([]*models.User, 0, pagination.GetLimit())

	totalUsers := len(allUsers)
	pagination.TotalRows = int64(totalUsers)
	totalPages := int(math.Ceil(float64(totalUsers) / float64(pagination.GetLimit())))
	pagination.TotalPages = totalPages

	start := pagination.GetOffset()
	end := start + pagination.GetLimit()
	if end > totalUsers {
		end = totalUsers
	}

	for i := start; i < end; i++ {
		users = append(users, allUsers[i])
	}
	pagination.Rows = users
	return pagination, nil

}

func (u *users) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	user, ok := u.byEmail[email]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}

	return userCopy(user), nil
}

func (u *users) GetByID(ctx context.Context, id uint) (*models.User, error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	user, ok := u.byId[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}

	return userCopy(user), nil
}

func (u *users) Create(ctx context.Context, user *models.User) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	userToCreate := userCopy(user)
	userToCreate.ID = uint(len(u.byEmail) + 1)

	u.byEmail[user.Email] = userToCreate
	u.byId[userToCreate.ID] = userToCreate

	return nil
}

func (u *users) Update(ctx context.Context, user *models.User) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	_, ok := u.byId[user.ID]
	if !ok {
		return gorm.ErrRecordNotFound
	}

	userToUpdate := userCopy(user)
	u.byEmail[user.Email] = userToUpdate
	u.byId[user.ID] = userToUpdate
	return nil
}

func (u *users) Delete(ctx context.Context, id uint) (*models.User, error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	user, ok := u.byId[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	delete(u.byEmail, user.Email)
	delete(u.byId, user.ID)
	return user, nil
}
