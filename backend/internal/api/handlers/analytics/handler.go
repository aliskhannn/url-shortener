package analytics

import (
	"context"
	"fmt"
	"net/http"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"

	"github.com/aliskhannn/url-shortener/internal/api/respond"
	"github.com/aliskhannn/url-shortener/internal/config"
	analyticssvc "github.com/aliskhannn/url-shortener/internal/service/analytics"
)

// analyticsService defines the interface that the Handler depends on.
type analyticsService interface {
	GetAnalyticsSummary(ctx context.Context, strategy retry.Strategy, alias string) (*analyticssvc.SummaryOfAnalytics, error)
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

	summary, err := h.analyticsService.GetAnalyticsSummary(c.Request.Context(), h.cfg.Retry, alias)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("alias", alias).Msg("failed to get link analytics")
		respond.Fail(c.Writer, http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}

	respond.JSON(c.Writer, http.StatusOK, summary)
}
