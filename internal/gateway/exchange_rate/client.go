package exchangerate

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/slackerkids/plata-currency-exchange.git/internal/model"
)

const (
	latestResultsPath = "latest"
)

type RatesResponse struct {
	Success   bool               `json:"success"`
	TimeStamp int                `json:"timestamp"`
	Base      string             `json:"base"`
	Date      string             `json:"date"`
	Rates     map[string]float64 `json:"rates"`
}

type Client struct {
	baseURL string
	apiKey  string
}

func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

func (c *Client) GetQuoteByCurrencyCode(ctx context.Context, base, quote string) (*model.QuoteResult, error) {
	request, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+latestResultsPath, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("get request: %w", err)
	}
	defer request.Body.Close()

	values := request.URL.Query()

	values.Add("access_key", c.apiKey)
	values.Add("base", base)
	values.Add("symbols", quote)

	request.URL.RawQuery = values.Encode()

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("get response: %w", err)
	}
	defer response.Body.Close()

	var result RatesResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding rates response: %w", err)
	}

	return &model.QuoteResult{
		Base:  base,
		Quote: quote,
		Rate:  result.Rates[quote],
	}, nil
}
