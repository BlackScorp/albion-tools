// Package marketapi retrieves current market prices from the Albion Data API.
package marketapi

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/blackscorp/albion-helper/internal/catalog"
)

const maxURLLength = 4096

// Server identifies one of Albion's regional API hosts.
type Server string

const (
	Europe   Server = "Europe"
	Americas Server = "Americas"
	Asia     Server = "Asia"
)

func (s Server) baseURL() (string, error) {
	switch s {
	case Europe:
		return "https://europe.albion-online-data.com", nil
	case Americas:
		return "https://west.albion-online-data.com", nil
	case Asia:
		return "https://east.albion-online-data.com", nil
	default:
		return "", fmt.Errorf("unknown Albion server %q", s)
	}
}

// Config customizes a Client. Zero values use production-safe defaults.
type Config struct {
	HTTPClient      *http.Client
	BaseURL         string
	RequestTimeout  time.Duration
	RequestInterval time.Duration
	MaxURLLength    int
	Now             func() time.Time
}

// Client retrieves and decodes current price observations.
type Client struct {
	httpClient      *http.Client
	baseURL         string
	timeout         time.Duration
	requestInterval time.Duration
	maxURLLength    int
	now             func() time.Time

	rateMu      sync.Mutex
	nextRequest time.Time
}

// NewClient creates a client for a production Albion API host.
func NewClient(server Server) (*Client, error) {
	baseURL, err := server.baseURL()
	if err != nil {
		return nil, err
	}
	return NewClientWithConfig(Config{BaseURL: baseURL})
}

// NewClientWithConfig creates a client and is also suitable for local HTTP tests.
func NewClientWithConfig(config Config) (*Client, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("API base URL must not be empty")
	}
	parsed, err := url.Parse(config.BaseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid API base URL %q", config.BaseURL)
	}
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{}
	}
	if config.RequestTimeout <= 0 {
		config.RequestTimeout = 15 * time.Second
	}
	if config.RequestInterval <= 0 {
		// One request per second stays below both published API limits.
		config.RequestInterval = time.Second
	}
	if config.MaxURLLength <= 0 {
		config.MaxURLLength = maxURLLength
	}
	if config.Now == nil {
		config.Now = time.Now
	}
	return &Client{
		httpClient:      config.HTTPClient,
		baseURL:         strings.TrimRight(config.BaseURL, "/"),
		timeout:         config.RequestTimeout,
		requestInterval: config.RequestInterval,
		maxURLLength:    config.MaxURLLength,
		now:             config.Now,
	}, nil
}

type apiPrice struct {
	ItemID           string `json:"item_id"`
	City             string `json:"city"`
	Quality          int    `json:"quality"`
	SellPriceMin     int64  `json:"sell_price_min"`
	SellPriceMinDate string `json:"sell_price_min_date"`
	BuyPriceMax      int64  `json:"buy_price_max"`
	BuyPriceMaxDate  string `json:"buy_price_max_date"`
}

// FetchPrices retrieves all current observations for the supplied item IDs.
// IDs are split into requests whose complete URL stays below the API limit.
func (c *Client) FetchPrices(ctx context.Context, itemIDs []string) ([]catalog.Price, error) {
	return c.FetchPricesAt(ctx, itemIDs, nil)
}

// FetchPricesAt retrieves observations for the selected market locations. An
// empty location list requests every location, as supported by the API.
func (c *Client) FetchPricesAt(ctx context.Context, itemIDs []string, markets []catalog.Market) ([]catalog.Price, error) {
	if len(itemIDs) == 0 {
		return []catalog.Price{}, nil
	}
	locationQuery := locationFilter(markets)
	batches, err := c.batches(itemIDs, locationQuery)
	if err != nil {
		return nil, err
	}
	prices := make([]catalog.Price, 0)
	for _, batch := range batches {
		batchPrices, err := c.fetchBatch(ctx, batch, locationQuery)
		if err != nil {
			return nil, err
		}
		prices = append(prices, batchPrices...)
	}
	return prices, nil
}

