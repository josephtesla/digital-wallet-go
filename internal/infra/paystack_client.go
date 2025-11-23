package infra

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type PaystackClient struct {
	SecretKey string
	BaseURL   string
	Client    *http.Client
}

type PaystackResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type InitializeTransactionRequest struct {
	Amount      int64  `json:"amount"`
	Email       string `json:"email"`
	Reference   string `json:"reference"`
	CallbackURL string `json:"callback_url"`
}

type InitializeTransactionResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

type VerifyTransactionResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID              int    `json:"id"`
		Status          string `json:"status"`
		Reference       string `json:"reference"`
		Amount          int64  `json:"amount"`
		Currency        string `json:"currency"`
		Channel         string `json:"channel"`
		GatewayResponse string `json:"gateway_response"`
		PaidAt          string `json:"paid_at"`
		CreatedAt       string `json:"created_at"`
	} `json:"data"`
}

type CreateTransferRecipientRequest struct {
	Type          string `json:"type"`
	Name          string `json:"name"`
	AccountNumber string `json:"account_number"`
	BankCode      string `json:"bank_code"`
	Currency      string `json:"currency"`
}

type CreateTransferRecipientResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Active        bool   `json:"active"`
		CreatedAt     string `json:"created_at"`
		Currency      string `json:"currency"`
		Domain        string `json:"domain"`
		Email         string `json:"email"`
		ID            int    `json:"id"`
		Integration   int    `json:"integration"`
		Name          string `json:"name"`
		RecipientCode string `json:"recipient_code"`
		Type          string `json:"type"`
		UpdatedAt     string `json:"updated_at"`
	} `json:"data"`
}

type InitiateTransferRequest struct {
	Source    string `json:"source"`
	Amount    int64  `json:"amount"`
	Recipient string `json:"recipient"`
	Reason    string `json:"reason"`
	Reference string `json:"reference"`
}

type InitiateTransferResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Integration  int    `json:"integration"`
		Domain       string `json:"domain"`
		Amount       int64  `json:"amount"`
		Currency     string `json:"currency"`
		Source       string `json:"source"`
		Reason       string `json:"reason"`
		Recipient    int    `json:"recipient"`
		Status       string `json:"status"`
		TransferCode string `json:"transfer_code"`
		ID           int    `json:"id"`
		CreatedAt    string `json:"created_at"`
		UpdatedAt    string `json:"updated_at"`
	} `json:"data"`
}

func InitPaystackClient(secretKey string) *PaystackClient {
	return &PaystackClient{
		SecretKey: secretKey,
		BaseURL:   "https://api.paystack.co",
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *PaystackClient) makeRequest(method, endpoint string, body interface{}) (*PaystackResponse, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, p.BaseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.SecretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var paystackResp PaystackResponse
	if err := json.Unmarshal(respBody, &paystackResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &paystackResp, nil
}

func (p *PaystackClient) InitializeTransaction(req *InitializeTransactionRequest) (*InitializeTransactionResponse, error) {
	resp, err := p.makeRequest("POST", "/transaction/initialize", req)
	if err != nil {
		return nil, err
	}

	var initResp InitializeTransactionResponse
	respBytes, _ := json.Marshal(resp)
	if err := json.Unmarshal(respBytes, &initResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal initialize response: %w", err)
	}

	return &initResp, nil
}

func (p *PaystackClient) VerifyTransaction(reference string) (*VerifyTransactionResponse, error) {
	resp, err := p.makeRequest("GET", "/transaction/verify/"+reference, nil)
	if err != nil {
		return nil, err
	}

	var verifyResp VerifyTransactionResponse
	respBytes, _ := json.Marshal(resp)
	if err := json.Unmarshal(respBytes, &verifyResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal verify response: %w", err)
	}

	return &verifyResp, nil
}

func (p *PaystackClient) CreateTransferRecipient(req *CreateTransferRecipientRequest) (*CreateTransferRecipientResponse, error) {
	resp, err := p.makeRequest("POST", "/transferrecipient", req)
	if err != nil {
		return nil, err
	}

	var recipientResp CreateTransferRecipientResponse
	respBytes, _ := json.Marshal(resp)
	if err := json.Unmarshal(respBytes, &recipientResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recipient response: %w", err)
	}

	return &recipientResp, nil
}

func (p *PaystackClient) InitiateTransfer(req *InitiateTransferRequest) (*InitiateTransferResponse, error) {
	resp, err := p.makeRequest("POST", "/transfer", req)
	if err != nil {
		return nil, err
	}

	var transferResp InitiateTransferResponse
	respBytes, _ := json.Marshal(resp)
	if err := json.Unmarshal(respBytes, &transferResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transfer response: %w", err)
	}

	return &transferResp, nil
}
