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

// update portfolio increase portfolio
type Updateport struct {
	Userid   string      `json:"userid"`
	Chamaid  string      `json:"chamaid"`
	Stockamt apd.Decimal `json:"stockamt"`
	Ticker   string      `json:"ticker"`
}

// this checks if the stock is there and if not it creates it for the user
func (update *Updateport) Checker() string {
	conn := databaseconn.Connection()
	ctx, ctxerr := context.WithTimeout(context.Background(), 60*time.Second)
	defer ctxerr()
	var stockamt apd.Decimal
	qerr := conn.QueryRow(ctx, "select stockamt from public.holdings where userid=$1,chamaid=$2,ticker=$3", update.Userid, update.Chamaid, update.Ticker).Scan(&stockamt)
	if qerr != nil {
		if qerr == pgx.ErrNoRows {
			src.Logger.Println(norecordsfound)

			cmdtag, ierr := conn.Exec(ctx, "insert into public.holdings(userid,chamaid,ticker,quantity,currentstockamt,previousstockamt)values ($1,$2,$3,$4,$5,$6)", update.Userid, update.Chamaid, update.Ticker, update.Stockamt, update.Stockamt, 0.0)
			if ierr != nil {
				src.Logger.Fatalf("there was an issue inserting into the database")
				return ""
			}
			if cmdtag.RowsAffected() == 0 {
				src.Logger.Fatalf("Unable to create this users stock")
				return ""
			}
		}
		src.Logger.Fatalf("there was an error contacting db in update portfolio %v", qerr)
		return ""
	}
	Updateholdings(update, conn, ctx, stockamt)

	return "success"
}

// this is to update individual holding by adding to the existing ones
func Updateholdings(update *Updateport, conn *pgxpool.Pool, ctx context.Context, stockamt apd.Decimal) string {

	var newstockamt apd.Decimal
	apd.BaseContext.Add(&newstockamt, &stockamt, &update.Stockamt)
	cmdtag, uerr := conn.Exec(ctx, "update public.holdings set currentstockamt=$1, previousstockamt=$2 where userid=$3,chamaid=$4,ticker=$5",
		newstockamt, stockamt, update.Userid, update.Chamaid, update.Ticker)
	if uerr != nil {
		src.Logger.Fatalf("there was an error updating the holding amount in update portfolio %v", uerr)
		return ""
	}
	if cmdtag.RowsAffected() == 0 {
		src.Logger.Fatalln("update failed in updateportfolio.updateholdings")
		return ""
	}
	updatechama(update, conn, ctx)
	return "success"
}

// this is to add the persons stock to the chama's stock portfolio
func updatechama(update *Updateport, conn *pgxpool.Pool, ctx context.Context) string {
	/*
		take the chama id and go to the chama table
		retrieve the holdings column update the particular stock
	*/
	var chamaholdings []byte

	qerr := conn.QueryRow(ctx, "select chamaholdings from public.chamas where chamaid = $1", update.Chamaid).Scan(&chamaholdings)
	if qerr != nil {
		src.Logger.Fatalf("There was an issue querying the database %v", qerr)
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

	stock := update.Ticker
	current := summary.Stocks[stock]
	fmt.Printf("before:%v", current)
	apd.BaseContext.Add(&current.Stockamount, &current.Stockamount, &update.Stockamt)
	fmt.Println("after", current)
	summary.Stocks[stock] = current
	updatedjson, _ := json.Marshal(summary)

	cmdtag, uerr := conn.Exec(ctx, "update public.chamas set chamaholdings = $2 where chamaid = $1", update.Chamaid, updatedjson)

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
