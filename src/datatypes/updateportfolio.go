package datatypes

import (
	"chacalc/src"
	"chacalc/src/databaseconn"
	"context"
	"time"
)

// update portfolio
type updateport struct {
	Userid   string `json:"userid"`
	Chamaid  string `json:"chamaid"`
	Stockamt string `json:"stockamt"`
	Ticker   string `json:"ticker"`
}

func (update *updateport) Checker() string {
	conn := databaseconn.Connection()
	ctx, ctxerr := context.WithTimeout(context.Background(), 60*time.Second)
	if ctxerr != nil {
		src.Logger.Fatalf("there was an error creating context for updating portfolio", ctxerr)
		return ""
	}
	row, qerr := conn.Query(ctx, "select * from public.holdings where userid=$1,chamaid=$2,ticker=$3", update.Userid, update.Chamaid, update.Ticker)
	if qerr != nil {
		src.Logger.Fatal("there was an error contacting db in update portfolio")
	}
	if row.CommandTag().RowsAffected() == 0 {

	}
}
