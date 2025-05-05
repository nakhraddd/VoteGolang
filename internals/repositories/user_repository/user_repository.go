package user_repository

import (
	"VoteGolang/internals/data/user_data"
	"gorm.io/gorm"
)

// UserRepository handles database operations related to users.
type UserRepository interface {
	Create(user *user_data.User) error
	GetByID(id string) (*user_data.User, error)
	GetByUsername(username string) (*user_data.User, error)
	Update(user *user_data.User) error
	Delete(id string) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *user_data.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(id string) (*user_data.User, error) {
	var user user_data.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(username string) (*user_data.User, error) {
	var user user_data.User
	err := r.db.First(&user, "username = ?", username).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *user_data.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id string) error {
	return r.db.Delete(&user_data.User{}, "id = ?", id).Error
}
