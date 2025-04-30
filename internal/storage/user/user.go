package auth

import (
	"context"
	"database/sql"

	"digital-wallet/internal/constant/errors"
	"digital-wallet/internal/constant/errors/sqlcerr"
	"digital-wallet/internal/constant/model/db"
	"digital-wallet/internal/constant/model/dto"
	"digital-wallet/internal/constant/model/persistencedb"
	st "digital-wallet/internal/storage"
	"digital-wallet/platform/logger"

	"go.uber.org/zap"
)

type auth struct {
	log         logger.Logger
	userStorage persistencedb.PersistenceDB
}

func Init(
	log logger.Logger,
	userStorage persistencedb.PersistenceDB,
) st.Auth {
	return &auth{
		log:         log,
		userStorage: userStorage,
	}
}

// get user by email
func (a *auth) FindUserByPhone(ctx context.Context, email string) (*dto.User, error) {
	user, err := a.userStorage.FindUserByEmail(ctx, email)
	if err != nil {
		if sqlcerr.Is(err, sqlcerr.ErrNoRows) {
			a.log.Error(ctx, "user not found", zap.String("email", email))
			return nil, errors.ErrResourceNotFound.Wrap(err, "user not found")
		}
		err = errors.ErrUnableToGet.Wrap(err, "unable to check for program name")
		a.log.Error(ctx, "unable to check for program name", zap.Error(err))
		return nil, err
	}
	return &dto.User{
		ID: user.ID,
		// Username:  user.Username,
		Email: user.Email,
		// FirstName: user.FirstName,
		// LastName:  user.LastName,
		// Phone:     user.Phone,
		Password:  user.Password,
		Status:    user.Status,
		Tier:      user.Tier,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// UpdateLastLogin updates the last login time of the user
func (a *auth) UpdateLastLogin(ctx context.Context, userID string) error {
	err := a.userStorage.UpdateLastLogin(ctx, userID)
	if err != nil {
		a.log.Error(ctx, "failed to update last login", zap.Error(err))
		return errors.ErrInternalServerError.Wrap(err, "failed to update last login")
	}
	return nil
}

func (a *auth) UserRegister(ctx context.Context, param dto.RegisterRequest) (*dto.User, error) {
	data, err := a.userStorage.UserRegister(ctx, db.UserRegisterParams{
		Name:     param.Name,
		Email:    param.Email,
		Phone:    sql.NullString{String: param.Phone, Valid: param.Phone != ""},
		Password: param.Password,
		Tier:     param.Tier,
	})
	if err != nil {
		err = errors.ErrUnableToCreate.Wrap(err, "unable to add user")
		a.log.Error(ctx, "unable to add user", zap.Error(err), zap.Any("user-info", param))
	}
	return &dto.User{
		ID:        data.ID,
		Email:     data.Email,
		Phone:     data.Phone.String,
		Tier:      data.Tier,
		CreatedAt: data.CreatedAt,
		UpdatedAt: data.UpdatedAt,
		Status:    data.Status,
	}, nil

}
