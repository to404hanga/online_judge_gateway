package service

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	ojmodel "github.com/to404hanga/online_judge_common/model"
	constants "github.com/to404hanga/online_judge_gateway/constant"
	"github.com/to404hanga/online_judge_gateway/domain"
	"github.com/to404hanga/pkg404/cachex/lru"
	loggerv2 "github.com/to404hanga/pkg404/logger/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService interface {
	Login(ctx context.Context, req *domain.LoginRequest) (uint64, error)
	Info(ctx context.Context, userId uint64) (*domain.InfoResponse, error)
}

type AuthServiceImpl struct {
	db    *gorm.DB
	log   loggerv2.Logger
	cache *lru.Cache
}

var _ AuthService = (*AuthServiceImpl)(nil)

func NewAuthService(db *gorm.DB, rds redis.Cmdable, log loggerv2.Logger, cache *lru.Cache) AuthService {
	return &AuthServiceImpl{
		db:    db,
		log:   log,
		cache: cache,
	}
}

func (s *AuthServiceImpl) Login(ctx context.Context, req *domain.LoginRequest) (uint64, error) {
	var user ojmodel.User
	err := s.db.Model(&ojmodel.User{}).
		Where("username = ?", req.Username).
		Where("status = ?", ojmodel.UserStatusNormal).
		Select("id", "username", "realname", "realname", "role", "password").
		First(&user).Error
	if err != nil {
		return 0, fmt.Errorf("get user from db error: %w", err)
	}

	// 密码校验
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		return 0, fmt.Errorf("password not match")
	}

	s.cache.Add(fmt.Sprintf(constants.CacheUserKey, user.ID), constants.CacheUser{
		Username: user.Username,
		Realname: user.Realname,
		Role:     user.Role.Int8(),
	})

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
		Username: user.Username,
		Realname: user.Realname,
		Role:     user.Role.Int8(),
		Status:   user.Status.Int8(),
	}, nil
}
