package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/andrewmarklloyd/heroku-addon/internal/pkg/account"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/webhook"
)

func (s WebServer) handleStripeWebhook(w http.ResponseWriter, req *http.Request) {
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
	payload, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	event := stripe.Event{}

	if err := json.Unmarshal(payload, &event); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Webhook error while parsing basic request. %v\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Replace this endpoint secret with your endpoint's unique secret
	// If you are testing with the CLI, find the secret by running 'stripe listen'
	// If you are using an endpoint defined with the API or dashboard, look in your webhook settings
	// at https://dashboard.stripe.com/webhooks
	endpointSecret := ""
	signatureHeader := req.Header.Get("Stripe-Signature")
	event, err = webhook.ConstructEventWithOptions(payload, signatureHeader, endpointSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true, // TODO: fix this
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Webhook signature verification failed. %v\n", err)
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		return
	}
	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case "charge.succeeded":
		s.logger.Info("charge.succeeded event received")
		var charge stripe.Charge
		err := json.Unmarshal(event.Data.Raw, &charge)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = s.handleChargeSucceeded(charge)
		if err != nil {
			s.logger.Errorf("handling charge succeeded: %s", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

	default:
		// fmt.Fprintf(os.Stderr, "Unhandled event type: %s\n", event.Type)
	}

	w.WriteHeader(http.StatusOK)
}

func (s WebServer) handleChargeSucceeded(charge stripe.Charge) error {
	a, err := s.postgresClient.GetAccountFromStripeCustID(s.cryptoUtil, charge.Customer.ID)
	if err != nil {
		return fmt.Errorf("getting account from stripe customer id: %w", err)
	}

	instanceName, ok := charge.Metadata["name"]
	if !ok {
		return fmt.Errorf("name key in charge metadata not found")
	}

	instancePlan, ok := charge.Metadata["plan"]
	if !ok {
		return fmt.Errorf("plan key in charge metadata not found")
	}

	i := account.Instance{
		AccountID: a.UUID,
		Id:        uuid.New().String(),
		Plan:      instancePlan,
		Name:      instanceName,
	}

	err = s.postgresClient.CreateOrUpdateInstance(i)
	if err != nil {
		s.logger.Errorf("creating instance: %s", err)
		return fmt.Errorf("creating instance: %w", err)
	}

	return nil
}
