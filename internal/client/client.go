package client

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

var _ = url.PathEscape

const defaultBaseURL = "https://rest.websupport.sk"

type Client struct {
	BaseURL    string
	APIKey     string
	APISecret  string
	HTTPClient *http.Client
}

func New(apiKey, apiSecret, baseURL string) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Client{
		BaseURL:    baseURL,
		APIKey:     apiKey,
		APISecret:  apiSecret,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Record matches the Websupport REST API record shape.
type Record struct {
	ID       int64  `json:"id,omitempty"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Content  string `json:"content"`
	TTL      int    `json:"ttl,omitempty"`
	Priority *int   `json:"prio,omitempty"`
	Note     string `json:"note,omitempty"`
}

type listResponse struct {
	Items []Record `json:"items"`
}

type apiError struct {
	Status     string                 `json:"status"`
	Message    string                 `json:"message"`
	Errors     map[string]interface{} `json:"errors,omitempty"`
	Item       *Record                `json:"item,omitempty"`
	HTTPStatus int                    `json:"-"`
	Body       string                 `json:"-"`
}

func (e *apiError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("websupport API error (%d): %s", e.HTTPStatus, e.Message)
	}
	return fmt.Sprintf("websupport API error (%d): %s", e.HTTPStatus, e.Body)
}

// IsNotFound reports whether err represents a 404 from the Websupport API.
func IsNotFound(err error) bool {
	if e, ok := err.(*apiError); ok {
		return e.HTTPStatus == http.StatusNotFound
	}
	return false
}

func (c *Client) do(method, endpoint string, body interface{}, out interface{}) error {
	full := c.BaseURL + endpoint

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, full, reqBody)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	// Websupport HMAC-SHA1 auth: signature over "{METHOD} {PATH} {UNIX_TS}",
	// sent via Basic auth (username=apikey, password=hex(hmac)).
	ts := time.Now().Unix()
	canonical := fmt.Sprintf("%s %s %d", method, endpoint, ts)
	mac := hmac.New(sha1.New, []byte(c.APISecret))
	mac.Write([]byte(canonical))
	signature := hex.EncodeToString(mac.Sum(nil))

	req.SetBasicAuth(c.APIKey, signature)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Date", time.Unix(ts, 0).UTC().Format(time.RFC1123))
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		ae := &apiError{HTTPStatus: resp.StatusCode, Body: string(respBody)}
		_ = json.Unmarshal(respBody, ae)
		return ae
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decode response: %w (body: %s)", err, string(respBody))
		}
	}
	return nil
}

// CreateRecord creates a DNS record in the given zone.
func (c *Client) CreateRecord(zone string, r Record) (*Record, error) {
	endpoint := fmt.Sprintf("/v1/user/self/zone/%s/record", url.PathEscape(zone))
	var resp struct {
		Status string `json:"status"`
		Item   Record `json:"item"`
	}
	if err := c.do(http.MethodPost, endpoint, r, &resp); err != nil {
		return nil, err
	}
	return &resp.Item, nil
}

// GetRecord fetches a record by ID.
func (c *Client) GetRecord(zone string, id int64) (*Record, error) {
	endpoint := fmt.Sprintf("/v1/user/self/zone/%s/record/%d", url.PathEscape(zone), id)
	var r Record
	if err := c.do(http.MethodGet, endpoint, nil, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// UpdateRecord patches an existing record by ID.
func (c *Client) UpdateRecord(zone string, id int64, r Record) (*Record, error) {
	endpoint := fmt.Sprintf("/v1/user/self/zone/%s/record/%d", url.PathEscape(zone), id)
	var resp struct {
		Status string `json:"status"`
		Item   Record `json:"item"`
	}
	if err := c.do(http.MethodPut, endpoint, r, &resp); err != nil {
		return nil, err
	}
	return &resp.Item, nil
}

// DeleteRecord deletes a record by ID. 404 is treated as success.
func (c *Client) DeleteRecord(zone string, id int64) error {
	endpoint := fmt.Sprintf("/v1/user/self/zone/%s/record/%d", url.PathEscape(zone), id)
	if err := c.do(http.MethodDelete, endpoint, nil, nil); err != nil {
		if IsNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}

// ListRecords returns all records in a zone.
func (c *Client) ListRecords(zone string) ([]Record, error) {
	endpoint := fmt.Sprintf("/v1/user/self/zone/%s/record", url.PathEscape(zone))
	var lr listResponse
	if err := c.do(http.MethodGet, endpoint, nil, &lr); err != nil {
		return nil, err
	}
	return lr.Items, nil
}
