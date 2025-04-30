package domain

import (
	"digital-wallet/initiator/foundation"
	autz "digital-wallet/internal/glue/login"
	tire "digital-wallet/internal/glue/tire"
	"digital-wallet/internal/glue/webhook"

	"digital-wallet/internal/handler/middleware"
	"digital-wallet/platform/logger"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag/example/basic/docs"
)

func InitRouter(
	group *gin.RouterGroup,
	handler HandlerLayer,
	log logger.Logger,
	module ModuleLayer,
	state foundation.State,
	redis redis.Client,

) {
	docs.SwaggerInfo.Schemes = viper.GetStringSlice("swagger.schemes")
	docs.SwaggerInfo.Host = viper.GetString("swagger.host")
	docs.SwaggerInfo.BasePath = "/v1"
	group.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	authMiddleware := middleware.InitAuthMiddleware(
		log.Named("auth-middleware"),
		middleware.AuthMiddlewareModules{
			Auth: module.user,
		},
		&redis, viper.GetInt("limit.rate"),
	)

	autz.InitAutzRoute(
		group,
		handler.autzHandler,
		log.Named("auth-route"),
		state.AuthDomains,
		authMiddleware,
	)
	tire.InitTireRoute(
		group,
		handler.tire,
		log.Named("tire-route"),
		state.AuthDomains,
		authMiddleware,
	)
	webhook.InitWebhookRoute(
		group,
		handler.webhook,
		log.Named("webhook-route"),
		state.AuthDomains,
		authMiddleware,
	)

}
