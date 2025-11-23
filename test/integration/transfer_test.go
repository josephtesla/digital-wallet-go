package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/josephtesla/digital-wallet-go/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransferFunds(t *testing.T) {
	router := setupTestServer(t)
	db := setupTestDB(t)

	// Create test users and wallets
	user1 := &models.User{
		Name:     "User 1",
		Email:    "user1@example.com",
		Password: "hashed_password",
	}
	err := db.Create(user1).Error
	require.NoError(t, err)

	user2 := &models.User{
		Name:     "User 2",
		Email:    "user2@example.com",
		Password: "hashed_password",
	}
	err = db.Create(user2).Error
	require.NoError(t, err)

	wallet1 := &models.Wallet{
		UserID:   user1.ID,
		Currency: "NGN",
		Balance:  200000, // 2000 NGN in kobo
		Status:   "active",
	}
	err = db.Create(wallet1).Error
	require.NoError(t, err)

	wallet2 := &models.Wallet{
		UserID:   user2.ID,
		Currency: "NGN",
		Balance:  100000, // 1000 NGN in kobo
		Status:   "active",
	}
	err = db.Create(wallet2).Error
	require.NoError(t, err)

	// Update users with wallet IDs
	user1.WalletID = wallet1.ID
	user2.WalletID = wallet2.ID
	err = db.Save(user1).Error
	require.NoError(t, err)
	err = db.Save(user2).Error
	require.NoError(t, err)

	// Transfer request
	reqBody := map[string]interface{}{
		"from_wallet_id": wallet1.ID.String(),
		"to_wallet_id":   wallet2.ID.String(),
		"amount":         50000, // 500 NGN in kobo
		"description":    "Test transfer",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/transfers", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.Equal(t, "Transfer completed successfully", response["message"])

	// Verify balances
	var updatedWallet1, updatedWallet2 models.Wallet
	err = db.First(&updatedWallet1, wallet1.ID).Error
	require.NoError(t, err)
	err = db.First(&updatedWallet2, wallet2.ID).Error
	require.NoError(t, err)

	assert.Equal(t, int64(150000), updatedWallet1.Balance) // 2000 - 500 = 1500 NGN
	assert.Equal(t, int64(150000), updatedWallet2.Balance) // 1000 + 500 = 1500 NGN
}

func TestInsufficientFundsTransfer(t *testing.T) {
	router := setupTestServer(t)
	db := setupTestDB(t)

	// Create test users and wallets
	user1 := &models.User{
		Name:     "User 1",
		Email:    "user1@example.com",
		Password: "hashed_password",
	}
	err := db.Create(user1).Error
	require.NoError(t, err)

	user2 := &models.User{
		Name:     "User 2",
		Email:    "user2@example.com",
		Password: "hashed_password",
	}
	err = db.Create(user2).Error
	require.NoError(t, err)

	wallet1 := &models.Wallet{
		UserID:   user1.ID,
		Currency: "NGN",
		Balance:  10000, // 100 NGN in kobo
		Status:   "active",
	}
	err = db.Create(wallet1).Error
	require.NoError(t, err)

	wallet2 := &models.Wallet{
		UserID:   user2.ID,
		Currency: "NGN",
		Balance:  100000, // 1000 NGN in kobo
		Status:   "active",
	}
	err = db.Create(wallet2).Error
	require.NoError(t, err)

	// Update users with wallet IDs
	user1.WalletID = wallet1.ID
	user2.WalletID = wallet2.ID
	err = db.Save(user1).Error
	require.NoError(t, err)
	err = db.Save(user2).Error
	require.NoError(t, err)

	// Transfer request with insufficient funds
	reqBody := map[string]interface{}{
		"from_wallet_id": wallet1.ID.String(),
		"to_wallet_id":   wallet2.ID.String(),
		"amount":         50000, // 500 NGN in kobo (more than available)
		"description":    "Test transfer",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/transfers", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.Contains(t, response["message"].(string), "Transfer failed")
}
