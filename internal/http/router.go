package http

import (
	"log/slog"
	"net/http"

	authmw "github.com/sopatech/afterwave.fm/internal/auth"
	"github.com/sopatech/afterwave.fm/internal/users"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func NewRouter(logger *slog.Logger, userH *users.Handler, jwtSecret string) http.Handler {
	mux := http.NewServeMux()

	wrap := func(h http.Handler) http.Handler {
		return chain(h,
			Recoverer(logger),
			RealIP,
			RequestLogger(logger),
		)
	}

	auth := authmw.Authenticate(jwtSecret)

	// Public
	mux.Handle("POST /signup", wrap(http.HandlerFunc(userH.Signup)))
	mux.Handle("POST /login", wrap(http.HandlerFunc(userH.Login)))

	// Protected
	mux.Handle("DELETE /account", wrap(auth(http.HandlerFunc(userH.DeleteAccount))))

	return otelhttp.NewHandler(mux, "http.server")
}
