package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

func (c *Client) DeleteIntegration(ctx context.Context, organizationId string, projectId string, integrationId string) error {

	url := *c.apiURL
	url.Path = "/organizations/{organizationId}/projects/{projectId}/integrations/{integrationId}"
	url.Path = strings.Replace(url.Path, "{"+"organizationId"+"}", organizationId, -1)
	url.Path = strings.Replace(url.Path, "{"+"projectId"+"}", projectId, -1)
	url.Path = strings.Replace(url.Path, "{"+"integrationId"+"}", integrationId, -1)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), nil)
	if err != nil {
		return fmt.Errorf("error constructing request for DeleteIntegration: %w", err)
	}
	request.Header.Add("Content-Type", "application/json")
	if err := c.addAuthorizationHeader(request); err != nil {
		return err
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("error sending request for DeleteIntegration: %w", err)
	}
	defer closeIgnoreError(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return translateStatusCode(resp.StatusCode, "DeleteIntegration", resp.Body)
	}

	return nil
}
