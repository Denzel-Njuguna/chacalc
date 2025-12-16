package datatypes

import (
	"github.com/cockroachdb/apd"
)

type Indcontr struct {
	Userid      string      `json:"userid"`
	Stockamount string      `json:"stockamount"`
	Stockname   string      `json:"stockname"`
	Buyingprice apd.Decimal `json:"buyprice"`
}
