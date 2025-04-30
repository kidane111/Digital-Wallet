package tire

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
)

type trie struct {
	log            logger.Logger
	tireModule     module.TireAccess
	contextTimeout time.Duration
}

func Init(log logger.Logger, tireModule module.TireAccess,
	contextTimeout time.Duration) rest.Tire {
	return &trie{
		log:            log,
		tireModule:     tireModule,
		contextTimeout: contextTimeout,
	}
}

// GetTierLimits godoc
// @Summary      Get tier limits
// @Description  Get the limits for a specific tier
// @Tags         Tiers
// @Param        tier  path      string  true  "Tier name"
// @Success      200   {object}  domain.TierLimits
// @Failure      400   {object}  gin.H
// @Failure      500   {object}  gin.H
// @Router       /tiers/{tier}/limits [get]
func (h *trie) GetTierLimits(c *gin.Context) {
	tier := c.Param("tier")
	ctx, cancel := context.WithTimeout(c, h.contextTimeout)
	defer cancel()
	limits, err := h.tireModule.GetTierLimits(ctx, tier)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusCreated, limits)
}

// UpdateTierLimits godoc
// @Summary      Update tier limits
// @Description  Update the limits for a specific tier
// @Tags         Tiers
// @Param        tier   path      string              true  "Tier name"
// @Param        body   body      domain.TierLimits   true  "Tier limits data"
// @Success      202    {string}  string              "Accepted"
// @Failure      400    {object}  gin.H
// @Failure      500    {object}  gin.H
// @Router       /tiers/{tier}
func (h *trie) UpdateTireLimits(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, h.contextTimeout)
	defer cancel()
	_ = c.Param("tier")

	var limits dto.TierConfig
	if err := c.ShouldBindJSON(&limits); err != nil {
		_ = c.Error(errors.ErrInvalidUserInput.Wrap(err,
			"invalid tier limits data"))
		return
	}

	err := h.tireModule.UpdateTierLimits(ctx, limits)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusNoContent, nil)

}
