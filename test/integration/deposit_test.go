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

func TestInitializeDeposit(t *testing.T) {
	router := setupTestServer(t)
	db := setupTestDB(t)

	// Create test user and wallet
	user := &models.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "hashed_password",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	wallet := &models.Wallet{
		UserID:   user.ID,
		Currency: "NGN",
		Balance:  0,
		Status:   "active",
	}
	err = db.Create(wallet).Error
	require.NoError(t, err)

	// Update user with wallet ID
	user.WalletID = wallet.ID
	err = db.Save(user).Error
	require.NoError(t, err)

	// Deposit request
	reqBody := map[string]interface{}{
		"user_id":   user.ID.String(),
		"wallet_id": wallet.ID.String(),
		"amount":    100000, // 1000 NGN in kobo
		"email":     "test@example.com",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/deposits/init", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Note: This will fail in test environment due to Paystack API, but we can test the structure
	assert.Equal(t, http.StatusInternalServerError, w.Code) // Expected due to test Paystack key

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.Contains(t, response["message"].(string), "Failed to initialize deposit")
}

func TestPaystackWebhook(t *testing.T) {
	router := setupTestServer(t)
	db := setupTestDB(t)

	// Create test user and wallet
	user := &models.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "hashed_password",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	wallet := &models.Wallet{
		UserID:   user.ID,
		Currency: "NGN",
		Balance:  0,
		Status:   "active",
	}
	err = db.Create(wallet).Error
	require.NoError(t, err)

	// Update user with wallet ID
	user.WalletID = wallet.ID
	err = db.Save(user).Error
	require.NoError(t, err)

	// Create a payment record
	payment := &models.Payment{
		UserID:          user.ID,
		WalletID:        wallet.ID,
		Amount:          100000, // 1000 NGN in kobo
		Currency:        "NGN",
		Type:            "deposit",
		Status:          "pending",
		PaystackRef:     "test_ref_123",
		PaystackTransID: "123456",
		Description:     "Test deposit",
	}
	err = db.Create(payment).Error
	require.NoError(t, err)

	// Webhook payload
	webhookData := map[string]interface{}{
		"event":     "charge.success",
		"reference": "test_ref_123",
		"status":    "success",
		"amount":    100000,
	}

	jsonBody, _ := json.Marshal(webhookData)
	req, _ := http.NewRequest("POST", "/api/v1/webhooks/paystack", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.Equal(t, "Webhook processed successfully", response["message"])
}
