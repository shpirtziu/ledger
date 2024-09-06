// Code generated by Speakeasy (https://speakeasyapi.com). DO NOT EDIT.

package components

type Cursor struct {
	PageSize int64      `json:"pageSize"`
	HasMore  bool       `json:"hasMore"`
	Previous *string    `json:"previous,omitempty"`
	Next     *string    `json:"next,omitempty"`
	Data     []V2Ledger `json:"data"`
}

func (o *Cursor) GetPageSize() int64 {
	if o == nil {
		return 0
	}
	return o.PageSize
}

func (o *Cursor) GetHasMore() bool {
	if o == nil {
		return false
	}
	return o.HasMore
}

func (o *Cursor) GetPrevious() *string {
	if o == nil {
		return nil
	}
	return o.Previous
}

func (o *Cursor) GetNext() *string {
	if o == nil {
		return nil
	}
	return o.Next
}

func (o *Cursor) GetData() []V2Ledger {
	if o == nil {
		return []V2Ledger{}
	}
	return o.Data
}

type V2LedgerListResponse struct {
	Cursor Cursor `json:"cursor"`
}

func (o *V2LedgerListResponse) GetCursor() Cursor {
	if o == nil {
		return Cursor{}
	}
	return o.Cursor
}
