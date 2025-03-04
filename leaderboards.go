package main

import (
	"context"
	"database/sql"

	"github.com/heroiclabs/nakama-common/runtime"
)

func CreateLeaderboard(name string, ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {

	id := name
	authoritative := false
	sort := "desc"
	operator := "set"
	reset := ""
	metadata := map[string]interface{}{}

	if err := nk.LeaderboardCreate(ctx, id, authoritative, sort, operator, reset, metadata, false); err != nil {
		// Handle error.
		logger.Info("Leadboard Creation Failed!")
	}

	return nil
}
