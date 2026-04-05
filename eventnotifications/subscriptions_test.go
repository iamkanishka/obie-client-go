package eventnotifications_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iamkanishka/obie-client-go/eventnotifications"
	"github.com/iamkanishka/obie-client-go/internal/transport"
	"github.com/iamkanishka/obie-client-go/models"
	"github.com/iamkanishka/obie-client-go/obie"
)

// ─── test infrastructure ─────────────────────────────────────────────────

type testDoer struct{ client *http.Client }

func (d *testDoer) Get(ctx context.Context, url string, out any) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return &obie.APIError{StatusCode: resp.StatusCode}
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (d *testDoer) Post(ctx context.Context, url string, body, out any, _ transport.DoOptions) error {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(b)))
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return &obie.APIError{StatusCode: resp.StatusCode}
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (d *testDoer) Put(ctx context.Context, url string, body, out any, _ transport.DoOptions) error {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPut, url, strings.NewReader(string(b)))
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return &obie.APIError{StatusCode: resp.StatusCode}
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (d *testDoer) Delete(ctx context.Context, url string) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return &obie.APIError{StatusCode: resp.StatusCode}
	}
	return nil
}

type stubSigner struct{}

func (s *stubSigner) SignJSON(_ any) (string, error) { return "hdr..sig", nil }

func newSvc(t *testing.T, mux *http.ServeMux) (*eventnotifications.Service, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return eventnotifications.New(&testDoer{client: srv.Client()}, &stubSigner{}, srv.URL), srv
}

func jsonH(v any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
  if err := json.NewEncoder(w).Encode(v); err != nil {
  	http.Error(w, err.Error(), http.StatusInternalServerError)
  }
	}
}

// ─── Event Subscription tests ────────────────────────────────────────────

func TestCreateEventSubscription(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/event-subscriptions",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(models.OBEventSubscriptionResponse1{
				Data: models.OBEventSubscriptionResponseData1{
					EventSubscriptionId: "sub-101",
					CallbackUrl:         "https://tpp.example.com/events",
					Version:             "3.1",
				},
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})

	svc, _ := newSvc(t, mux)
	resp, err := svc.CreateEventSubscription(context.Background(), &models.OBEventSubscription1{
		Data: models.OBEventSubscriptionData1{
			CallbackUrl: "https://tpp.example.com/events",
			Version:     "3.1",
		},
	})
	if err != nil {
		t.Fatalf("CreateEventSubscription: %v", err)
	}
	if resp.Data.EventSubscriptionId != "sub-101" {
		t.Errorf("EventSubscriptionId: got %q", resp.Data.EventSubscriptionId)
	}
}

func TestCreateEventSubscription_WithEventTypes(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/event-subscriptions",
		func(w http.ResponseWriter, r *http.Request) {
			var req models.OBEventSubscription1
   if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
   	http.Error(w, err.Error(), http.StatusBadRequest)
   	return
   }
			if len(req.Data.EventTypes) == 0 {
				t.Error("expected EventTypes to be populated")
			}
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(models.OBEventSubscriptionResponse1{
				Data: models.OBEventSubscriptionResponseData1{
					EventSubscriptionId: "sub-102",
					Version:             "3.1",
					EventTypes:          req.Data.EventTypes,
				},
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})

	svc, _ := newSvc(t, mux)
	resp, err := svc.CreateEventSubscription(context.Background(), &models.OBEventSubscription1{
		Data: models.OBEventSubscriptionData1{
			Version: "3.1",
			EventTypes: []models.EventNotificationType{
				models.EventNotificationResourceUpdate,
				models.EventNotificationConsentAuthorizationRevoked,
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateEventSubscription with event types: %v", err)
	}
	if len(resp.Data.EventTypes) != 2 {
		t.Errorf("EventTypes count: got %d, want 2", len(resp.Data.EventTypes))
	}
}

func TestGetEventSubscriptions(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/event-subscriptions",
		jsonH(models.OBEventSubscriptionsResponse1{
			Data: models.OBEventSubscriptionsResponseData1{
				EventSubscription: []models.OBEventSubscriptionResponseData1{
					{EventSubscriptionId: "sub-101", CallbackUrl: "https://tpp.example.com/events", Version: "3.1"},
				},
			},
		}))

	svc, _ := newSvc(t, mux)
	resp, err := svc.GetEventSubscriptions(context.Background())
	if err != nil {
		t.Fatalf("GetEventSubscriptions: %v", err)
	}
	if len(resp.Data.EventSubscription) != 1 {
		t.Errorf("subscription count: got %d, want 1", len(resp.Data.EventSubscription))
	}
	if resp.Data.EventSubscription[0].EventSubscriptionId != "sub-101" {
		t.Errorf("EventSubscriptionId: got %q", resp.Data.EventSubscription[0].EventSubscriptionId)
	}
}

func TestUpdateEventSubscription(t *testing.T) {
	putCalled := false
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/event-subscriptions/sub-101",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				t.Errorf("expected PUT, got %s", r.Method)
			}
			putCalled = true
			if err := json.NewEncoder(w).Encode(models.OBEventSubscriptionResponse1{
				Data: models.OBEventSubscriptionResponseData1{
					EventSubscriptionId: "sub-101",
					CallbackUrl:         "https://tpp.example.com/events/v2",
					Version:             "3.1",
				},
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})

	svc, _ := newSvc(t, mux)
	resp, err := svc.UpdateEventSubscription(context.Background(), "sub-101",
		&models.OBEventSubscriptionResponse1{
			Data: models.OBEventSubscriptionResponseData1{
				EventSubscriptionId: "sub-101",
				CallbackUrl:         "https://tpp.example.com/events/v2",
				Version:             "3.1",
			},
		})
	if err != nil {
		t.Fatalf("UpdateEventSubscription: %v", err)
	}
	if !putCalled {
		t.Error("expected PUT to be called")
	}
	if resp.Data.CallbackUrl != "https://tpp.example.com/events/v2" {
		t.Errorf("CallbackUrl: got %q", resp.Data.CallbackUrl)
	}
}

func TestDeleteEventSubscription(t *testing.T) {
	deleted := false
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/event-subscriptions/sub-101",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Errorf("expected DELETE, got %s", r.Method)
				return
			}
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		})

	svc, _ := newSvc(t, mux)
	if err := svc.DeleteEventSubscription(context.Background(), "sub-101"); err != nil {
		t.Fatalf("DeleteEventSubscription: %v", err)
	}
	if !deleted {
		t.Error("expected DELETE request to be made")
	}
}

