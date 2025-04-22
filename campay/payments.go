package campay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/golang-jwt/jwt"
)

//go:generate mockgen -source ./payments.go -destination pymntsmocks/payments.mock.go -package pymntsmocks

type PaymentService interface {
	InitiateCampayMobileMoneyPayments(ctx context.Context, req CampayPaymentsRequest) (*CampayPaymentsResponse, error)
	Withdraw(ctx context.Context, req WithdrawalRequest) (*WithdrawalResponse, error)
	VerifyCampayWebHookSignature(ctx context.Context, signature string, campayWebHookSecretKey string) error
}

// CampayAccessToken represents the response body from campay.
type CampayAccessToken struct {
	AccessToken string `json:"token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// PymentServiceImpl service implementation
type PymentServiceImpl struct {
	CampayAppUserName string
	CampayBaseURL     string
	CampayAppPassword string
}

var _ PaymentService = &PymentServiceImpl{}

// NewPaymentClient initialize a new campay payment instance.
// use https://www.demo.campay.com/api  Base url for sandbox and https://www.campay.net/api for production (when app goes live).
func NewPaymentClient(campayAppUserName string, campayAppPassword string, campayBaseURL string) (*PymentServiceImpl, error) {
	return &PymentServiceImpl{
		CampayAppUserName: campayAppUserName,
		CampayBaseURL:     campayBaseURL,
		CampayAppPassword: campayAppPassword,
	}, nil
}

// inititates the payments to campay. don't forget the required @amount,@phone and @from fields args.
func (p *PymentServiceImpl) InitiateCampayMobileMoneyPayments(ctx context.Context, req CampayPaymentsRequest) (*CampayPaymentsResponse, error) {
	client := &http.Client{}

	accessToken, err := p.GetCampayAccessToken(ctx, client)
	if err != nil {
		return nil, err
	}

	paymentRequests, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf(`failed to parse the payment requests. why: %w `, err)
	}

	paymentsRequestsWithContext, err := http.NewRequestWithContext(ctx, http.MethodPost, p.CampayBaseURL+"/collect/", bytes.NewReader(paymentRequests))
	if err != nil {
		return nil, fmt.Errorf(`failed to configure requests to initiate payments to campay. why: %w`, err)
	}

	paymentsRequestsWithContext.Header.Add("Authorization", "Token "+accessToken.AccessToken)
	paymentsRequestsWithContext.Header.Add("Content-Type", "application/json")

	paymentResponse, err := client.Do(paymentsRequestsWithContext)
	if err != nil {
		return nil, fmt.Errorf(`fail to initiate payment requests to campay. why: %w`, err)
	}

	defer paymentResponse.Body.Close()

	if paymentResponse.StatusCode == http.StatusBadRequest {
		body, _ := io.ReadAll(paymentResponse.Body)
		return nil, fmt.Errorf(`failed response from campay: status_code=%d response_body=%s`, paymentResponse.StatusCode, string(body))
	}

	pymntBody, err := io.ReadAll(paymentResponse.Body)
	if err != nil {
		return nil, fmt.Errorf(`failed to read payment response body from campay. why: %s`, err)
	}

	var pymntResponse CampayPaymentsResponse

	if err := json.Unmarshal(pymntBody, &pymntResponse); err != nil {
		return nil, fmt.Errorf(`error umarshaling payment response body. why: %w`, err)
	}

	return &pymntResponse, nil
}

// GetCampayAccessToken retrieves the access token associated with app name and password.
func (p *PymentServiceImpl) GetCampayAccessToken(ctx context.Context, client *http.Client) (*CampayAccessToken, error) {
	campayAppCredentials := map[string]string{
		"username": p.CampayAppUserName,
		"password": p.CampayAppPassword,
	}

	tokenBody, err := json.Marshal(campayAppCredentials)
	if err != nil {
		return nil, fmt.Errorf(`failed to serialized token user credentials. why: %w`, err)
	}

	tokenRequestWithContext, err := http.NewRequestWithContext(ctx, http.MethodPost, p.CampayBaseURL+"/token/", bytes.NewReader(tokenBody))
	if err != nil {
		return nil, fmt.Errorf(`failed to configure request request to campay. why: %w`, err)
	}

	tokenRequestWithContext.Header.Add("Content-Type", "application/json")

	response, err := client.Do(tokenRequestWithContext)
	if err != nil {
		return nil, fmt.Errorf(`failed to requests token from campay. why: %w`, err)
	}

	defer response.Body.Close()

	if response.StatusCode == http.StatusBadRequest {
		body, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("access token request failed: status_code=%v response_body=%s", response.StatusCode, string(body))
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf(`failed to read token response from campay. why: %w`, err)
	}

	var campayTokenResponse CampayAccessToken

	if err := json.Unmarshal(body, &campayTokenResponse); err != nil {
		return nil, fmt.Errorf(`error unmarshaling access token response body. why: %w`, err)
	}

	return &campayTokenResponse, nil
}

// VerifyCampayWebHookSignature authenticates the webhook signature to make sure the response was signed by campay.
func (p *PymentServiceImpl) VerifyCampayWebHookSignature(ctx context.Context, signature string, campayWebHookSecretKey string) error {
	token, err := jwt.Parse(signature, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			alg := token.Header["alg"]
			return nil, fmt.Errorf("unexpected signing algorithm: alg=%v %s", alg, "UNEXPECTED_SIGNING_ALG")
		}
		return []byte(campayWebHookSecretKey), nil
	})

	if !token.Valid {
		return err
	}

	return nil
}

// inititates a withdrawal to the provided number. don't forget the required @amount, @to fields args.
func (p *PymentServiceImpl) Withdraw(ctx context.Context, req WithdrawalRequest) (*WithdrawalResponse, error) {
	client := &http.Client{}

	accessToken, err := p.GetCampayAccessToken(ctx, client)
	if err != nil {
		return nil, err
	}

	paymentRequests, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf(`failed to parse the payment requests. why: %w `, err)
	}

	paymentsRequestsWithContext, err := http.NewRequestWithContext(ctx, http.MethodPost, p.CampayBaseURL+"/withdraw/", bytes.NewReader(paymentRequests))
	if err != nil {
		return nil, fmt.Errorf(`failed to configure requests to initiate withdrawal. why: %w`, err)
	}

	paymentsRequestsWithContext.Header.Add("Authorization", "Token "+accessToken.AccessToken)
	paymentsRequestsWithContext.Header.Add("Content-Type", "application/json")

	paymentResponse, err := client.Do(paymentsRequestsWithContext)
	if err != nil {
		return nil, fmt.Errorf(`fail to initiate payment requests to campay. why: %w`, err)
	}

	defer paymentResponse.Body.Close()

	if paymentResponse.StatusCode == http.StatusBadRequest {
		body, _ := io.ReadAll(paymentResponse.Body)
		return nil, fmt.Errorf(`failed response from campay: status_code=%d response_body=%s`, paymentResponse.StatusCode, string(body))
	}

	pymntBody, err := io.ReadAll(paymentResponse.Body)
	if err != nil {
		return nil, fmt.Errorf(`failed to read payment response body from campay. why: %s`, err)
	}

	var pymntResponse WithdrawalResponse

	if err := json.Unmarshal(pymntBody, &pymntResponse); err != nil {
		return nil, fmt.Errorf(`error umarshaling payment response body. why: %w`, err)
	}

	return &pymntResponse, nil
}
