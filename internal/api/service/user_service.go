package service

import (
	"context"
	"go-fitness/external/logger/sl"
	"go-fitness/internal/api/types"
	"log/slog"
)

type UserService struct {
	log      *slog.Logger
	userRepo UserRepository
}

type UserRepository interface {
	GetUserByUUID(ctx context.Context, uuid string) (types.User, error)
	GetRoleByUserID(ctx context.Context, userID int64) (types.Role, error)
}

func NewUserService(
	log *slog.Logger,
	userRepo UserRepository,
) *UserService {
	return &UserService{
		log:      log,
		userRepo: userRepo,
	}
}

func (s *UserService) GetUserByUUID(ctx context.Context, uuid string) (types.User, error) {
	const op = "UserService.GetUserByUUID"

	log := s.log.With(
		sl.String("op", op),
		sl.String("uuid", uuid),
	)

	user, err := s.userRepo.GetUserByUUID(ctx, uuid)
	if err != nil {
		log.Error("failed to get user by uuid", sl.Err(err))
		return user, err
	}

	return user, nil
}

func (s *UserService) GetRoleByUserID(ctx context.Context, userID int64) (types.Role, error) {
	const op = "UserService.GetRoleByUserID"

	log := s.log.With(
		sl.String("op", op),
		sl.Int64("user_id", userID),
	)

	role, err := s.userRepo.GetRoleByUserID(ctx, userID)
	if err != nil {
		log.Error("failed to get role by user id", sl.Err(err))
		return role, err
	}

	return role, nil
}
