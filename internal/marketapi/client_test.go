package marketapi

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/blackscorp/albion-helper/internal/catalog"
)

func TestFetchPricesDecodesGzipAndMapsMVPQuotes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/v2/stats/prices/T4_WOOD,T4_CAPE.json" {
			t.Fatalf("request path = %q", request.URL.Path)
		}
		if request.Header.Get("Accept-Encoding") != "gzip" {
			t.Fatal("request did not advertise gzip")
		}
		w.Header().Set("Content-Encoding", "gzip")
		writer := gzip.NewWriter(w)
		defer writer.Close()
		_ = json.NewEncoder(writer).Encode([]apiPrice{
			{ItemID: "T4_WOOD", City: "Martlock", Quality: 2, SellPriceMin: 410, SellPriceMinDate: "2026-09-16T10:00:00", BuyPriceMax: 525, BuyPriceMaxDate: "2026-09-16T10:05:00"},
		})
	}))
	defer server.Close()

	client, err := NewClientWithConfig(Config{BaseURL: server.URL, RequestInterval: time.Nanosecond, RequestTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	got, err := client.FetchPrices(context.Background(), []string{"T4_WOOD", "T4_CAPE"})
	if err != nil {
		t.Fatalf("FetchPrices() error = %v", err)
	}
	want := []catalog.Price{{ItemID: "T4_WOOD", Market: catalog.Market("Martlock"), Quality: 2, Buy: 410, Sell: 525, UpdatedAt: time.Date(2026, time.September, 16, 10, 5, 0, 0, time.UTC).Unix()}}
	if fmt.Sprintf("%#v", got) != fmt.Sprintf("%#v", want) {
		t.Fatalf("FetchPrices() = %#v, want %#v", got, want)
	}
}

func TestFetchPricesBatchesBelowConfiguredURLLimit(t *testing.T) {
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		paths = append(paths, request.URL.Path)
		_, _ = io.WriteString(w, `[]`)
	}))
	defer server.Close()

	client, err := NewClientWithConfig(Config{BaseURL: server.URL, MaxURLLength: len(server.URL) + len("/api/v2/stats/prices/T4_A.json") + 1, RequestInterval: time.Nanosecond})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.FetchPrices(context.Background(), []string{"T4_A", "T4_B", "T4_C"}); err != nil {
		t.Fatalf("FetchPrices() error = %v", err)
	}
	if len(paths) != 3 {
		t.Fatalf("request count = %d, want 3 (%v)", len(paths), paths)
	}
}

func TestFetchPricesAtAddsSelectedLocations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if got := request.URL.Query().Get("locations"); got != "Caerleon,Blackmarket,Brecilien" {
			t.Fatalf("locations query = %q", got)
		}
		_, _ = io.WriteString(w, `[]`)
	}))
	defer server.Close()
	client, err := NewClientWithConfig(Config{BaseURL: server.URL, RequestInterval: time.Nanosecond})
	if err != nil {
		t.Fatal(err)
	}
	markets := []catalog.Market{catalog.Caerleon, catalog.BlackMarket, catalog.Brecilien}
	if _, err := client.FetchPricesAt(context.Background(), []string{"T4_WOOD"}, markets); err != nil {
		t.Fatalf("FetchPricesAt() error = %v", err)
	}
}

func TestFetchPricesAllowsPartialAPIResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		_, _ = io.WriteString(w, `[{"item_id":"T4_WOOD","city":"Martlock","quality":1,"sell_price_min":10,"buy_price_max":20}]`)
	}))
	defer server.Close()

	client, err := NewClientWithConfig(Config{BaseURL: server.URL, RequestInterval: time.Nanosecond, Now: func() time.Time { return time.Unix(123, 0) }})
	if err != nil {
		t.Fatal(err)
	}
	got, err := client.FetchPrices(context.Background(), []string{"T4_WOOD", "T4_CAPE"})
	if err != nil || len(got) != 1 || got[0].UpdatedAt != 123 {
		t.Fatalf("FetchPrices() = %#v, error %v; want one partial row with fallback timestamp", got, err)
	}
}

func TestFetchPricesReportsInvalidJSONAndHTTPError(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		want    string
	}{
		{name: "invalid JSON", handler: func(w http.ResponseWriter, request *http.Request) { _, _ = io.WriteString(w, `{`) }, want: "decode prices response"},
		{name: "HTTP error", handler: func(w http.ResponseWriter, request *http.Request) {
			http.Error(w, "upstream failed", http.StatusBadGateway)
		}, want: "returned HTTP 502"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()
			client, err := NewClientWithConfig(Config{BaseURL: server.URL, RequestInterval: time.Nanosecond})
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.FetchPrices(context.Background(), []string{"T4_WOOD"})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("FetchPrices() error = %v, want substring %q", err, tt.want)
			}
		})
	}
}

func TestFetchPricesHonorsTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		<-request.Context().Done()
	}))
	defer server.Close()
	client, err := NewClientWithConfig(Config{BaseURL: server.URL, RequestTimeout: 10 * time.Millisecond, RequestInterval: time.Nanosecond})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.FetchPrices(context.Background(), []string{"T4_WOOD"})
	if err == nil || !strings.Contains(err.Error(), "request prices") {
		t.Fatalf("FetchPrices() error = %v, want timeout context", err)
	}
}

func TestNewClientRejectsUnknownServer(t *testing.T) {
	if _, err := NewClient(Server("Test")); err == nil {
		t.Fatal("NewClient() accepted unknown server")
	}
}
