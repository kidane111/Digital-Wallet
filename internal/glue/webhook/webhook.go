package webhook

import (
	"digital-wallet/internal/constant/state"
	routing "digital-wallet/internal/glue"
	rest "digital-wallet/internal/handler"
	"digital-wallet/internal/handler/middleware"
	"digital-wallet/platform/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitWebhookRoute(
	group *gin.RouterGroup,
	webhook rest.Event,
	log logger.Logger,
	authDomains state.AuthDomains,
	authMiddleware middleware.AuthMiddleware,

) {
	webhookRoute := []routing.Router{
		{
			Method:  http.MethodPost,
			Path:    "/webhook/notify",
			Handler: webhook.HandleWebhook,
			Middlewares: []gin.HandlerFunc{
				authMiddleware.TierAccessControl([]string{"premium",
					"enterprise"}), // Only allow these tiers
				authMiddleware.VarifaySignature(), // varify signature
			},
			Domain: []state.Domain{
				authDomains.User,
			},
		},
	}
	routing.RegisterRoute(group, webhookRoute, log, authDomains)
}
