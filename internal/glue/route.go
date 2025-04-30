package glue

import (
	"context"
	"digital-wallet/internal/constant/state"
	"digital-wallet/platform/logger"
	"fmt"

	"github.com/gin-gonic/gin"
)

type Router struct {
	Method      string
	Path        string
	Handler     gin.HandlerFunc
	Middlewares []gin.HandlerFunc
	Domain      []state.Domain
	UnAuthorize bool
}

func RegisterRoute(
	grp *gin.RouterGroup,
	routes []Router,
	zapLogger logger.Logger,
	authDomains state.AuthDomains,
) {
	for _, route := range routes {
		for _, domain := range route.Domain {
			var handler []gin.HandlerFunc

			var endpoint string

			switch domain.Name {
			case authDomains.User.Name:
				endpoint = route.Path

			default:
				zapLogger.Fatal(context.Background(),
					fmt.Sprintf("Invalid Domain %s Registered on Route", domain))
			}
			handler = append(handler, route.Middlewares...)
			handler = append(handler, route.Handler)
			grp.Handle(route.Method, endpoint, handler...)
		}
	}
}
