package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/heroiclabs/nakama-common/runtime"
)

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	logger.Info("Hello World!")

	// Create all the leaderboards
	CreateLeaderboard("multihandpoker_jackpots", ctx, logger, db, nk, initializer)
	CreateLeaderboard("videopokercasino_jackpots", ctx, logger, db, nk, initializer)
	CreateLeaderboard("kenocasino_jackpots", ctx, logger, db, nk, initializer)

	// register the RPC calls
	if err := initializer.RegisterRpc("lookupUserByReferral", lookupUserByReferral); err != nil {
		return err
	}

	initializer.RegisterBeforeGetAccount(beforeGetAccount)

	return nil
}

// generateReferralCode generates a random 8-digit referral code as a string.
func generateReferralCode() string {
	// Generate a random number between 10000000 and 99999999.
	num := rand.Intn(90000000) + 10000000
	return strconv.Itoa(num)
}

// generateUniqueReferralCode generates an 8-digit referral code and ensures it is unique
// by checking the "referral_codes" storage collection.
func generateUniqueReferralCode(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, maxAttempts int) (string, error) {
	for i := 0; i < maxAttempts; i++ {
		// Generate a random 8-digit number as a string.
		code := generateReferralCode() // e.g., returns something like "12345678"

		// Use the StorageRead API to check if this referral code already exists.
		records, err := nk.StorageRead(ctx, []*runtime.StorageRead{{
			Collection: "referral_codes",
			Key:        code,
		}})
		if err != nil {
			logger.Error("Error reading storage for referral code '%s': %v", code, err)
			continue // In case of an error, try again.
		}

		// If no records are found, the code is unique.
		if len(records) == 0 {
			return code, nil
		}
	}
	return "", fmt.Errorf("failed to generate a unique referral code after %d attempts", maxAttempts)
}

// beforeGetAccount is a hook that runs before the GetAccount function.
func beforeGetAccount(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) error {

	// get the user id from the session
	userID := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)

	// Retrieve the account to check for existing referral code.
	account, err := nk.AccountGetId(ctx, userID)
	if err != nil {
		logger.Error("sessionCreated: error retrieving account for user %s: %v", userID, err)
		return err
	}

	// Check if the user's metadata already includes a referral code.
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(account.User.Metadata), &metadata); err != nil {
		logger.Error("Error unmarshaling metadata: %v", err)
		return err
	}
	if metadata != nil {
		if code, ok := metadata["referralCode"].(string); ok {
			logger.Info("User %s already has referralCode: %v", userID, code)
			return nil // Nothing to update.
		}
	}

	// Generate a unique referral code.
	newCode, err := generateUniqueReferralCode(ctx, logger, nk, 10)
	if err != nil {
		logger.Error("Error generating referral code for user %s: %v", userID, err)
		return err
	}

	// Update the user's metadata with the new referral code.
	var metadataUpdate = map[string]interface{}{
		"referralCode": newCode,
	}

	// add the referral code to a storage collection
	if _, err := nk.StorageWrite(ctx, []*runtime.StorageWrite{
		{
			Collection:      "referral_codes",
			Key:             newCode,
			PermissionRead:  1,
			PermissionWrite: 1,
			Value:           `{"userId":"` + userID + `"}`,
		},
	}); err != nil {
		logger.Error("sessionCreated: error writing referral code to storage: %v", err)
		return err
	}

	// update just the metadata
	if err := nk.AccountUpdateId(ctx, userID, "", metadataUpdate, "", "", "", "", ""); err != nil {
		logger.Error("sessionCreated: error updating user metadata for user %s: %v", userID, err)
		return err
	}

	logger.Info("sessionCreated: set referralCode for user %s: %s", userID, newCode)
	return nil
}
