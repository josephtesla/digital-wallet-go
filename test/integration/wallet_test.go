package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/josephtesla/digital-wallet-go/internal/api"
	"github.com/josephtesla/digital-wallet-go/internal/infra"
	"github.com/josephtesla/digital-wallet-go/internal/models"
	"github.com/josephtesla/digital-wallet-go/internal/repository"
	"github.com/josephtesla/digital-wallet-go/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations
	err = db.AutoMigrate(
		&models.User{},
		&models.Wallet{},
		&models.Payment{},
		&models.LedgerTransaction{},
		&models.LedgerEntry{},
		&models.BankAccount{},
		&models.Idempotency{},
		&models.SystemAccount{},
	)
	require.NoError(t, err)

	return db
}

func setupTestRedis(t *testing.T) *redis.Client {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set")
	}

	opt, err := redis.ParseURL(redisURL)
	require.NoError(t, err)

	client := redis.NewClient(opt)

	// Test connection
	ctx := context.Background()
	_, err = client.Ping(ctx).Result()
	require.NoError(t, err)

	return client
}

func setupTestServer(t *testing.T) *gin.Engine {
	db := setupTestDB(t)
	redisClient := setupTestRedis(t)
	logger := infra.InitLogger("debug")
	paystackClient := infra.InitPaystackClient("test_key")

	repos := repository.NewRepositories(db)
	services := service.NewServices(repos, redisClient, paystackClient, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	api.SetupRoutes(router, services, logger)

	return router
}

func TestCreateWallet(t *testing.T) {
	router := setupTestServer(t)

	// Create a test user first
	user := &models.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "hashed_password",
	}

	db := setupTestDB(t)
	err := db.Create(user).Error
	require.NoError(t, err)

	// Create wallet request
	reqBody := map[string]interface{}{
		"user_id":  user.ID.String(),
		"currency": "NGN",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/wallets", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.Equal(t, "Wallet created successfully", response["message"])
}

func TestGetWalletBalance(t *testing.T) {
	router := setupTestServer(t)

	// Create test data
	db := setupTestDB(t)
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
		Balance:  100000, // 1000 NGN in kobo
		Status:   "active",
	}
	err = db.Create(wallet).Error
	require.NoError(t, err)

	// Update user with wallet ID
	user.WalletID = wallet.ID
	err = db.Save(user).Error
	require.NoError(t, err)

	// Test get balance
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/wallets/%s/balance", wallet.ID.String()), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.Equal(t, float64(100000), response["data"].(map[string]interface{})["balance"])
	assert.Equal(t, "NGN", response["data"].(map[string]interface{})["currency"])
}
