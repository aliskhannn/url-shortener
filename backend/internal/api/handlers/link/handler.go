package link

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

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
	linkService      linkService
	analyticsService analyticsService
	validator        *validator.Validate
	cfg              *config.Config
}

// NewHandler creates a new Handler instance.
func NewHandler(
	ls linkService,
	as analyticsService,
	v *validator.Validate,
	cfg *config.Config,
) *Handler {
	return &Handler{linkService: ls, analyticsService: as, validator: v, cfg: cfg}
}

type CreateRequest struct {
	URL   string `json:"url" validate:"required"`
	Alias string `json:"alias"`
}

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
		if errors.Is(err, linksvc.ErrAliasAlreadyExists) {
			zlog.Logger.Error().Err(err).Msg("alias already exists")
			respond.Fail(c.Writer, http.StatusConflict, fmt.Errorf("alias already exists"))
			return
		}

		zlog.Logger.Error().Err(err).Str("alias", link.Alias).Msg("failed to shorten link")
		respond.Fail(c.Writer, http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}

	// Respond with created link ID.
	respond.Created(c.Writer, res)
}

func (h *Handler) RedirectLink(c *ginext.Context) {
	alias := c.Param("alias")

	if alias == "" {
		zlog.Logger.Warn().Msg("missing alias")
		respond.Fail(c.Writer, http.StatusBadRequest, fmt.Errorf("missing alias"))
		return
	}

	link, err := h.linkService.GetLinkByAlias(c.Request.Context(), h.cfg.Retry, alias)
	if err != nil {
		if errors.Is(err, linkrepo.ErrAliasNotFound) {
			zlog.Logger.Err(err).Str("alias", alias).Msg("alias not found")
			respond.Fail(c.Writer, http.StatusNotFound, fmt.Errorf("alias not found"))
			return
		}

		// Internal server error.
		zlog.Logger.Error().Err(err).Str("alias", alias).Msg("failed to get link")
		respond.Fail(c.Writer, http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}

	event := h.buildAnalytics(link.Alias, c.Request)
	zlog.Logger.Info().Interface("event", event).Msg("got event")

	go h.saveAnalyticsAsync(event)

	http.Redirect(c.Writer, c.Request, link.URL, http.StatusFound)
}

func (h *Handler) saveAnalyticsAsync(event model.Analytics) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	id, err := h.analyticsService.SaveAnalytics(ctx, h.cfg.Retry, event)
	if err != nil {
		zlog.Logger.Error().
			Err(err).
			Str("alias", event.Alias).
			Msg("failed to save analytics")
	}

	zlog.Logger.Info().Str("id", id.String()).Msg("saved analytics")
}

func (h *Handler) buildAnalytics(alias string, r *http.Request) model.Analytics {
	ua := user_agent.New(r.UserAgent())

	browserName, _ := ua.Browser()
	device := "desktop"
	if ua.Mobile() {
		device = "mobile"
	} else if ua.Bot() {
		device = "bot"
	}

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
