package service

import (
	"context"
	"errors"
	"log"

	dto "api-jellyfish/internal/dto/auth"
	"api-jellyfish/internal/entity"
	"api-jellyfish/internal/pkg/helper"
	jwt "api-jellyfish/internal/pkg/jwt"
	repository "api-jellyfish/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type IAuthService interface {
	Login(ctx context.Context, req dto.LoginRequest) (*dto.TokenResponse, error)
	RefreshToken(ctx context.Context, token string) (*dto.TokenResponse, error)
}

type AuthService struct {
	repo           repository.IAuthRepository
	roleRepo       repository.IRoleRepository
	permissionRepo repository.IPermissionRepository
	auth           jwt.IJWTAuth
}

func NewAuthService(
	repo repository.IAuthRepository,
	roleRepo repository.IRoleRepository,
	permissionRepo repository.IPermissionRepository,
	auth jwt.IJWTAuth) IAuthService {
	return &AuthService{repo: repo, roleRepo: roleRepo, permissionRepo: permissionRepo, auth: auth}
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.TokenResponse, error) {
	user, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}
	log.Println(user)
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid username or password")
	}
	roles := s.roleRepo.GetRoleByUser(ctx, user.ID)
	roleIds := helper.Map(roles, func(x entity.Role) uuid.UUID {
		return x.ID
	})
	roleName := helper.Map(roles, func(x entity.Role) string {
		return x.Name
	})
	permissions := s.permissionRepo.GetPermissionByRole(ctx, roleIds)
	permissionCode := helper.Map(permissions, func(x entity.Permission) string {
		return x.Code
	})

	var data = jwt.AuthClaims{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Fullname:    user.Fullname,
		Roles:       roleIds,
		RoleString:  roleName,
		Permissions: permissionCode,
	}
	access, refreshToken, _ := s.auth.GenerateToken(data)

	return &dto.TokenResponse{Data: data, AccessToken: access, RefreshToken: refreshToken}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, token string) (*dto.TokenResponse, error) {
	claims, err := s.auth.ValidateToken(token)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	access, refreshToken, _ := s.auth.GenerateToken(*claims)

	return &dto.TokenResponse{Data: claims, AccessToken: access, RefreshToken: refreshToken}, nil
}
