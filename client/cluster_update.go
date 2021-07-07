package client

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"net/http"
	"path"
)

type ExpandManagedClusterDiskRequest struct {
	OrganizationID string
	ProjectID      string
	ClusterID      string `json:"clusterId"`
	DiskSizeGB     int32  `json:"diskSizeGb"`
}

func (c *Client) ManagedClusterExpandDisk(ctx context.Context, req *ExpandManagedClusterDiskRequest) diag.Diagnostics {
	requestBody, err := json.Marshal(req)
	if err != nil {
		return diag.Errorf("error marshalling request: %w", err)
	}

	requestURL := *c.apiURL
	requestURL.Path = path.Join("mesdb", "v1", "organizations", req.OrganizationID, "projects", req.ProjectID, "clusters", req.ClusterID, "disk", "expand")

	request, err := http.NewRequestWithContext(ctx, http.MethodPut, requestURL.String(), bytes.NewReader(requestBody))
	if err != nil {
		return diag.Errorf("error constructing request: %w", err)
	}
	request.Header.Add("Content-Type", "application/json")
	if err := c.addAuthorizationHeader(request); err != nil {
		return err
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return diag.Errorf("error sending request: %w", err)
	}
	defer closeIgnoreError(resp.Body)

	if resp.StatusCode != http.StatusNoContent {
		return translateStatusCode(resp.StatusCode, "expanding disks for managed cluster", resp.Body)
	}

	return nil
}