func (c *Client) batches(itemIDs []string, locations string) ([][]string, error) {
	var batches [][]string
	current := make([]string, 0)
	for _, itemID := range itemIDs {
		if itemID == "" || strings.ContainsAny(itemID, "/?#,\t\r\n") {
			return nil, fmt.Errorf("invalid item ID %q", itemID)
		}
		candidate := append(append([]string(nil), current...), itemID)
		if len(c.endpoint(candidate, locations)) > c.maxURLLength {
			if len(current) == 0 {
				return nil, fmt.Errorf("item ID %q cannot fit below URL limit", itemID)
			}
			batches = append(batches, current)
			current = []string{itemID}
		} else {
			current = candidate
		}
	}
	if len(current) > 0 {
		batches = append(batches, current)
	}
	return batches, nil
}

func locationFilter(markets []catalog.Market) string {
	values := make([]string, 0, len(markets))
	for _, market := range markets {
		if market == catalog.BlackMarket {
			values = append(values, "Blackmarket")
		} else {
			values = append(values, string(market))
		}
	}
	return strings.Join(values, ",")
}

func (c *Client) endpoint(itemIDs []string, locations string) string {
	encodedIDs := make([]string, 0, len(itemIDs))
	for _, itemID := range itemIDs {
		encodedIDs = append(encodedIDs, url.PathEscape(itemID))
	}
	path := "/api/v2/stats/prices/" + strings.Join(encodedIDs, ",") + ".json"
	if locations == "" {
		return c.baseURL + path
	}
	return c.baseURL + path + "?locations=" + url.QueryEscape(locations)
}

func (c *Client) fetchBatch(ctx context.Context, itemIDs []string, locations string) ([]catalog.Price, error) {
	if err := c.waitForRate(ctx); err != nil {
		return nil, err
	}
	requestCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	request, err := http.NewRequestWithContext(requestCtx, http.MethodGet, c.endpoint(itemIDs, locations), nil)
	if err != nil {
		return nil, fmt.Errorf("create prices request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Accept-Encoding", "gzip")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request prices for %s: %w", strings.Join(itemIDs, ","), err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("prices request for %s returned HTTP %s", strings.Join(itemIDs, ","), response.Status)
	}
	body, err := responseBody(response)
	if err != nil {
		return nil, fmt.Errorf("read prices response for %s: %w", strings.Join(itemIDs, ","), err)
	}
	var observations []apiPrice
	if err := json.Unmarshal(body, &observations); err != nil {
		return nil, fmt.Errorf("decode prices response for %s: %w", strings.Join(itemIDs, ","), err)
	}
	prices := make([]catalog.Price, 0, len(observations))
	for _, observation := range observations {
		updatedAt, err := c.observationTime(observation.SellPriceMinDate, observation.BuyPriceMaxDate)
		if err != nil {
			return nil, fmt.Errorf("decode timestamp for %s: %w", observation.ItemID, err)
		}
		prices = append(prices, catalog.Price{
			ItemID:    observation.ItemID,
			Market:    catalog.Market(observation.City),
			Quality:   catalog.Quality(observation.Quality),
			Buy:       observation.SellPriceMin,
			Sell:      observation.BuyPriceMax,
			UpdatedAt: updatedAt,
		})
	}
	return prices, nil
}

func responseBody(response *http.Response) ([]byte, error) {
	var reader io.Reader = response.Body
	if strings.EqualFold(response.Header.Get("Content-Encoding"), "gzip") {
		gzipReader, err := gzip.NewReader(response.Body)
		if err != nil {
			return nil, fmt.Errorf("open gzip body: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}
	return io.ReadAll(reader)
}

func (c *Client) observationTime(sellDate, buyDate string) (int64, error) {
	latest := time.Time{}
	for _, value := range []string{sellDate, buyDate} {
		if value == "" {
			continue
		}
		var parsed time.Time
		var err error
		for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05.999999999", "2006-01-02T15:04:05"} {
			parsed, err = time.Parse(layout, value)
			if err == nil {
				break
			}
		}
		if err != nil {
			return 0, fmt.Errorf("%q: %w", value, err)
		}
		if parsed.After(latest) {
			latest = parsed
		}
	}
	if latest.IsZero() {
		return c.now().Unix(), nil
	}
	return latest.Unix(), nil
}

func (c *Client) waitForRate(ctx context.Context) error {
	c.rateMu.Lock()
	now := c.now()
	start := now
	if c.nextRequest.After(start) {
		start = c.nextRequest
	}
	c.nextRequest = start.Add(c.requestInterval)
	c.rateMu.Unlock()

	wait := start.Sub(now)
	if wait <= 0 {
		return nil
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return fmt.Errorf("wait for API rate limit: %w", ctx.Err())
	case <-timer.C:
		return nil
	}
}
