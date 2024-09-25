package v2

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/formancehq/go-libs/api"
	"github.com/formancehq/go-libs/migrations"
	"github.com/formancehq/ledger/internal/api/common"
)

type Info struct {
	Name    string      `json:"name"`
	Storage StorageInfo `json:"storage"`
}

type StorageInfo struct {
	Migrations []migrations.Info `json:"migrations"`
}

func getLedgerInfo(w http.ResponseWriter, r *http.Request) {
	ledger := common.LedgerFromContext(r.Context())

	var err error
	res := Info{
		Name:    chi.URLParam(r, "ledger"),
		Storage: StorageInfo{},
	}
	res.Storage.Migrations, err = ledger.GetMigrationsInfo(r.Context())
	if err != nil {
		api.InternalServerError(w, r, err)
		return
	}

	api.Ok(w, res)
}