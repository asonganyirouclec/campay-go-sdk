package campay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Iknite-Space/campay-go-sdk/internal/models"
	"github.com/rs/zerolog"
)

var logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()

//go:generate mockgen -source ./campay.go -destination mocks/campays.mock.go -package mocks

type PaymentService interface {
	InitiatePayments(ctx context.Context, req models.RequestBody) error
}

type PymentServiceImpl struct {
	UserName string
	baseURL  string
	UserPwd  string
}

//nolint:exhaustivestruct
var _ PaymentService = &PymentServiceImpl{}

func NewPayment(user string, pwd string, baseURL string) (*PymentServiceImpl, error) {
	return &PymentServiceImpl{UserName: user, UserPwd: pwd, baseURL: baseURL}, nil
}

// inititates the payments to campay. don't forget the required @amount,@phone and @from.
//
//nolint:funlen
func (p *PymentServiceImpl) InitiatePayments(ctx context.Context, req models.RequestBody) error {
	client := &http.Client{}

	token, err := p.getAcessToken(client)
	if err != nil {
		errMsg := "failed to get Access Token"
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msgf("%s :-> %v", errMsg, err)

		return fmt.Errorf("%s :-> %w", errMsg, err)
	}

	initiateBody, err := json.Marshal(req)
	if err != nil {
		errMsg := "failed to Marhsal initiate pyment Req"
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msgf("%s :-> %v", errMsg, err)

		return fmt.Errorf("%s :-> %w", errMsg, err)
	}
	//nolint:noctx
	pymntsReq, err := http.NewRequest(http.MethodPost, p.baseURL+"/collect/", bytes.NewReader(initiateBody))
	if err != nil {
		errMsg := "initiate pymnt error"
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msgf("%s :-> %v", errMsg, err)

		return fmt.Errorf("%s :-> %w", errMsg, err)
	}

	pymntsReq.Header.Add("Authorization", "Token "+token.AccessToken)
	pymntsReq.Header.Add("Content-Type", "application/json")

	pymntRes, err := client.Do(pymntsReq)
	if err != nil {
		errMsg := "failed to initiate payments"
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msgf("%s :->  %v", errMsg, err)

		return fmt.Errorf("%s :-> %w", errMsg, err)
	}
	defer pymntRes.Body.Close()

	if pymntRes.StatusCode == http.StatusBadRequest {
		body, _ := io.ReadAll(pymntRes.Body)

		errMsg := "bad status code for pymnt"
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msgf("%s :->  %v", errMsg, err)

		//nolint:goerr113
		return fmt.Errorf(" %s :-> %v %s", errMsg, pymntRes.StatusCode, string(body))
	}

	pymntBody, err := io.ReadAll(pymntRes.Body)
	if err != nil {
		errMsg := "failed to read pymnt body"
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msgf("%s :->  %v", errMsg, err)

		return fmt.Errorf("%s :-> %w", errMsg, err)
	}

	var pymntResponse models.ResponseBody

	if err := json.Unmarshal(pymntBody, &pymntResponse); err != nil {
		errMsg := "error umarshaling pyment response body"
		logger.Error().Str("correlationID", fmt.Sprint(ctx.Value("correlationID"))).Msgf("%s :->  %v", errMsg, err)

		return fmt.Errorf("%s :-> %w", errMsg, err)
	}

	return err
}
