package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type TreasuryRateResponse struct {
	Data []struct {
		RecordDate       string `json:"record_date"`
		CountryCurrency  string `json:"country_currency_desc"`
		ExchangeRate     string `json:"exchange_rate"`
		EffectiveDate    string `json:"effective_date"`
	} `json:"data"`
}

type TreasuryAPIDAO struct {
	baseURL string
	client  *http.Client
}

func NewTreasuryAPIDAO() *TreasuryAPIDAO {
	return &TreasuryAPIDAO{
		baseURL: "https://api.fiscaldata.treasury.gov/services/api/fiscal_service/v1/accounting/od/rates_of_exchange",
		client:  &http.Client{},
	}
}

func (d *TreasuryAPIDAO) FetchRates(ctx context.Context, currency string, startDate, endDate string) (*TreasuryRateResponse, error) {
	filter := fmt.Sprintf("currency:eq:%s,record_date:gte:%s,record_date:lte:%s", currency, startDate, endDate)
	
	u, _ := url.Parse(d.baseURL)
	q := u.Query()
	q.Set("filter", filter)
	q.Set("sort", "-record_date")
	q.Set("page[size]", "1")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("treasury api error: status %d", resp.StatusCode)
	}

	var result TreasuryRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
