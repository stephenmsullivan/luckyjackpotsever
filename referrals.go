package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
)

// lookupUserByReferral is an RPC function that looks up a user by referral code.
// It expects a JSON payload like: {"referralCode": "12345678"}.
func lookupUserByReferral(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
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
	return string(jsonResponse), nil
}
