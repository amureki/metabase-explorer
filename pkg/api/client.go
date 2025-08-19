package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type MetabaseClient struct {
	BaseURL    string
	APIToken   string
	HTTPClient *http.Client
}

func NewMetabaseClient(baseURL, apiToken string) *MetabaseClient {
	return &MetabaseClient{
		BaseURL:    baseURL,
		APIToken:   apiToken,
		HTTPClient: &http.Client{},
	}
}

func (c *MetabaseClient) TestConnection() error {
	baseURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %v", err)
	}

	apiURL, err := baseURL.Parse("/api/user/current")
	if err != nil {
		return fmt.Errorf("failed to construct API URL: %v", err)
	}

	req, _ := http.NewRequest("GET", apiURL.String(), nil)
	req.Header.Set("X-API-Key", c.APIToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API token authentication failed with status: %d - %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *MetabaseClient) GetDatabases() ([]Database, error) {
	baseURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %v", err)
	}

	apiURL, err := baseURL.Parse("/api/database")
	if err != nil {
		return nil, fmt.Errorf("failed to construct API URL: %v", err)
	}

	req, _ := http.NewRequest("GET", apiURL.String(), nil)
	req.Header.Set("X-API-Key", c.APIToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get databases: %d - %s", resp.StatusCode, string(body))
	}

	var result map[string][]Database
	json.NewDecoder(resp.Body).Decode(&result)
	return result["data"], nil
}

func (c *MetabaseClient) GetTables(databaseID int) ([]Table, error) {
	baseURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %v", err)
	}

	apiURL, err := baseURL.Parse(fmt.Sprintf("/api/database/%d/metadata", databaseID))
	if err != nil {
		return nil, fmt.Errorf("failed to construct API URL: %v", err)
	}

	req, _ := http.NewRequest("GET", apiURL.String(), nil)
	req.Header.Set("X-API-Key", c.APIToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get tables: %d - %s", resp.StatusCode, string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	var metadata struct {
		Tables []Table `json:"tables"`
	}

	if err := json.Unmarshal(body, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return metadata.Tables, nil
}

func (c *MetabaseClient) GetTableFields(tableID int) ([]Field, error) {
	baseURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %v", err)
	}

	apiURL, err := baseURL.Parse(fmt.Sprintf("/api/table/%d/query_metadata", tableID))
	if err != nil {
		return nil, fmt.Errorf("failed to construct API URL: %v", err)
	}

	req, _ := http.NewRequest("GET", apiURL.String(), nil)
	req.Header.Set("X-API-Key", c.APIToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get table fields: %d - %s", resp.StatusCode, string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	var queryMeta struct {
		Fields []Field `json:"fields"`
	}

	if err := json.Unmarshal(body, &queryMeta); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return queryMeta.Fields, nil
}
