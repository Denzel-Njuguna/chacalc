package datatypes

import (
	"chacalc/src"
	"chacalc/src/databaseconn"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cockroachdb/apd"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// this is for an individual to sell their stock
type Soldstock struct {
	Userid      string      `json:"userid"`
	Chamaid     string      `json:"chamaid"`
	Ticker      string      `json:"stockname"`
	Stockamount apd.Decimal `json:"stockamount"`
}

// this function calculate the sale of stoc
func (sell *Soldstock) Calculatesale() string {
	conn := databaseconn.Connection()
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	var currentstockamt apd.Decimal
	var previousstockamt apd.Decimal
	// here we are getting the particular stock from since we are using ticker symbol
	scanerr := conn.QueryRow(ctx, "select currentstockamt,previousstockamt from public.currentholds where userid=$1 and stockname=$2", sell.Userid, sell.Stockamount).Scan(&currentstockamt, &previousstockamt)
	// this is to check if there were no retrieved rows
	if scanerr != nil {
		if scanerr == pgx.ErrNoRows {
			src.Logger.Println(norecordsfound)
			return norecordsfound
		}
		src.Logger.Fatalf("there was an error querying the database %v", scanerr)
		return ""
	}

	var newstockamt apd.Decimal
	// this is subtracting the sold stock from the existing stock
	_, stockerr := apd.BaseContext.Sub(&newstockamt, &currentstockamt, &sell.Stockamount)
	if stockerr != nil {
		src.Logger.Fatalf("could not calculate the remaining stock")
	}
	// here we now update the variables before writing to the database again
	previousstockamt = currentstockamt
	currentstockamt = newstockamt

	cmdtag, upderr := conn.Exec(ctx, "update public.currentholds set currentstockamt=$3,previoustockamt= $4", currentstockamt, previousstockamt)
	if upderr != nil {
		src.Logger.Fatalf(updaterowerr, upderr)
		return ""
	}
	if cmdtag.RowsAffected() == 0 {
		src.Logger.Println("no rows were affected")
	}
	src.Logger.Fatalln(Subtractchama(sell, conn, ctx))
	return "successfully updated the users account"
}
func Subtractchama(sale *Soldstock, conn *pgxpool.Pool, ctx context.Context) string {
	var chamaholdings []byte
	err := conn.QueryRow(ctx, "select chamaholdings from public.chamas where chamaid = $1", sale.Chamaid).Scan(&chamaholdings)
	if err != nil {
		src.Logger.Fatalf("there was an issue querying the database %v", err)
		return ""
	}
	var summary Chamaholdings
	if len(chamaholdings) > 0 {
		json.Unmarshal(chamaholdings, &summary)
	} else {
		summary = Chamaholdings{
			Stocks: make(map[string]stocksummary),
		}
	}
	stock := sale.Ticker
	current := summary.Stocks[stock]
	fmt.Println("before", current)
	apd.BaseContext.Sub(&current.Stockamount, &current.Stockamount, &sale.Stockamount)
	fmt.Println("after", current)
	summary.Stocks[stock] = current
	updatedjson, _ := json.Marshal(summary)
	cmdtag, uerr := conn.Exec(ctx, "update public.chamas set chamaholdings = $2 where chamaid = $1", sale.Chamaid, updatedjson)

	if uerr != nil {
		src.Logger.Fatalf("there was an issue updating the chamaholdings %v", uerr)
		return ""
	}

	if cmdtag.RowsAffected() == 0 {
		src.Logger.Println("No rows were affected")
		return ""
	}
	return "success"
}
