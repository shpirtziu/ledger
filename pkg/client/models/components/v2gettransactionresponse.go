// Code generated by Speakeasy (https://speakeasyapi.com). DO NOT EDIT.

package components

type V2GetTransactionResponse struct {
	Data V2ExpandedTransaction `json:"data"`
}

func (o *V2GetTransactionResponse) GetData() V2ExpandedTransaction {
	if o == nil {
		return V2ExpandedTransaction{}
	}
	return o.Data
}
