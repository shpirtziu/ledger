// Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.

package operations

import (
	"github.com/formancehq/stack/ledger/client/internal/utils"
	"github.com/formancehq/stack/ledger/client/models/components"
	"time"
)

type ListLogsRequest struct {
	// Name of the ledger.
	Ledger string `pathParam:"style=simple,explode=false,name=ledger"`
	// The maximum number of results to return per page.
	//
	PageSize *int64 `default:"15" queryParam:"style=form,explode=true,name=pageSize"`
	// Pagination cursor, will return the logs after a given ID. (in descending order).
	After *string `queryParam:"style=form,explode=true,name=after"`
	// Filter transactions that occurred after this timestamp.
	// The format is RFC3339 and is inclusive (for example, "2023-01-02T15:04:01Z" includes the first second of 4th minute).
	//
	StartTime *time.Time `queryParam:"style=form,explode=true,name=startTime"`
	// Filter transactions that occurred before this timestamp.
	// The format is RFC3339 and is exclusive (for example, "2023-01-02T15:04:01Z" excludes the first second of 4th minute).
	//
	EndTime *time.Time `queryParam:"style=form,explode=true,name=endTime"`
	// Parameter used in pagination requests. Maximum page size is set to 1000.
	// Set to the value of next for the next page of results.
	// Set to the value of previous for the previous page of results.
	// No other parameters can be set when this parameter is set.
	//
	Cursor *string `queryParam:"style=form,explode=true,name=cursor"`
}

func (l ListLogsRequest) MarshalJSON() ([]byte, error) {
	return utils.MarshalJSON(l, "", false)
}

func (l *ListLogsRequest) UnmarshalJSON(data []byte) error {
	if err := utils.UnmarshalJSON(data, &l, "", false, false); err != nil {
		return err
	}
	return nil
}

func (o *ListLogsRequest) GetLedger() string {
	if o == nil {
		return ""
	}
	return o.Ledger
}

func (o *ListLogsRequest) GetPageSize() *int64 {
	if o == nil {
		return nil
	}
	return o.PageSize
}

func (o *ListLogsRequest) GetAfter() *string {
	if o == nil {
		return nil
	}
	return o.After
}

func (o *ListLogsRequest) GetStartTime() *time.Time {
	if o == nil {
		return nil
	}
	return o.StartTime
}

func (o *ListLogsRequest) GetEndTime() *time.Time {
	if o == nil {
		return nil
	}
	return o.EndTime
}

func (o *ListLogsRequest) GetCursor() *string {
	if o == nil {
		return nil
	}
	return o.Cursor
}

type ListLogsResponse struct {
	HTTPMeta components.HTTPMetadata `json:"-"`
	// OK
	LogsCursorResponse *components.LogsCursorResponse
}

func (o *ListLogsResponse) GetHTTPMeta() components.HTTPMetadata {
	if o == nil {
		return components.HTTPMetadata{}
	}
	return o.HTTPMeta
}

func (o *ListLogsResponse) GetLogsCursorResponse() *components.LogsCursorResponse {
	if o == nil {
		return nil
	}
	return o.LogsCursorResponse
}