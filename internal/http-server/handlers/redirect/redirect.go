package redirect

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	resp "github.com/goget-milk/url-shortener/internal/lib/api/response"
	"github.com/goget-milk/url-shortener/internal/lib/logger/sl"
	storage "github.com/goget-milk/url-shortener/internal/stotrage"
	"log/slog"
	"net/http"
)

// URLGetter is an interface for getting url by alia
type URLGetter interface {
	GetURL(alias string) (string, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.50.0 --name=URLGetter
func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.redirect.New"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")
			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		resURL, err := urlGetter.GetURL(alias)

		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", "alias", alias)
			render.JSON(w, r, resp.Error("not found"))
			return
		}

		if err != nil {
			log.Error("failed to get url", sl.Err(err))
			render.JSON(w, r, resp.Error("internal error"))
			return
		}
		log.Info("got url", slog.String("url", resURL))
		// redirect to found url
		http.Redirect(w, r, resURL, http.StatusFound)
	}
}
