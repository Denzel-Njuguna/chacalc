package datatypes

import (
	"chacalc/src"
	"chacalc/src/databaseconn"
	"context"
	"time"

	"github.com/cockroachdb/apd"
	"github.com/jackc/pgx/v5"
)

// this is for an individual to sell their stock
type Soldstock struct {
	Userid      string      `json:"userid"`
	Stock       string      `json:"stockname"`
	Sellprice   apd.Decimal `json:"sellprice"`
	Stockamount apd.Decimal `json:"stockamount"`
}

// this function calculate the sale of stoc
func (sell *Soldstock) Calculatesale() string {
	conn := databaseconn.Connection()
	defer conn.Close()
	ctx, err := context.WithTimeout(context.Background(), 60*time.Second)
	if err != nil {
		src.Logger.Fatalf("there was an issue creating a context for calculate database query %v", err)
	}
	var currentcashamt apd.Decimal
	var previouscashamt apd.Decimal
	var currentstockamt apd.Decimal
	var previousstockamt apd.Decimal
	// here we are getting the particular stock from since we are using ticker symbol
	scanerr := conn.QueryRow(ctx, "select currentcashamt,previouscashamt, currentstockamt,previousstockamt from public.currentholds where userid=$1 and stockname=$2", sell.Userid, sell.Stock).Scan(&currentcashamt, &previouscashamt, &currentstockamt, &previousstockamt)
	// this is to check if there were no retrieved rows
	if scanerr != nil {
		if scanerr == pgx.ErrNoRows {
			src.Logger.Println(norecordsfound)
			return norecordsfound
		}
		src.Logger.Fatalf("there was an error querying the database %v", scanerr)
		return ""
	}

	var newstockamt, newcashamt, cashamt apd.Decimal
	apd.BaseContext.Mul(&cashamt, &sell.Sellprice, &sell.Stockamount)

	_, suberr := apd.BaseContext.Sub(&newcashamt, &currentcashamt, &cashamt)
	if suberr != nil {
		src.Logger.Fatalf("there was an issue subtracting the stock amounts in sale: %v", suberr)
	}
	// this is subtracting the sold stock from the existing stock
	_, stockerr := apd.BaseContext.Sub(&newstockamt, &currentstockamt, &sell.Stockamount)
	if stockerr != nil {
		src.Logger.Fatalf("could not calculate the remaining stock")
	}
	// here we now update the variables before writing to the database again
	previousstockamt = currentstockamt
	currentstockamt = newstockamt
	previouscashamt = currentcashamt
	currentcashamt = newcashamt

	cmdtag, upderr := conn.Exec(ctx, "update public.currentholds set currentcashamt=$1,previouscashamt=$2,currentstockamt=$3,previoustockamt= $4", currentcashamt, previouscashamt, currentstockamt, previousstockamt)
	if upderr != nil {
		src.Logger.Fatalf(updaterowerr, upderr)
		return ""
	}
	if cmdtag.RowsAffected() == 0 {
		src.Logger.Println("no rows were affected")
	}
	return "successfully updated the users account"

}
