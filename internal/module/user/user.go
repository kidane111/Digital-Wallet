package auth

import (
	"context"

	"digital-wallet/internal/constant/errors"
	"digital-wallet/internal/constant/model/dto"
	"digital-wallet/internal/module"
	storage "digital-wallet/internal/storage"
	"digital-wallet/platform/logger"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type auth struct {
	log         logger.Logger
	userStorage storage.Auth
	session     storage.Catch
}

func Init(
	log logger.Logger,
	userStorage storage.Auth,
	session storage.Catch,
) module.Auth {
	return &auth{
		log:         log,
		userStorage: userStorage,
		session:     session,
	}
}

func (a *auth) Login(ctx context.Context, req dto.Login) (*dto.LoginResponse, error) {
	if err := req.Validate(); err != nil {
		wrapped := errors.ErrInvalidUserInput.Wrap(err, "invalid login input")
		a.log.Error(ctx, "invalid login input", zap.Error(wrapped))
		return nil, wrapped
	}
	user, err := a.userStorage.FindUserByPhone(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	if user.Status != "active" {
		a.log.Error(ctx, "user is not active",
			zap.String("user_id", user.ID))
		return nil, errors.ErrAcessError.New("User is not active")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(req.Password),
		[]byte(user.Password)); err != nil {
		err := errors.ErrInvalidUserInput.Wrap(err, "invalid password")
		a.log.Error(ctx, "invalid passworld ", zap.Error(err),
			zap.Any("request", req))
		return nil, err
	}
	if err := a.userStorage.UpdateLastLogin(ctx, user.ID); err != nil {
		return nil, err
	}
	token, err := a.session.CreateSession(ctx, user.ID, user.Tier)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		UserID:    user.ID,
		Tier:      user.Tier,
		Token:     token.Token,
		ExpiresAt: token.ExpiresAt,
	}, nil
}

func (a *auth) UserRegister(ctx context.Context, req dto.RegisterRequest) (*dto.User, error) {
	if err := req.Validate(); err != nil {
		wrapped := errors.ErrInvalidUserInput.Wrap(err, "invalid  input")
		a.log.Error(ctx, "invalid input", zap.Error(wrapped))
		return nil, wrapped
	}
	_, err := a.userStorage.FindUserByPhone(ctx, req.Phone)
	if err == nil {
		return nil, err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		wrapped := errors.ErrInternalServerError.Wrap(err, "failed to hash password")
		a.log.Error(ctx, "failed to hash password", zap.Error(wrapped))
		return nil, wrapped
	}
	req.Password = string(hashedPassword)

	return a.userStorage.UserRegister(ctx, req)

}
