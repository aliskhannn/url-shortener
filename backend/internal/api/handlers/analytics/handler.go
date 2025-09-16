package analytics

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"

	"github.com/aliskhannn/url-shortener/internal/api/respond"
	"github.com/aliskhannn/url-shortener/internal/config"
	"github.com/aliskhannn/url-shortener/internal/model"
	analyticsrepo "github.com/aliskhannn/url-shortener/internal/repository/analytics"
)

// analyticsService defines the interface that the Handler depends on.
type analyticsService interface {
	GetAnalytics(ctx context.Context, strategy retry.Strategy, alias string) (model.Analytics, error)
}

// Handler handles HTTP requests related to link.
type Handler struct {
	analyticsService analyticsService
	cfg              *config.Config
}

// NewHandler creates a new Handler instance.
func NewHandler(
	as analyticsService,
	cfg *config.Config,
) *Handler {
	return &Handler{analyticsService: as, cfg: cfg}
}

func (h *Handler) GetAnalytics(c *ginext.Context) {
	alias := c.Param("alias")

	if alias == "" {
		zlog.Logger.Warn().Msg("missing alias")
		respond.Fail(c.Writer, http.StatusBadRequest, fmt.Errorf("missing alias"))
		return
	}

	event, err := h.analyticsService.GetAnalytics(c.Request.Context(), h.cfg.Retry, alias)
	if err != nil {
		if errors.Is(err, analyticsrepo.ErrAnalyticsNotFound) {
			zlog.Logger.Err(err).Str("alias", alias).Msg("link analytics not found")
			respond.Fail(c.Writer, http.StatusNotFound, fmt.Errorf("link analytics not found"))
			return
		}

		// Internal server error.
		zlog.Logger.Error().Err(err).Str("alias", alias).Msg("failed to get link analytics")
		respond.Fail(c.Writer, http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}

	respond.OK(c.Writer, event)
}
