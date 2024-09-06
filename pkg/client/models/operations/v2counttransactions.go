// Code generated by Speakeasy (https://speakeasyapi.com). DO NOT EDIT.

package operations

import (
	"github.com/formancehq/stack/ledger/client/internal/utils"
	"github.com/formancehq/stack/ledger/client/models/components"
	"time"
)

type V2CountTransactionsRequest struct {
	// Name of the ledger.
	Ledger      string         `pathParam:"style=simple,explode=false,name=ledger"`
	Pit         *time.Time     `queryParam:"style=form,explode=true,name=pit"`
	RequestBody map[string]any `request:"mediaType=application/json"`
}

func (v V2CountTransactionsRequest) MarshalJSON() ([]byte, error) {
	return utils.MarshalJSON(v, "", false)
}

func (v *V2CountTransactionsRequest) UnmarshalJSON(data []byte) error {
	if err := utils.UnmarshalJSON(data, &v, "", false, false); err != nil {
		return err
	}
	return nil
}

func (o *V2CountTransactionsRequest) GetLedger() string {
	if o == nil {
		return ""
	}
	return o.Ledger
}

func (o *V2CountTransactionsRequest) GetPit() *time.Time {
	if o == nil {
		return nil
	}
	return o.Pit
}

func (o *V2CountTransactionsRequest) GetRequestBody() map[string]any {
	if o == nil {
		return nil
	}
	return o.RequestBody
}

type V2CountTransactionsResponse struct {
	HTTPMeta components.HTTPMetadata `json:"-"`
	Headers  map[string][]string
}

func (o *V2CountTransactionsResponse) GetHTTPMeta() components.HTTPMetadata {
	if o == nil {
		return components.HTTPMetadata{}
	}
	return o.HTTPMeta
}

func (o *V2CountTransactionsResponse) GetHeaders() map[string][]string {
	if o == nil {
		return map[string][]string{}
	}
	return o.Headers
}
