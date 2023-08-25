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
	"github.com/stripe/stripe-go/v75/paymentintent"
	"github.com/stripe/stripe-go/v75/webhook"
)

func (s WebServer) newPaymentIntent(w http.ResponseWriter, req *http.Request) {
	userInfo, err := s.getUserInfo(req)
	if err != nil {
		s.logger.Errorf("getting user info: %s", err)
		http.Error(w, "could not get user", http.StatusBadRequest)
		return
	}

	if userInfo.Provenance == "heroku" {
		s.logger.Errorf("heroku user cannot create payment intent")
		http.Error(w, `{"error":"heroku user cannot create payment intent"}`, http.StatusBadRequest)
		return
	}

	type instanceRequest struct {
		Name string `json:"name"`
		Plan string `json:"plan"`
	}
	var ir instanceRequest
	err = json.NewDecoder(req.Body).Decode(&ir)
	if err != nil {
		http.Error(w, `{"error":"parsing request"}`, http.StatusBadRequest)
		return
	}

	if ir.Name == "" || ir.Plan == "" {
		http.Error(w, `{"error":"name and plan are required"}`, http.StatusBadRequest)
		return
	}

	pricePennies := account.LookupPricingPlan(ir.Plan).PriceDollars * 100
	stripe.Key = s.stripeKey
	params := &stripe.PaymentIntentParams{
		Amount: stripe.Int64(int64(pricePennies)),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		Customer: stripe.String(userInfo.StripeID),
		Metadata: map[string]string{
			"plan": ir.Plan,
			"name": ir.Name,
		},
	}
	pi, err := paymentintent.New(params)
	if err != nil {
		s.logger.Errorf("creating payment intent %s", err.Error())
		http.Error(w, `{"error":"error creating payment intent"}`, http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, `{"status":"success","clientSecret":"%s"}`, pi.ClientSecret)
}

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

	signatureHeader := req.Header.Get("Stripe-Signature")
	event, err = webhook.ConstructEventWithOptions(payload, signatureHeader, s.stripeWebhookSigningSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true, // TODO: fix this
	})
	if err != nil {
		s.logger.Error("webhook signature verification failed: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

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
