package webhook

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"digital-wallet/internal/constant/model/dto"
	"digital-wallet/internal/constant/model/response"
	rest "digital-wallet/internal/handler"
	"digital-wallet/internal/module"
	"digital-wallet/platform/logger"

	"github.com/gin-gonic/gin"
)

type webhook struct {
	log            logger.Logger
	eventModule    module.Event
	contextTimeout time.Duration
}

// Init initializes the webhook handler
func Init(log logger.Logger, eventModule module.Event, contextTimeout time.Duration) rest.Event {
	return &webhook{
		log:            log,
		eventModule:    eventModule,
		contextTimeout: contextTimeout,
	}
}

// HandleWebhook handles incoming webhook events
// @Summary      Handle Webhook
// @Description  Process incoming webhook events
// @Tags         Webhook
// @Accept       json
// @Produce      json
// @Success      200   {object}  gin.H
// @Failure      400   {object}  gin.H
// @Failure      500   {object}  gin.H
// @Router       /webhook [post]
func (h *webhook) HandleWebhook(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.contextTimeout)
	defer cancel()

	// Read and parse the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		_ = c.Error(err)
	}

	var payload dto.Payload
	if err := json.Unmarshal(body, &payload); err != nil {
		_ = c.Error(err)

	}

	data, err := h.eventModule.HandleWebhook(ctx, payload)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusCreated, data)

}
