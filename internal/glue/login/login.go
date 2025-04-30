package event

import (
	"digital-wallet/internal/constant/state"
	routing "digital-wallet/internal/glue"
	rest "digital-wallet/internal/handler"
	"digital-wallet/internal/handler/middleware"
	"digital-wallet/platform/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitAutzRoute(
	group *gin.RouterGroup,
	login rest.Auth,
	log logger.Logger,
	authDomains state.AuthDomains,
	authMiddleware middleware.AuthMiddleware,

) {
	userRoute := []routing.Router{
		{
			Method:      http.MethodPost,
			Path:        "/login",
			Handler:     login.Login,
			Middlewares: []gin.HandlerFunc{},
			Domain: []state.Domain{
				authDomains.User,
			},
		},
		{
			Method:      http.MethodPost,
			Path:        "/register",
			Handler:     login.Login,
			Middlewares: []gin.HandlerFunc{},
			Domain: []state.Domain{
				authDomains.User,
			},
		},
		{
			Method:  http.MethodPatch,
			Path:    "/tire_config",
			Handler: login.Login,
			Middlewares: []gin.HandlerFunc{
				authMiddleware.Authenticate(),
				authMiddleware.TierAccessControl([]string{"premium",
					"enterprise"}), // Only allow these tiers
				authMiddleware.RateLimit(),
			},
			Domain: []state.Domain{
				authDomains.User,
			},
		},
	}
	routing.RegisterRoute(group, userRoute, log, authDomains)
}
