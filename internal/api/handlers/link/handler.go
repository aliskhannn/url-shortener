package link

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mssola/user_agent"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"

	"github.com/aliskhannn/url-shortener/internal/api/respond"
	"github.com/aliskhannn/url-shortener/internal/config"
	"github.com/aliskhannn/url-shortener/internal/model"
	linkrepo "github.com/aliskhannn/url-shortener/internal/repository/link"
	linksvc "github.com/aliskhannn/url-shortener/internal/service/link"
)

// linkService defines the interface that the Handler depends on.
type linkService interface {
	CreateLink(ctx context.Context, strategy retry.Strategy, link model.Link) (model.Link, error)
	GetLinkByAlias(ctx context.Context, strategy retry.Strategy, alias string) (model.Link, error)
}

// analyticsService defines the interface that the Handler depends on.
type analyticsService interface {
	SaveAnalytics(ctx context.Context, strategy retry.Strategy, event model.Analytics) (uuid.UUID, error)
}

// Handler handles HTTP requests related to link.
type Handler struct {
	ctx              context.Context
	cfg              *config.Config
	validator        *validator.Validate
	linkService      linkService
	analyticsService analyticsService
}

// NewHandler creates a new Handler instance.
func NewHandler(
	ctx context.Context,
	cfg *config.Config,
	v *validator.Validate,
	ls linkService,
	as analyticsService,
) *Handler {
	return &Handler{ctx: ctx, cfg: cfg, validator: v, linkService: ls, analyticsService: as}
}

// CreateRequest represents the expected JSON payload for creating a shortened link.
type CreateRequest struct {
	URL   string `json:"url" validate:"required"`
	Alias string `json:"alias"`
}

// ShortenLink handles POST /shorten requests.
// It validates input, creates a new short link, and returns it.
func (h *Handler) ShortenLink(c *ginext.Context) {
	var req CreateRequest

	// Decode JSON request body into CreateRequest struct.
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		zlog.Logger.Err(err).Msg("failed to decode request body")
		respond.Fail(c.Writer, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	// Validate request fields using go-playground/validator.
	if err := h.validator.Struct(req); err != nil {
		zlog.Logger.Warn().Err(err).Msg("failed to validate request body")
		respond.Fail(c.Writer, http.StatusBadRequest, fmt.Errorf("validation error: %s", err.Error()))
		return
	}

	// Construct a Link model.
	link := model.Link{
		URL:   req.URL,
		Alias: req.Alias,
	}

	// Create a shorted link using the service layer.
	res, err := h.linkService.CreateLink(c.Request.Context(), h.cfg.Retry, link)
	if err != nil {
		// Handle duplicate alias.
		if errors.Is(err, linksvc.ErrAliasAlreadyExists) {
			zlog.Logger.Error().Err(err).Msg("alias already exists")
			respond.Fail(c.Writer, http.StatusConflict, fmt.Errorf("alias already exists"))
			return
		}

		// Internal errors.
		zlog.Logger.Error().Err(err).Str("alias", link.Alias).Msg("failed to shorten link")
		respond.Fail(c.Writer, http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}

	respond.Created(c.Writer, res)
}

// RedirectLink handles GET /s/:alias requests.
// It resolves a short alias to the original URL, saves analytics, and redirects the user.
func (h *Handler) RedirectLink(c *ginext.Context) {
	alias := c.Param("alias")

	if alias == "" {
		zlog.Logger.Warn().Msg("missing alias")
		respond.Fail(c.Writer, http.StatusBadRequest, fmt.Errorf("missing alias"))
		return
	}

	// Lookup the link in the service (cache → DB).
	link, err := h.linkService.GetLinkByAlias(c.Request.Context(), h.cfg.Retry, alias)
	if err != nil {
		// Handle case when alias does not exist.
		if errors.Is(err, linkrepo.ErrAliasNotFound) {
			zlog.Logger.Err(err).Str("alias", alias).Msg("alias not found")
			respond.Fail(c.Writer, http.StatusNotFound, fmt.Errorf("alias not found"))
			return
		}

		// Internal errors.
		zlog.Logger.Error().Err(err).Str("alias", alias).Msg("failed to get link")
		respond.Fail(c.Writer, http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}

	// Build an analytics event from the request.
	event := h.buildAnalytics(link.Alias, c.Request)
	zlog.Logger.Info().Interface("event", event).Msg("got event")

	// Save analytics asynchronously.
	go h.saveAnalyticsAsync(event)

	http.Redirect(c.Writer, c.Request, link.URL, http.StatusFound)
}

// saveAnalyticsAsync saves analytics data asynchronously using the global context.
func (h *Handler) saveAnalyticsAsync(event model.Analytics) {
	id, err := h.analyticsService.SaveAnalytics(h.ctx, h.cfg.Retry, event)
	if err != nil {
		zlog.Logger.Error().
			Err(err).
			Str("alias", event.Alias).
			Msg("failed to save analytics")
		return
	}

	zlog.Logger.Info().Str("id", id.String()).Msg("saved analytics")
}

// buildAnalytics constructs an Analytics model from the HTTP request.
func (h *Handler) buildAnalytics(alias string, r *http.Request) model.Analytics {
	ua := user_agent.New(r.UserAgent())

	// Detect browser name.
	browserName, _ := ua.Browser()

	// Detect a device type.
	device := "desktop"
	if ua.Mobile() {
		device = "mobile"
	} else if ua.Bot() {
		device = "bot"
	}

	// Extract client IP (RemoteAddr may include port).
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}

	return model.Analytics{
		Alias:     alias,
		UserAgent: r.UserAgent(),
		Device:    device,
		OS:        ua.OS(),
		Browser:   browserName,
		IP:        ip,
	}
}
