package models

import "time"

// ────────────────────────────────────────────────────────────────────────────
// Event Subscription
// Ref: /resources-and-data-models/event-notifications/event-subscription/
// ────────────────────────────────────────────────────────────────────────────

// OBEventSubscription1 is the request body for POST /event-subscriptions.
type OBEventSubscription1 struct {
	Data OBEventSubscriptionData1 `json:"Data"`
}

// OBEventSubscriptionData1 carries subscription parameters.
type OBEventSubscriptionData1 struct {
	// CallbackUrl is the TPP's endpoint for push notifications.
	// Omit to use polling only.
	CallbackUrl string                  `json:"CallbackUrl,omitempty"`
	// Version must match the API version (e.g. "3.1").
	Version     string                  `json:"Version"`
	// EventTypes: list of specific event types to subscribe to.
	// Omit to subscribe to all events supported by the ASPSP.
	EventTypes  []EventNotificationType `json:"EventTypes,omitempty"`
}

// OBEventSubscriptionResponse1 is the response for POST/PUT /event-subscriptions/{id}.
type OBEventSubscriptionResponse1 struct {
	Data  OBEventSubscriptionResponseData1 `json:"Data"`
	Links Links                            `json:"Links"`
	Meta  Meta                             `json:"Meta"`
}

// OBEventSubscriptionResponseData1 extends the request data with the ASPSP-assigned ID.
type OBEventSubscriptionResponseData1 struct {
	EventSubscriptionId string                  `json:"EventSubscriptionId"`
	CallbackUrl         string                  `json:"CallbackUrl,omitempty"`
	Version             string                  `json:"Version"`
	EventTypes          []EventNotificationType `json:"EventTypes,omitempty"`
}

// OBEventSubscriptionsResponse1 is the response for GET /event-subscriptions.
type OBEventSubscriptionsResponse1 struct {
	Data  OBEventSubscriptionsResponseData1 `json:"Data"`
	Links Links                             `json:"Links"`
	Meta  Meta                              `json:"Meta"`
}

// OBEventSubscriptionsResponseData1 wraps all subscriptions for a TPP.
type OBEventSubscriptionsResponseData1 struct {
	EventSubscription []OBEventSubscriptionResponseData1 `json:"EventSubscription"`
}

// ────────────────────────────────────────────────────────────────────────────
// Callback URL
// Ref: /resources-and-data-models/event-notifications/callback-url/
// ────────────────────────────────────────────────────────────────────────────

// OBCallbackUrl1 is the request body for POST /callback-urls.
type OBCallbackUrl1 struct {
	Data OBCallbackUrlData1 `json:"Data"`
}

// OBCallbackUrlData1 carries the callback URL and API version.
type OBCallbackUrlData1 struct {
	Url     string `json:"Url"`
	Version string `json:"Version"`
}

// OBCallbackUrlResponse1 is the response for POST/PUT /callback-urls/{id}.
type OBCallbackUrlResponse1 struct {
	Data  OBCallbackUrlResponseData1 `json:"Data"`
	Links Links                      `json:"Links"`
	Meta  Meta                       `json:"Meta"`
}

// OBCallbackUrlResponseData1 extends the request with the ASPSP-assigned ID.
type OBCallbackUrlResponseData1 struct {
	CallbackUrlId string `json:"CallbackUrlId"`
	Url           string `json:"Url"`
	Version       string `json:"Version"`
}

// OBCallbackUrlsResponse1 is the response for GET /callback-urls.
type OBCallbackUrlsResponse1 struct {
	Data  OBCallbackUrlsResponseData1 `json:"Data"`
	Links Links                       `json:"Links"`
	Meta  Meta                        `json:"Meta"`
}

// OBCallbackUrlsResponseData1 wraps all callback URLs registered by the TPP.
type OBCallbackUrlsResponseData1 struct {
	CallbackUrl []OBCallbackUrlResponseData1 `json:"CallbackUrl"`
}

