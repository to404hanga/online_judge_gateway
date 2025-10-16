package service

import (
	"context"
	"fmt"

	ojmodel "github.com/to404hanga/online_judge_common/model"
	"github.com/to404hanga/onlinue_judge_gateway/domain"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService interface {
	Login(ctx context.Context, req *domain.LoginRequest) (uint64, error)
	Info(ctx context.Context, userId uint64) (*domain.InfoResponse, error)
}

type AuthServiceImpl struct {
	db *gorm.DB
}

var _ AuthService = (*AuthServiceImpl)(nil)

func NewAuthService(db *gorm.DB) AuthService {
	return &AuthServiceImpl{
		db: db,
	}
}

func (s *AuthServiceImpl) Login(ctx context.Context, req *domain.LoginRequest) (uint64, error) {
	var user ojmodel.User
	err := s.db.Model(&ojmodel.User{}).
		Where("username = ?", req.Username).
		Where("status = ?", ojmodel.UserStatusNormal).
		Select("id", "password"). // 只获取 id 和 password
		First(&user).Error
	if err != nil {
		return 0, fmt.Errorf("get user from db error: %w", err)
	}

	// 密码校验
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		return 0, fmt.Errorf("password not match")
	}

	return user.ID, nil
}

func (s *AuthServiceImpl) Info(ctx context.Context, userId uint64) (*domain.InfoResponse, error) {
	var user ojmodel.User
	err := s.db.Model(&ojmodel.User{}).
		Where("id = ?", userId).
		Select("username", "realname", "role", "status", "created_at", "updated_at").
		First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("get user from db error: %w", err)
	}

	return &domain.InfoResponse{
		Username:  user.Username,
		Realname:  user.Realname,
		Role:      user.Role.String(),
		Status:    user.Status.String(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}
