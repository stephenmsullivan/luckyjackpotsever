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

// lookupUserByReferral is an RPC function that looks up a user by referral code.
// It expects a JSON payload like: {"referralCode": "12345678"}.
func LookupUserByReferral(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	// Define a structure for the incoming JSON payload.
	var input struct {
		ReferralCode string `json:"referralCode"`
	}
	if err := json.Unmarshal([]byte(payload), &input); err != nil {
		return "", fmt.Errorf("error parsing payload: %v", err)
	}

	// Read the record from the "referral_codes" collection using the referral code as the key.
	records, err := nk.StorageRead(ctx, []*runtime.StorageRead{{
		Collection: "referral_codes",
		Key:        input.ReferralCode,
	}})
	if err != nil {
		return "", fmt.Errorf("error reading referral_codes table: %v", err)
	}

	// If no record is found, return a not found response.
	if records == nil || len(records) == 0 {
		return "", fmt.Errorf("referral code not found")
	}

	// Assume we use the first record (there should be only one per unique referral code).
	// The record is stored as a map, and we expect it to contain the key "userId".
	userData := records[0].GetValue()

	// Parse the byte data into a map
	var userDataMap map[string]interface{}
	if err := json.Unmarshal([]byte(userData), &userDataMap); err != nil {
		return "", fmt.Errorf("error parsing user data: %v", err)
	}

	// Extract the userId
	userId, ok := userDataMap["userId"].(string)
	if !ok || userId == "" {
		return "", fmt.Errorf("userId not found in referral record")
	}

	// Return the found user information as JSON string
	response := map[string]interface{}{
		"found":  true,
		"userId": userId,
	}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("error marshaling response: %v", err)
	}

	// lookup the user account
	account, err := nk.AccountGetId(ctx, userId)
	if err != nil {
		return "", fmt.Errorf("error retrieving account for user %s: %v", userId, err)
	}

	// get the user metadata for the response
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(account.User.Metadata), &metadata); err != nil {
		return "", fmt.Errorf("error unmarshaling metadata: %v", err)
	}
	response["metadata"] = metadata

	return string(jsonResponse), nil
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
