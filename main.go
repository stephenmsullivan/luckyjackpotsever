package main

import (
	"context"
	"database/sql"

	"github.com/heroiclabs/nakama-common/runtime"
)

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	logger.Info("Hello World!")

	// Create all the leaderboards
	CreateLeaderboard("multihandpoker_jackpots", ctx, logger, db, nk, initializer)
	CreateLeaderboard("videopokercasino_jackpots", ctx, logger, db, nk, initializer)

	return nil
}
