package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type GetIntegrationResponse struct {
	Integration Integration `json:"integration"`
}

type IntegrationData struct {
	OpsGenieIntegrationData *OpsGenieIntegrationData
	SlackIntegrationData    *SlackIntegrationData
}

type IntegrationStatus string

// List of IntegrationStatus
const (
	ACTIVE  IntegrationStatus = "active"
	DELETED IntegrationStatus = "deleted"
)

type Integration struct {
	Created        time.Time              `json:"created"`
	Data           map[string]interface{} `json:"data"`
	Description    string                 `json:"description"`
	Id             string                 `json:"id"`
	OrganizationId string                 `json:"organizationId"`
	ProjectId      string                 `json:"projectId"`
	Status         IntegrationStatus      `json:"status"`
	Updated        time.Time              `json:"updated"`
}

type ListIntegrationsResponse struct {
	Integrations []Integration `json:"integrations"`
}

type OpsGenieIntegrationData struct {
	// API key used with the Ops Genie integration API
	ApiKeyDisplay string `json:"apiKeyDisplay"`
	// Required. Must be set to \"opsGenie\"
	Sink string `json:"sink"`
	// Source of data for integration
	Source string `json:"source"`
}

type SlackIntegrationData struct {
	// Slack Channel to send messages to
	ChannelId string `json:"channelId"`
	// API token for the Slack bot
	TokenDisplay string `json:"tokenDisplay"`
	// Required. Must be set to \"slack\"
	Sink string `json:"sink"`
	// Source of data for integration
	Source string `json:"source"`
}

func (c *Client) GetIntegration(ctx context.Context, organizationId string, projectId string, integrationId string) (*GetIntegrationResponse, error) {

	url := *c.apiURL
	url.Path = "/integrate/v1/organizations/{organizationId}/projects/{projectId}/integrations/{integrationId}"
	url.Path = strings.Replace(url.Path, "{"+"organizationId"+"}", organizationId, -1)
	url.Path = strings.Replace(url.Path, "{"+"projectId"+"}", projectId, -1)
	url.Path = strings.Replace(url.Path, "{"+"integrationId"+"}", integrationId, -1)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error constructing request for GetIntegration: %w", err)
	}
	request.Header.Add("Content-Type", "application/json")
	if err := c.addAuthorizationHeader(request); err != nil {
		return nil, err
	}

	log.Printf("[BESPIN-C] 1 url=%q\n", url.Path)

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error sending request for GetIntegration: %w", err)
	}
	defer closeIgnoreError(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		log.Printf("[BESPIN-C] 2 resp.StatusCode=%d\n", resp.StatusCode)
		return nil, translateStatusCode(resp.StatusCode, "GetIntegration", resp.Body)
	}

	decoder := json.NewDecoder(resp.Body)
	result := GetIntegrationResponse{}
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &result, nil
}

func (c *Client) ListIntegrations(ctx context.Context, organizationId string, projectId string) (*ListIntegrationsResponse, error) {

	url := *c.apiURL
	url.Path = "/organizations/{organizationId}/projects/{projectId}/integrations"
	url.Path = strings.Replace(url.Path, "{"+"organizationId"+"}", organizationId, -1)
	url.Path = strings.Replace(url.Path, "{"+"projectId"+"}", projectId, -1)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error constructing request for ListIntegrations: %w", err)
	}
	request.Header.Add("Content-Type", "application/json")
	if err := c.addAuthorizationHeader(request); err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error sending request for ListIntegrations: %w", err)
	}
	defer closeIgnoreError(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return nil, translateStatusCode(resp.StatusCode, "ListIntegrations", resp.Body)
	}

	decoder := json.NewDecoder(resp.Body)
	result := ListIntegrationsResponse{}
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &result, nil
}
