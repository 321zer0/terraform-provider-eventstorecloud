package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type UpdateIntegrationRequest struct {
	Data        *map[string]interface{} `json:"data,omitempty"`
	Description *string                 `json:"description,omitempty"`
}

type UpdateOpsGenieIntegrationData struct {
	// API key used with the Ops Genie integration API
	ApiKey *string `json:"apiKey,omitempty"`
}

type UpdateSlackIntegrationData struct {
	// Slack Channel to send messages to
	ChannelId *string `json:"channelId,omitempty"`
	// API token for the Slack bot
	Token *string `json:"token,omitempty"`
}

func (c *Client) UpdateIntegration(ctx context.Context, organizationId string, projectId string, integrationId string, updateIntegrationRequest UpdateIntegrationRequest) error {
	requestBody, err := json.Marshal(updateIntegrationRequest)
	if err != nil {
		return fmt.Errorf("error marshalling request: %w", err)
	}

	url := *c.apiURL
	url.Path = "/integrate/v1/organizations/{organizationId}/projects/{projectId}/integrations/{integrationId}"
	url.Path = strings.Replace(url.Path, "{"+"organizationId"+"}", organizationId, -1)
	url.Path = strings.Replace(url.Path, "{"+"projectId"+"}", projectId, -1)
	url.Path = strings.Replace(url.Path, "{"+"integrationId"+"}", integrationId, -1)

	request, err := http.NewRequestWithContext(ctx, http.MethodPut, url.String(), bytes.NewReader(requestBody))
	if err != nil {
		log.Println("[BESPIN-U] 3 error consturction")
		return fmt.Errorf("error constructing request for UpdateIntegration: %w", err)
	}
	request.Header.Add("Content-Type", "application/json")
	if err := c.addAuthorizationHeader(request); err != nil {
		log.Println("[BESPIN-U] 4 auth bad")
		return err
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		log.Println("[BESPIN-U] 5")
		log.Printf("[BESPIN-U] statusCode=%d\n", resp.StatusCode)
		return fmt.Errorf("error sending request for UpdateIntegration: %w", err)
	}
	defer closeIgnoreError(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return translateStatusCode(resp.StatusCode, "UpdateIntegration", resp.Body)
	}

	return nil
}
