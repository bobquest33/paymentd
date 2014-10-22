package currency

import (
	"encoding/json"
)

// Currency represents a currency in ISO4217 code
type Currency struct {
	CodeISO4217 string
}

// Empty returns true if the currency is considered empty/uninitialized
func (c Currency) Empty() bool {
	return c.CodeISO4217 == ""
}

func (c Currency) MarshalJSON() ([]byte, error) {

	return json.Marshal(c.CodeISO4217)

}
