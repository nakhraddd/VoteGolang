package repositories

import (
	"VoteGolang/internals/domain"
	"context"
	"time"

	"gorm.io/gorm"
)

type userGormRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userGormRepository{db: db}
}

func (r *userGormRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *userGormRepository) GetByID(id uint) (*domain.User, error) {
	var user domain.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userGormRepository) GetByUsername(username string) (*domain.User, error) {
	var user domain.User
	err := r.db.First(&user, "username = ?", username).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userGormRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.First(&user, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userGormRepository) Update(user *domain.User) error {
	return r.db.Save(user).Error
}

func (r *userGormRepository) Delete(id uint) error {
	return r.db.Delete(&domain.User{}, "id = ?", id).Error
}

func (r *userGormRepository) MarkEmailVerified(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", userID).
		Update("email_verified", true).
		Error
}

func (r *userGormRepository) DeleteUnverifiedUser(cutoff time.Time) (int64, error) {
	result := r.db.
		Unscoped().
		Where("email_verified = ? AND created_at < ?", false, cutoff).
		Delete(&domain.User{})
	return result.RowsAffected, result.Error
}
