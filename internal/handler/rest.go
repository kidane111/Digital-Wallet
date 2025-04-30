package rest

import (
	"github.com/gin-gonic/gin"
)

type Auth interface {
	Login(c *gin.Context)
	Register(c *gin.Context)
}

type Tire interface {
	GetTierLimits(c *gin.Context)
	UpdateTireLimits(c *gin.Context)
}
type Event interface {
	HandleWebhook(c *gin.Context)
}
