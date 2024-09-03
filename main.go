package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Iknite-Space/campay-go-sdk/campay"
	"github.com/ardanlabs/conf/v3"
	"github.com/joho/godotenv"
)

var cfg struct {
	AWS struct {
		CampayAppUserName string `conf:"env:CAMPAY_USERNAME,required"`
		CampayAppPassword string `conf:"env:CAMPAY_PASSWORD,required"`
		CampayBaseURL     string `conf:"env:CAMPAY_BASE_URL,required"`
		CampayWebHookKey  string `conf:"env:WEBHOOK_APP_KEY,required"`
	}
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	// loadDevEnv loads .env file if present
	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	help, err := conf.Parse("", &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Printf("%v\n", help)

			return nil
		}

		return fmt.Errorf("parsing config: %w", err)
	}

	payments, err := campay.NewPaymentClient(cfg.AWS.CampayAppUserName, cfg.AWS.CampayAppPassword, cfg.AWS.CampayBaseURL)
	if err != nil {
		return fmt.Errorf("failed to create campay payment client: %w", err)
	}

	paymentReq := campay.CampayPaymentsRequest{
		Amount:      "5",
		From:        "+237671738755",
		Description: "Payment for subscription",
		ExternalRef: "payment_12345",
	}
	
	ctx := context.Background()

	paymentResponse, err := payments.InitiateCampayMobileMoneyPayments(ctx, paymentReq)
	if err != nil {
		return fmt.Errorf("failed to initiate payment: %w", err)
	}

	fmt.Printf("Payment initiated successfully. Reference: %s\n", paymentResponse.Reference)

	return nil
}
