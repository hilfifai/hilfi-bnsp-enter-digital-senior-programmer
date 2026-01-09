package repository

import (
	entity "api-jellyfish/internal/entity"
	database "api-jellyfish/internal/pkg/db"
	"api-jellyfish/internal/pkg/pagination"
	"context"

	"github.com/google/uuid"
)

type IRoleRepository interface {
	GetRoleByUser(uctx context.Context, userID uuid.UUID) []entity.Role
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error)
	Create(ctx context.Context, Role *entity.Role) error
	Update(ctx context.Context, Role *entity.Role) error
	GetAll(ctx context.Context) ([]entity.Role, error)
	GetAllPaginated(ctx context.Context, pagination *pagination.Pagination) (pagination.PaginatedResponse[entity.Role], error)
}
type RoleRepository struct {
	database *database.Database
}

func NewRoleRepository(database *database.Database) IRoleRepository {
	return &RoleRepository{database: database}
}

func (r *RoleRepository) GetRoleByUser(ctx context.Context, userID uuid.UUID) []entity.Role {
	var role_ids []entity.Role
	err := r.database.DB.SelectContext(ctx, &role_ids, "SELECT mr.* FROM m_user_role mur left join m_role mr on mr.id = mur.role_id WHERE mur.user_id = $1", userID)
	if err != nil {
		return role_ids
	}
	return role_ids
}

func (r *RoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
	var Role entity.Role
	err := r.database.DB.GetContext(ctx, &Role, "select * from m_role where id = $1", id)
	if err != nil {
		return nil, database.MapPGError(err)
	}
	return &Role, nil
}

func (r *RoleRepository) Create(ctx context.Context, role *entity.Role) error {
	_, err := r.database.DB.ExecContext(ctx, "insert into m_role (role_group_id, code, name, is_active, description, created_by) values ($1, $2, $3, $4, $5, $6)",
		role.RoleGroupID, role.Code, role.Name, role.Active, role.Description, role.CreatedBy)
	return database.MapPGError(err)
}

func (r *RoleRepository) Update(ctx context.Context, role *entity.Role) error {
	_, err := r.database.DB.ExecContext(ctx, "update m_role set role_group_id = $1, code = $2, name = $3, description = $4, is_active = $5, updated_by = $6, updated_at = now() where id = $7",
		role.RoleGroupID, role.Code, role.Name, role.Description, role.Active, role.UpdatedBy, role.ID)
	return database.MapPGError(err)
}

func (r *RoleRepository) GetAll(ctx context.Context) ([]entity.Role, error) {
	var Roles []entity.Role
	err := r.database.DB.SelectContext(ctx, &Roles, "select * from m_role where is_active is true")
	if err != nil {
		return nil, database.MapPGError(err)
	}
	return Roles, nil
}

func (r *RoleRepository) GetAllPaginated(ctx context.Context, p *pagination.Pagination) (pagination.PaginatedResponse[entity.Role], error) {
	baseQuery := `SELECT id, role_group_id, code, name, description, is_active, created_at, updated_at FROM m_role`
	countQuery := "SELECT COUNT(*) FROM m_role"

	return pagination.FetchPaginated[entity.Role](
		ctx,
		r.database.DB,
		baseQuery,
		countQuery,
		p,
	)
}
