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

func InitTireRoute(
	group *gin.RouterGroup,
	trie rest.Tire,
	log logger.Logger,
	authDomains state.AuthDomains,
	authMiddleware middleware.AuthMiddleware,

) {
	tireRoute := []routing.Router{
		{
			Method:  http.MethodPatch,
			Path:    "/limit",
			Handler: trie.GetTierLimits,
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

		{
			Method:  http.MethodPatch,
			Path:    "/tire_config",
			Handler: trie.UpdateTireLimits,
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
	routing.RegisterRoute(group, tireRoute, log, authDomains)
}
