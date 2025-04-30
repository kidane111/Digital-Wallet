package auth

import (
	"context"
	"net/http"
	"time"

	"digital-wallet/internal/constant/errors"
	"digital-wallet/internal/constant/model/dto"
	"digital-wallet/internal/constant/model/response"
	rest "digital-wallet/internal/handler"
	"digital-wallet/internal/module"
	"digital-wallet/platform/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type login struct {
	log            logger.Logger
	authModule     module.Auth
	contextTimeout time.Duration
}

func Init(log logger.Logger, authModule module.Auth,
	contextTimeout time.Duration) rest.Auth {
	return &login{
		log:            log,
		authModule:     authModule,
		contextTimeout: contextTimeout,
	}
}

// Login
//
//	@Summary		User Login
//	@Description	Login with phone/email and password.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			login	body	dto.LoginRequest	true	"Login credentials"
//	@Success		200		{object}	response.SuccessResponse{data=response.LoginResponse}
//	@Failure		400,401	{object}	response.ErrorResponse
//	@Router			/auth/login [post]
func (l *login) Login(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, l.contextTimeout)
	defer cancel()

	var req dto.Login
	if err := c.ShouldBindJSON(&req); err != nil {
		wrapped := errors.ErrInvalidUserInput.Wrap(err, "invalid login input")
		l.log.Error(ctx, "failed to bind login request", zap.Error(wrapped))
		_ = c.Error(wrapped)
		return
	}

	res, err := l.authModule.Login(ctx, req)
	if err != nil {
		l.log.Error(ctx, "login failed", zap.Error(err))
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, res)
}

// Register
//
//	@Summary		User Registration
//	@Description	Register a new user account
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			register	body		dto.RegisterRequest	true	"Registration data"
//	@Success		201			{object}	response.SuccessResponse{data=response.RegisterResponse}
//	@Failure		400,409		{object}	response.ErrorResponse
//	@Router			/auth/register [post]
func (l *login) Register(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, l.contextTimeout)
	defer cancel()

	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		wrapped := errors.ErrInvalidUserInput.Wrap(err, "invalid registration input")
		l.log.Error(ctx, "failed to bind registration request", zap.Error(wrapped))
		_ = c.Error(wrapped)
		return
	}

	res, err := l.authModule.UserRegister(ctx, req)
	if err != nil {
		l.log.Error(ctx, "registration failed", zap.Error(err))
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusCreated, res)
}
