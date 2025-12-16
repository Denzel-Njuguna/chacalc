package datatypes

import "github.com/cockroachdb/apd"

type stocksummary struct {
	Stockticker string      `json:"stockticker"`
	Stockamount apd.Decimal `json:"stockamt"`
}

type Chamaholdings struct {
	Stocks map[string]stocksummary `json:"stocks"`
}

// stock: {uchm, 200}
