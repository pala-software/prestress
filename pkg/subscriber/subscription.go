package subscriber

import "encoding/json"

type Change struct {
	RowOperation string          `json:"op"`
	RowKey       json.RawMessage `json:"key"`
	RowData      json.RawMessage `json:"data"`
}

type Subscription struct {
	Change chan Change
}

func (Subscription) Details() map[string]string {
	return map[string]string{}
}
