package common

import (
	"net/http"
	"strings"

	"github.com/formancehq/ledger/internal/controller/system"
	"github.com/formancehq/ledger/internal/tracing"

	"github.com/formancehq/go-libs/api"
	"github.com/formancehq/go-libs/platform/postgres"

	"github.com/pkg/errors"
)

const (
	ErrOutdatedSchema = "OUTDATED_SCHEMA"
)

func LedgerMiddleware(
	backend system.Controller,
	resolver func(*http.Request) string,
	excludePathFromSchemaCheck ...string,
) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			name := resolver(r)
			if name == "" {
				api.NotFound(w, errors.New("empty name"))
				return
			}

			ctx, span := tracing.Start(r.Context(), "OpenLedger")
			defer span.End()

			var err error
			l, err := backend.GetLedgerController(ctx, name)
			if err != nil {
				switch {
				case postgres.IsNotFoundError(err):
					api.WriteErrorResponse(w, http.StatusNotFound, "LEDGER_NOT_FOUND", err)
				default:
					api.InternalServerError(w, r, err)
				}
				return
			}
			ctx = ContextWithLedger(ctx, l)

			pathWithoutLedger := r.URL.Path[1:]
			nextSlash := strings.Index(pathWithoutLedger, "/")
			if nextSlash >= 0 {
				pathWithoutLedger = pathWithoutLedger[nextSlash:]
			} else {
				pathWithoutLedger = ""
			}

			excluded := false
			for _, path := range excludePathFromSchemaCheck {
				if pathWithoutLedger == path {
					excluded = true
					break
				}
			}

			if !excluded {
				isUpToDate, err := l.IsDatabaseUpToDate(ctx)
				if err != nil {
					api.InternalServerError(w, r, err)
					return
				}
				if !isUpToDate {
					api.BadRequest(w, ErrOutdatedSchema, errors.New("You need to upgrade your ledger schema to the last version"))
					return
				}
			}

			handler.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}