// ─── Callback URL tests ───────────────────────────────────────────────────

func TestCreateCallbackUrl(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/callback-urls",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(models.OBCallbackUrlResponse1{
				Data: models.OBCallbackUrlResponseData1{
					CallbackUrlId: "cb-1",
					Url:           "https://tpp.example.com/open-banking/v3.1/event-notifications",
					Version:       "3.1",
				},
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})

	svc, _ := newSvc(t, mux)
	resp, err := svc.CreateCallbackUrl(context.Background(), &models.OBCallbackUrl1{
		Data: models.OBCallbackUrlData1{
			Url:     "https://tpp.example.com/open-banking/v3.1/event-notifications",
			Version: "3.1",
		},
	})
	if err != nil {
		t.Fatalf("CreateCallbackUrl: %v", err)
	}
	if resp.Data.CallbackUrlId != "cb-1" {
		t.Errorf("CallbackUrlId: got %q", resp.Data.CallbackUrlId)
	}
}

func TestGetCallbackUrls(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/callback-urls",
		jsonH(models.OBCallbackUrlsResponse1{
			Data: models.OBCallbackUrlsResponseData1{
				CallbackUrl: []models.OBCallbackUrlResponseData1{
					{CallbackUrlId: "cb-1", Url: "https://tpp.example.com/events", Version: "3.1"},
					{CallbackUrlId: "cb-2", Url: "https://tpp.example.com/events/v2", Version: "3.1"},
				},
			},
		}))

	svc, _ := newSvc(t, mux)
	resp, err := svc.GetCallbackUrls(context.Background())
	if err != nil {
		t.Fatalf("GetCallbackUrls: %v", err)
	}
	if len(resp.Data.CallbackUrl) != 2 {
		t.Errorf("callback count: got %d, want 2", len(resp.Data.CallbackUrl))
	}
}

func TestUpdateCallbackUrl(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/callback-urls/cb-1",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				t.Errorf("expected PUT, got %s", r.Method)
			}
			if err := json.NewEncoder(w).Encode(models.OBCallbackUrlResponse1{
				Data: models.OBCallbackUrlResponseData1{
					CallbackUrlId: "cb-1",
					Url:           "https://tpp.example.com/events/updated",
					Version:       "3.1",
				},
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})

	svc, _ := newSvc(t, mux)
	resp, err := svc.UpdateCallbackUrl(context.Background(), "cb-1", &models.OBCallbackUrl1{
		Data: models.OBCallbackUrlData1{
			Url:     "https://tpp.example.com/events/updated",
			Version: "3.1",
		},
	})
	if err != nil {
		t.Fatalf("UpdateCallbackUrl: %v", err)
	}
	if resp.Data.Url != "https://tpp.example.com/events/updated" {
		t.Errorf("Url: got %q", resp.Data.Url)
	}
}

func TestDeleteCallbackUrl(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/callback-urls/cb-1",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Errorf("expected DELETE, got %s", r.Method)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})

	svc, _ := newSvc(t, mux)
	if err := svc.DeleteCallbackUrl(context.Background(), "cb-1"); err != nil {
		t.Fatalf("DeleteCallbackUrl: %v", err)
	}
}

// ─── Aggregated Polling tests ─────────────────────────────────────────────

