package repository

import (
	entity "api-jellyfish/internal/entity"
	database "api-jellyfish/internal/pkg/db"
	"context"

	"github.com/google/uuid"
)

type IAuthRepository interface {
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
}

type IUserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	FindAll(ctx context.Context, activeOnly bool) ([]entity.User, error)
	GetUsersByRole(ctx context.Context, role string) ([]entity.User, error)
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, id uuid.UUID, user *entity.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type AuthRepository struct {
	database *database.Database
}

func NewAuthRepository(database *database.Database) IAuthRepository {
	return &AuthRepository{database: database}
}

func (r *AuthRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	err := r.database.DB.GetContext(ctx, &user, "SELECT id, username, email, password, full_name ,is_active FROM m_user WHERE username = $1 and is_active is true", username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

type UserRepository struct {
	db interface{}
}

func NewUserRepository(db interface{}) IUserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.(interface {
		GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	}).GetContext(ctx, &user, "SELECT * FROM users WHERE id = $1 AND is_active = true", id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindAll(ctx context.Context, activeOnly bool) ([]entity.User, error) {
	var users []entity.User
	query := "SELECT * FROM users"
	if activeOnly {
		query += " WHERE is_active = true"
	}
	err := r.db.(interface {
		SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	}).SelectContext(ctx, &users, query)
	return users, err
}

func (r *UserRepository) GetUsersByRole(ctx context.Context, role string) ([]entity.User, error) {
	var users []entity.User
	query := `
		SELECT u.* FROM users u
		JOIN user_roles ur ON u.id = ur.user_id
		JOIN roles r ON ur.role_id = r.id
		WHERE r.code = $1 AND u.is_active = true
	`
	err := r.db.(interface {
		SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	}).SelectContext(ctx, &users, query, role)
	return users, err
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	return nil
}

func (r *UserRepository) Update(ctx context.Context, id uuid.UUID, user *entity.User) error {
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}
