package datatypes

import (
	"chacalc/src"
	"chacalc/src/databaseconn"
	"context"
	"time"

	"github.com/cockroachdb/apd"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// update portfolio increase portfolio
type updateport struct {
	Userid   string      `json:"userid"`
	Chamaid  string      `json:"chamaid"`
	Stockamt apd.Decimal `json:"stockamt"`
	Ticker   string      `json:"ticker"`
}

// this checks if the stock is there and if not it creates it for the user
func (update *updateport) Checker() string {
	conn := databaseconn.Connection()
	ctx, ctxerr := context.WithTimeout(context.Background(), 60*time.Second)
	if ctxerr != nil {
		src.Logger.Fatalf("there was an error creating context for updating portfolio", ctxerr)
		return ""
	}
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
func Updateholdings(update *updateport, conn *pgxpool.Pool, ctx context.Context, stockamt apd.Decimal) string {

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
	return "success"
}