// ────────────────────────────────────────────────────────────────────────────
// Aggregated Polling
// Ref: /resources-and-data-models/event-notifications/events/
// ────────────────────────────────────────────────────────────────────────────

// OBEventPolling1 is the request body for POST /events (aggregated polling).
// Sets: key = JTI of previously received event to acknowledge; value = the event JWT.
// Ack:  list of JTIs to acknowledge without error.
// SetErrs: key = JTI; value = error details for events that could not be processed.
type OBEventPolling1 struct {
	// Sets contains previously received event JWTs to acknowledge.
	// Key = JTI, Value = the original event JWT string.
	Sets              map[string]string                     `json:"sets,omitempty"`
	// MaxEvents limits the number of new events returned (0 = no limit).
	MaxEvents         *int                                  `json:"maxEvents,omitempty"`
	// ReturnImmediately: true = return empty if no events; false = long-poll.
	ReturnImmediately *bool                                 `json:"returnImmediately,omitempty"`
	// Ack contains JTIs of events to acknowledge (no error).
	Ack               []string                              `json:"ack,omitempty"`
	// SetErrs contains per-event processing errors.
	SetErrs           map[string]OBEventPollingError1       `json:"setErrs,omitempty"`
}

// OBEventPollingError1 describes an error processing a specific event.
type OBEventPollingError1 struct {
	// Err is the SET error code (e.g. "authentication_failed").
	Err         string `json:"err"`
	Description string `json:"description"`
}

// OBEventPollingResponse1 is the response from POST /events.
// Sets: key = JTI, value = event JWT for each new event.
// MoreAvailable: true if there are more events to retrieve.
type OBEventPollingResponse1 struct {
	Sets          map[string]string `json:"sets,omitempty"`
	MoreAvailable bool              `json:"moreAvailable"`
}

// ────────────────────────────────────────────────────────────────────────────
// Real-Time Event Notification
// Ref: /resources-and-data-models/event-notifications/event-notifications/
// ────────────────────────────────────────────────────────────────────────────

// OBEventNotification1 is the JWT payload delivered to a TPP's callback URL.
// The full payload is signed as a JWS — validate the signature before trusting it.
type OBEventNotification1 struct {
	// Iss: issuer — the ASPSP's base URL.
	Iss string `json:"iss"`
	// Iat: time the event was issued (Unix epoch seconds).
	Iat int64  `json:"iat"`
	// Jti: unique identifier for this notification (used as idempotency key for ack).
	Jti string `json:"jti"`
	// Aud: audience — the TPP's client_id.
	Aud string `json:"aud"`
	// Sub: subject — the URL of the resource that triggered the event.
	Sub string `json:"sub"`
	// Txn: transaction identifier.
	Txn string `json:"txn"`
	// Toe: time of event (Unix epoch seconds).
	Toe int64  `json:"toe"`
	// Events: map of event type URN → event payload.
	Events map[EventNotificationType]OBEvent1 `json:"events"`
}

// ToeTime converts the Toe Unix epoch to a time.Time.
func (n *OBEventNotification1) ToeTime() time.Time {
	return time.Unix(n.Toe, 0)
}

// IatTime converts the Iat Unix epoch to a time.Time.
func (n *OBEventNotification1) IatTime() time.Time {
	return time.Unix(n.Iat, 0)
}

// OBEvent1 is the generic event payload inside an event notification.
type OBEvent1 struct {
	Subject OBEventSubject1 `json:"subject"`
}

// OBEventSubject1 describes the resource that triggered the event.
type OBEventSubject1 struct {
	// SubjectType combines the rid and rty using the format:
	// "http://openbanking.org.uk/rid_<rid> http://openbanking.org.uk/rty_<rty>"
	SubjectType    string            `json:"subject_type"`
	HTTPStatusCode int               `json:"http_status_code,omitempty"`
	// Links maps link relation to URL, e.g. "http://openbanking.org.uk/rid" → payment URL.
	Links          map[string]string `json:"links,omitempty"`
	// Version is the API version that emitted this event.
	Version        string            `json:"version,omitempty"`
}
