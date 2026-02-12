package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client wraps HTTP calls to the EONAPI.
type Client struct {
	BaseURL    string
	Username   string
	APIKey     string
	HTTPClient *http.Client
}

// APIResponse is the generic shape returned by every EONAPI endpoint.
type APIResponse struct {
	APIVersion string      `json:"api_version,omitempty"`
	HTTPCode   string      `json:"http_code"`
	Result     interface{} `json:"result,omitempty"`
	Status     string      `json:"Status,omitempty"`
	EONAPIKey  string      `json:"EONAPI_KEY,omitempty"`
}

// NewClient creates a new EONAPI HTTP client.
func NewClient(baseURL, username, apiKey string, insecure bool) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure}, //nolint:gosec
	}
	return &Client{
		BaseURL:  strings.TrimRight(baseURL, "/"),
		Username: username,
		APIKey:   apiKey,
		HTTPClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: tr,
		},
	}
}

func (c *Client) authQS() string {
	return fmt.Sprintf("username=%s&apiKey=%s",
		url.QueryEscape(c.Username), url.QueryEscape(c.APIKey))
}

// Get performs an authenticated GET.
func (c *Client) Get(endpoint string) (*APIResponse, error) {
	u := fmt.Sprintf("%s/%s?&%s", c.BaseURL, endpoint, c.authQS())
	resp, err := c.HTTPClient.Get(u)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", endpoint, err)
	}
	defer resp.Body.Close()
	return decodeResp(resp)
}

// Post performs an authenticated POST with JSON body.
func (c *Client) Post(endpoint string, body interface{}) (*APIResponse, error) {
	u := fmt.Sprintf("%s/%s?&%s", c.BaseURL, endpoint, c.authQS())
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal %s: %w", endpoint, err)
	}
	req, err := http.NewRequest("POST", u, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", endpoint, err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST %s: %w", endpoint, err)
	}
	defer resp.Body.Close()
	return decodeResp(resp)
}

func decodeResp(resp *http.Response) (*APIResponse, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("401 Unauthorized – check username/apiKey")
	}
	var r APIResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("status %d, raw: %s", resp.StatusCode, string(body))
	}
	if resp.StatusCode >= 400 {
		return &r, fmt.Errorf("API %d: %s", resp.StatusCode, r.HTTPCode)
	}
	return &r, nil
}

// CheckAuth validates credentials against getAuthenticationStatus.
func (c *Client) CheckAuth() error {
	r, err := c.Get("getAuthenticationStatus")
	if err != nil {
		return err
	}
	if r.Status != "Authorized" {
		return fmt.Errorf("not authorized: %s", r.Status)
	}
	return nil
}

// ─── Host ─────────────────────────────────────────────────────────

func (c *Client) CreateHost(body map[string]interface{}) (*APIResponse, error) {
	return c.Post("createHost", body)
}

func (c *Client) GetHost(name string) (*APIResponse, error) {
	return c.Post("getHost", map[string]string{"hostName": name})
}

func (c *Client) DeleteHost(name string, export bool) (*APIResponse, error) {
	return c.Post("deleteHost", map[string]interface{}{
		"hostName": name, "exportConfiguration": export,
	})
}

func (c *Client) AddHostTemplateToHost(tpl, host string, export bool) (*APIResponse, error) {
	return c.Post("addHostTemplateToHost", map[string]interface{}{
		"templateHostName": tpl, "hostName": host, "exportConfiguration": export,
	})
}

// ─── Command (check) ──────────────────────────────────────────────

func (c *Client) AddCommand(name, line, desc string) (*APIResponse, error) {
	return c.Post("addCommand", map[string]string{
		"commandName": name, "commandLine": line, "commandDescription": desc,
	})
}

func (c *Client) GetCommand(name string) (*APIResponse, error) {
	return c.Post("getCommand", map[string]string{"commandName": name})
}

func (c *Client) ModifyCommand(name, newName, line, desc string) (*APIResponse, error) {
	return c.Post("modifyCommand", map[string]string{
		"commandName": name, "newCommandName": newName,
		"commandLine": line, "commandDescription": desc,
	})
}

func (c *Client) DeleteCommand(name string) (*APIResponse, error) {
	return c.Post("deleteCommand", map[string]string{"commandName": name})
}

// ─── Contact ──────────────────────────────────────────────────────

func (c *Client) CreateContact(body map[string]interface{}) (*APIResponse, error) {
	return c.Post("createContact", body)
}

func (c *Client) GetContact(name string) (*APIResponse, error) {
	return c.Post("getContact", map[string]interface{}{"contactName": name})
}

func (c *Client) ModifyContact(body map[string]interface{}) (*APIResponse, error) {
	return c.Post("modifyContact", body)
}

func (c *Client) DeleteContact(name string) (*APIResponse, error) {
	return c.Post("deleteContact", map[string]string{"contactName": name})
}

// ─── Contact Group ────────────────────────────────────────────────

func (c *Client) CreateContactGroup(name, desc string, export bool) (*APIResponse, error) {
	return c.Post("createContactGroup", map[string]interface{}{
		"contactGroupName": name, "description": desc, "exportConfiguration": export,
	})
}

func (c *Client) GetContactGroup(name string) (*APIResponse, error) {
	return c.Post("getContactGroup", map[string]interface{}{"contactGroupName": name})
}

func (c *Client) ModifyContactGroup(body map[string]interface{}) (*APIResponse, error) {
	return c.Post("modifyContactGroup", body)
}

func (c *Client) DeleteContactGroup(name string) (*APIResponse, error) {
	return c.Post("deleteContactGroup", map[string]string{"contactGroupName": name})
}

// ─── Export ───────────────────────────────────────────────────────

func (c *Client) ExportConfiguration(job string) (*APIResponse, error) {
	return c.Post("exportConfiguration", map[string]string{"JobName": job})
}
