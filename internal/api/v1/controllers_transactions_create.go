package v1

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	ledgercontroller "github.com/formancehq/ledger/internal/controller/ledger"

	"github.com/formancehq/go-libs/api"
	"github.com/formancehq/go-libs/metadata"
	"github.com/formancehq/go-libs/time"
	ledger "github.com/formancehq/ledger/internal"
	"github.com/formancehq/ledger/internal/api/common"
	"github.com/pkg/errors"
)

type Script struct {
	ledger.Script
	Vars map[string]json.RawMessage `json:"vars"`
}

func (s Script) ToCore() (*ledger.Script, error) {
	s.Script.Vars = map[string]string{}
	for k, v := range s.Vars {

		m := make(map[string]json.RawMessage)
		if err := json.Unmarshal(v, &m); err != nil {
			var rawValue string
			if err := json.Unmarshal(v, &rawValue); err != nil {
				panic(err)
			}
			s.Script.Vars[k] = rawValue
			continue
		}

		// Is a monetary
		var asset string
		if err := json.Unmarshal(m["asset"], &asset); err != nil {
			return nil, errors.Wrap(err, "unmarshalling asset")
		}
		amount := &big.Int{}
		if err := json.Unmarshal(m["amount"], amount); err != nil {
			return nil, errors.Wrap(err, "unmarshalling amount")
		}

		s.Script.Vars[k] = fmt.Sprintf("%s %s", asset, amount)
	}
	return &s.Script, nil
}

type CreateTransactionRequest struct {
	Postings  ledger.Postings   `json:"postings"`
	Script    Script            `json:"script"`
	Timestamp time.Time         `json:"timestamp"`
	Reference string            `json:"reference"`
	Metadata  metadata.Metadata `json:"metadata" swaggertype:"object"`
}

func createTransaction(w http.ResponseWriter, r *http.Request) {
	l := common.LedgerFromContext(r.Context())

	payload := CreateTransactionRequest{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		api.BadRequest(w, ErrValidation, errors.New("invalid transaction format"))
		return
	}

	if len(payload.Postings) > 0 && payload.Script.Plain != "" ||
		len(payload.Postings) == 0 && payload.Script.Plain == "" {
		api.BadRequest(w, ErrValidation, errors.New("invalid payload: should contain either postings or script"))
		return
	} else if len(payload.Postings) > 0 {
		if _, err := payload.Postings.Validate(); err != nil {
			api.BadRequest(w, ErrValidation, err)
			return
		}
		txData := ledger.TransactionData{
			Postings:  payload.Postings,
			Timestamp: payload.Timestamp,
			Reference: payload.Reference,
			Metadata:  payload.Metadata,
		}

		res, err := l.CreateTransaction(r.Context(), getCommandParameters(r), ledger.TxToScriptData(txData, false))
		if err != nil {
			switch {
			case errors.Is(err, &ledgercontroller.ErrInsufficientFunds{}):
				api.BadRequest(w, ErrInsufficientFund, err)
			case errors.Is(err, &ledgercontroller.ErrInvalidVars{}) || errors.Is(err, ledgercontroller.ErrCompilationFailed{}):
				api.BadRequest(w, ErrScriptCompilationFailed, err)
			case errors.Is(err, &ledgercontroller.ErrMetadataOverride{}):
				api.BadRequest(w, ErrScriptMetadataOverride, err)
			case errors.Is(err, ledgercontroller.ErrNoPostings):
				api.BadRequest(w, ErrValidation, err)
			case errors.Is(err, ledgercontroller.ErrReferenceConflict{}):
				api.WriteErrorResponse(w, http.StatusConflict, ErrConflict, err)
			default:
				api.InternalServerError(w, r, err)
			}
			return
		}

		api.Ok(w, []any{mapTransactionToV1(*res)})
		return
	}

	script, err := payload.Script.ToCore()
	if err != nil {
		api.BadRequest(w, ErrValidation, err)
		return
	}

	runScript := ledger.RunScript{
		Script:    *script,
		Timestamp: payload.Timestamp,
		Reference: payload.Reference,
		Metadata:  payload.Metadata,
	}

	// todo: handle missing error cases
	res, err := l.CreateTransaction(r.Context(), getCommandParameters(r), runScript)
	if err != nil {
		switch {
		case errors.Is(err, &ledgercontroller.ErrInsufficientFunds{}):
			api.BadRequest(w, ErrInsufficientFund, err)
			return
		default:
			api.InternalServerError(w, r, err)
			return
		}
		//switch {
		//case ledgercontroller.IsCommandError(err):
		//	//switch {
		//	//case command.IsErrMachine(err):
		//	//	switch {
		//	//	case machine.IsInsufficientFundError(err):
		//	//		api.BadRequest(w, ErrInsufficientFund, err)
		//	//		return
		//	//	}
		//	//case command.IsInvalidTransactionError(err, command.ErrInvalidTransactionCodeConflict):
		//	//	api.BadRequest(w, ErrConflict, err)
		//	//	return
		//	//case command.IsInvalidTransactionError(err, command.ErrInvalidTransactionCodeCompilationFailed):
		//	//	api.BadRequestWithDetails(w, ErrScriptCompilationFailed, err, backend.EncodeLink(err.Error()))
		//	//	return
		//	//}
		//	api.BadRequest(w, ErrValidation, err)
		//	return
		//}
		//api.InternalServerError(w, r, err)
		//return
	}

	api.Ok(w, []any{mapTransactionToV1(*res)})
}