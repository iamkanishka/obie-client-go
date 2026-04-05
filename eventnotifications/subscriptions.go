// Package eventnotifications implements the OBIE Event Notification API
// endpoints defined in v3.1.2:
//
//   - Event Subscriptions  (POST/GET/PUT/DELETE /event-subscriptions)
//   - Callback URLs        (POST/GET/PUT/DELETE /callback-urls)
//   - Aggregated Polling   (POST /events)
//
// Ref: https://openbankinguk.github.io/read-write-api-site2/standards/v3.1.3/resources-and-data-models/event-notifications/
package eventnotifications

import (
	"context"
	"fmt"

	"github.com/iamkanishka/obie-client-go/internal/transport"
	"github.com/iamkanishka/obie-client-go/models"
)

const basePath = "/open-banking/v3.1"

// Service implements all OBIE Event Notification endpoints.
type Service struct {
	http    transport.HTTPDoer
	signer  jwsSigner
	baseURL string
}

// jwsSigner signs request payloads.
type jwsSigner interface {
	SignJSON(v any) (string, error)
}

// New creates an event notification Service.
func New(h transport.HTTPDoer, signer jwsSigner, baseURL string) *Service {
	return &Service{http: h, signer: signer, baseURL: baseURL}
}

// ── Event Subscriptions ───────────────────────────────────────────────────

// CreateEventSubscription registers a new event subscription with the ASPSP.
//
// The TPP provides a callback URL and/or specific event types to subscribe to.
// The ASPSP responds with an EventSubscriptionId.
//
// POST /event-subscriptions
func (s *Service) CreateEventSubscription(
	ctx context.Context,
	req *models.OBEventSubscription1,
) (*models.OBEventSubscriptionResponse1, error) {
	sig, err := s.signer.SignJSON(req)
	if err != nil {
		return nil, fmt.Errorf("eventnotifications: sign subscription: %w", err)
	}
	opts := transport.DoOptions{JWSSignature: sig}
	var resp models.OBEventSubscriptionResponse1
	if err := s.http.Post(ctx,
		s.baseURL+basePath+"/event-subscriptions",
		req, &resp, opts); err != nil {
		return nil, fmt.Errorf("eventnotifications: CreateEventSubscription: %w", err)
	}
	return &resp, nil
}

// GetEventSubscriptions retrieves all event subscriptions for the TPP.
//
// GET /event-subscriptions
func (s *Service) GetEventSubscriptions(ctx context.Context) (*models.OBEventSubscriptionsResponse1, error) {
	var resp models.OBEventSubscriptionsResponse1
	if err := s.http.Get(ctx, s.baseURL+basePath+"/event-subscriptions", &resp); err != nil {
		return nil, fmt.Errorf("eventnotifications: GetEventSubscriptions: %w", err)
	}
	return &resp, nil
}

// UpdateEventSubscription replaces an existing event subscription.
//
// PUT /event-subscriptions/{EventSubscriptionId}
func (s *Service) UpdateEventSubscription(
	ctx context.Context,
	eventSubscriptionID string,
	req *models.OBEventSubscriptionResponse1,
) (*models.OBEventSubscriptionResponse1, error) {
	sig, err := s.signer.SignJSON(req)
	if err != nil {
		return nil, fmt.Errorf("eventnotifications: sign update subscription: %w", err)
	}
	opts := transport.DoOptions{JWSSignature: sig}
	var resp models.OBEventSubscriptionResponse1
	if err := s.http.Put(ctx,
		fmt.Sprintf("%s%s/event-subscriptions/%s", s.baseURL, basePath, eventSubscriptionID),
		req, &resp, opts); err != nil {
		return nil, fmt.Errorf("eventnotifications: UpdateEventSubscription(%s): %w", eventSubscriptionID, err)
	}
	return &resp, nil
}