func TestPollEvents_Empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/events",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if err := json.NewEncoder(w).Encode(models.OBEventPollingResponse1{
				MoreAvailable: false,
				Sets:          map[string]string{},
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})

	svc, _ := newSvc(t, mux)
	maxEvents := 5
	resp, err := svc.PollEvents(context.Background(), &models.OBEventPolling1{
		MaxEvents: &maxEvents,
	})
	if err != nil {
		t.Fatalf("PollEvents: %v", err)
	}
	if resp.MoreAvailable {
		t.Error("expected MoreAvailable=false")
	}
}

func TestPollEvents_WithEvents(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/events",
		func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewEncoder(w).Encode(models.OBEventPollingResponse1{
				MoreAvailable: true,
				Sets: map[string]string{
					"jti-001": "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJwYXktMTIzIn0.sig",
					"jti-002": "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJwYXktNDU2In0.sig",
				},
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})

	svc, _ := newSvc(t, mux)
	resp, err := svc.PollEvents(context.Background(), &models.OBEventPolling1{
		Ack: []string{},
	})
	if err != nil {
		t.Fatalf("PollEvents with events: %v", err)
	}
	if !resp.MoreAvailable {
		t.Error("expected MoreAvailable=true")
	}
	if len(resp.Sets) != 2 {
		t.Errorf("sets count: got %d, want 2", len(resp.Sets))
	}
}

func TestPollEvents_Acknowledge(t *testing.T) {
	var receivedAck []string
	mux := http.NewServeMux()
	mux.HandleFunc("/open-banking/v3.1/events",
		func(w http.ResponseWriter, r *http.Request) {
			var req models.OBEventPolling1
   if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
   	http.Error(w, err.Error(), http.StatusBadRequest)
   	return
   }
			receivedAck = req.Ack
			if err := json.NewEncoder(w).Encode(models.OBEventPollingResponse1{
				MoreAvailable: false,
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})

	svc, _ := newSvc(t, mux)
	svc.PollEvents(context.Background(), &models.OBEventPolling1{ //nolint:errcheck
		Ack: []string{"jti-001", "jti-002"},
	})

	if len(receivedAck) != 2 {
		t.Errorf("ack count: got %d, want 2", len(receivedAck))
	}
}

// ─── Full subscription lifecycle ─────────────────────────────────────────

func TestEventSubscriptionLifecycle(t *testing.T) {
	subs := map[string]models.OBEventSubscriptionResponseData1{}
	nextID := 1

	mux := http.NewServeMux()

	// POST
	mux.HandleFunc("/open-banking/v3.1/event-subscriptions",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				id := "sub-lifecycle-1"
				subs[id] = models.OBEventSubscriptionResponseData1{
					EventSubscriptionId: id,
					CallbackUrl:         "https://tpp.example.com/events",
					Version:             "3.1",
				}
				nextID++
				w.WriteHeader(http.StatusCreated)
    if err := json.NewEncoder(w).Encode(models.OBEventSubscriptionResponse1{Data: subs[id]}); err != nil {
    	http.Error(w, err.Error(), http.StatusInternalServerError)
    }
			case http.MethodGet:
				var list []models.OBEventSubscriptionResponseData1
				for _, s := range subs {
					list = append(list, s)
				}
				if err := json.NewEncoder(w).Encode(models.OBEventSubscriptionsResponse1{
					Data: models.OBEventSubscriptionsResponseData1{EventSubscription: list},
				}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			}
		})

	// PUT + DELETE
	mux.HandleFunc("/open-banking/v3.1/event-subscriptions/sub-lifecycle-1",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPut:
				var req models.OBEventSubscriptionResponse1
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    	http.Error(w, err.Error(), http.StatusBadRequest)
    	return
    }
				subs["sub-lifecycle-1"] = req.Data
    if err := json.NewEncoder(w).Encode(models.OBEventSubscriptionResponse1{Data: req.Data}); err != nil {
    	http.Error(w, err.Error(), http.StatusInternalServerError)
    }
			case http.MethodDelete:
				delete(subs, "sub-lifecycle-1")
				w.WriteHeader(http.StatusNoContent)
			}
		})

	svc, _ := newSvc(t, mux)
	ctx := context.Background()

	// 1. Create
	created, err := svc.CreateEventSubscription(ctx, &models.OBEventSubscription1{
		Data: models.OBEventSubscriptionData1{CallbackUrl: "https://tpp.example.com/events", Version: "3.1"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// 2. Get
	listed, err := svc.GetEventSubscriptions(ctx)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(listed.Data.EventSubscription) != 1 {
		t.Errorf("expected 1 subscription, got %d", len(listed.Data.EventSubscription))
	}

	// 3. Update
	_, err = svc.UpdateEventSubscription(ctx, created.Data.EventSubscriptionId,
		&models.OBEventSubscriptionResponse1{
			Data: models.OBEventSubscriptionResponseData1{
				EventSubscriptionId: created.Data.EventSubscriptionId,
				CallbackUrl:         "https://tpp.example.com/events/v2",
				Version:             "3.1",
			},
		})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	// 4. Delete
	if err := svc.DeleteEventSubscription(ctx, created.Data.EventSubscriptionId); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if len(subs) != 0 {
		t.Errorf("expected subs to be empty after delete, got %d", len(subs))
	}
}