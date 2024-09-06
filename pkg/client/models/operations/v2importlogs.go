// Code generated by Speakeasy (https://speakeasyapi.com). DO NOT EDIT.

package operations

import (
	"github.com/formancehq/stack/ledger/client/models/components"
)

type V2ImportLogsRequest struct {
	// Name of the ledger.
	Ledger      string  `pathParam:"style=simple,explode=false,name=ledger"`
	RequestBody *string `request:"mediaType=application/octet-stream"`
}

func (o *V2ImportLogsRequest) GetLedger() string {
	if o == nil {
		return ""
	}
	return o.Ledger
}

func (o *V2ImportLogsRequest) GetRequestBody() *string {
	if o == nil {
		return nil
	}
	return o.RequestBody
}

type V2ImportLogsResponse struct {
	HTTPMeta components.HTTPMetadata `json:"-"`
}

func (o *V2ImportLogsResponse) GetHTTPMeta() components.HTTPMetadata {
	if o == nil {
		return components.HTTPMetadata{}
	}
	return o.HTTPMeta
}