// DeleteEventSubscription removes an event subscription.
//
// DELETE /event-subscriptions/{EventSubscriptionId}
func (s *Service) DeleteEventSubscription(ctx context.Context, eventSubscriptionID string) error {
	if err := s.http.Delete(ctx,
		fmt.Sprintf("%s%s/event-subscriptions/%s", s.baseURL, basePath, eventSubscriptionID)); err != nil {
		return fmt.Errorf("eventnotifications: DeleteEventSubscription(%s): %w", eventSubscriptionID, err)
	}
	return nil
}

// ── Callback URLs ─────────────────────────────────────────────────────────

// CreateCallbackUrl registers a callback URL with the ASPSP.
//
// POST /callback-urls
func (s *Service) CreateCallbackUrl(
	ctx context.Context,
	req *models.OBCallbackUrl1,
) (*models.OBCallbackUrlResponse1, error) {
	sig, err := s.signer.SignJSON(req)
	if err != nil {
		return nil, fmt.Errorf("eventnotifications: sign callback url: %w", err)
	}
	opts := transport.DoOptions{JWSSignature: sig}
	var resp models.OBCallbackUrlResponse1
	if err := s.http.Post(ctx,
		s.baseURL+basePath+"/callback-urls",
		req, &resp, opts); err != nil {
		return nil, fmt.Errorf("eventnotifications: CreateCallbackUrl: %w", err)
	}
	return &resp, nil
}

// GetCallbackUrls retrieves all callback URLs registered by the TPP.
//
// GET /callback-urls
func (s *Service) GetCallbackUrls(ctx context.Context) (*models.OBCallbackUrlsResponse1, error) {
	var resp models.OBCallbackUrlsResponse1
	if err := s.http.Get(ctx, s.baseURL+basePath+"/callback-urls", &resp); err != nil {
		return nil, fmt.Errorf("eventnotifications: GetCallbackUrls: %w", err)
	}
	return &resp, nil
}

// UpdateCallbackUrl replaces an existing callback URL registration.
//
// PUT /callback-urls/{CallbackUrlId}
func (s *Service) UpdateCallbackUrl(
	ctx context.Context,
	callbackUrlID string,
	req *models.OBCallbackUrl1,
) (*models.OBCallbackUrlResponse1, error) {
	sig, err := s.signer.SignJSON(req)
	if err != nil {
		return nil, fmt.Errorf("eventnotifications: sign update callback url: %w", err)
	}
	opts := transport.DoOptions{JWSSignature: sig}
	var resp models.OBCallbackUrlResponse1
	if err := s.http.Put(ctx,
		fmt.Sprintf("%s%s/callback-urls/%s", s.baseURL, basePath, callbackUrlID),
		req, &resp, opts); err != nil {
		return nil, fmt.Errorf("eventnotifications: UpdateCallbackUrl(%s): %w", callbackUrlID, err)
	}
	return &resp, nil
}

// DeleteCallbackUrl removes a registered callback URL.
//
// DELETE /callback-urls/{CallbackUrlId}
func (s *Service) DeleteCallbackUrl(ctx context.Context, callbackUrlID string) error {
	if err := s.http.Delete(ctx,
		fmt.Sprintf("%s%s/callback-urls/%s", s.baseURL, basePath, callbackUrlID)); err != nil {
		return fmt.Errorf("eventnotifications: DeleteCallbackUrl(%s): %w", callbackUrlID, err)
	}
	return nil
}

// ── Aggregated Event Polling ──────────────────────────────────────────────

// PollEvents fetches and acknowledges event notifications using the
// aggregated polling approach.
//
// The TPP sends a list of JTIs to acknowledge and optionally limits the
// number of events returned.
//
// POST /events
func (s *Service) PollEvents(
	ctx context.Context,
	req *models.OBEventPolling1,
) (*models.OBEventPollingResponse1, error) {
	var resp models.OBEventPollingResponse1
	if err := s.http.Post(ctx,
		s.baseURL+basePath+"/events",
		req, &resp, transport.DoOptions{}); err != nil {
		return nil, fmt.Errorf("eventnotifications: PollEvents: %w", err)
	}
	return &resp, nil
}
