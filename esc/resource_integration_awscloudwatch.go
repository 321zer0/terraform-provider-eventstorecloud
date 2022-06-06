package esc

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/EventStore/terraform-provider-eventstorecloud/client"
)

func resourceIntegrationAwsCloudWatchAwsCloudWatch() *schema.Resource {

	return &schema.Resource{
		CreateContext: resourceIntegrationAwsCloudWatchCreate,
		ReadContext:   resourceIntegrationAwsCloudWatchRead,
		DeleteContext: resourceIntegrationAwsCloudWatchDelete,
		UpdateContext: resourceIntegrationAwsCloudWatchUpdate,

		Description: "Manages an integration for type AwsCloudWatch",

		Importer: &schema.ResourceImporter{
			StateContext: resourceImport,
		},

		Schema: map[string]*schema.Schema{
			"access_key_id": {
				Description: "AWS IAM access key",
				Required:    false,
				ForceNew:    false,
				Optional:    true,
				Sensitive:   true,
				Type:        schema.TypeString,
			},
			"description": {
				Description: "Human readable description of the integration",
				Required:    true,
				ForceNew:    false,
				Type:        schema.TypeString,
			},
			"group_name": {
				Description: "Name of the CloudWatch group",
				Required:    true,
				ForceNew:    false,
				Sensitive:   false,
				Type:        schema.TypeString,
			},
			"project_id": {
				Description: "ID of the project to which the integration applies",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"region": {
				Description: "AWS region for group",
				Required:    true,
				ForceNew:    false,
				Sensitive:   false,
				Type:        schema.TypeString,
			},
			"source": {
				Description: "Name of the CloudWatch group",
				Required:    true,
				ForceNew:    true,
				Sensitive:   false,
				Type:        schema.TypeString,
			},
			"secret_access_key": {
				Description: "AWS IAM secret access key",
				Required:    false,
				ForceNew:    false,
				Optional:    true,
				Sensitive:   true,
				Type:        schema.TypeString,
			},
		},
	}
}

func resourceIntegrationAwsCloudWatchCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*providerContext)

	projectId := d.Get("project_id").(string)

	accessKeyIdVal := d.Get("access_key_id")
	secretAccessKeyVal := d.Get("secret_access_key")
	if accessKeyIdVal == nil || secretAccessKeyVal == nil {
		var diags diag.Diagnostics
		if accessKeyIdVal == nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Missing access_key_id.",
			})
		}
		if secretAccessKeyVal == nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Missing secret_access_key.",
			})
		}
		return diags
	}

	data := map[string]interface{}{
		"accessKeyId":     accessKeyIdVal.(string),
		"groupName":       d.Get("group_name").(string),
		"region":          d.Get("region").(string),
		"secretAccessKey": secretAccessKeyVal.(string),
		"source":          d.Get("source").(string),
		"sink":            "awsCloudWatch",
	}
	request := client.CreateIntegrationRequest{
		Data:        data,
		Description: d.Get("description").(string),
	}

	resp, err := c.client.CreateIntegration(ctx, c.organizationId, projectId, request)
	if err != nil {
		return err
	}

	d.SetId(resp.Id)

	return resourceIntegrationAwsCloudWatchRead(ctx, d, meta)
}

func resourceIntegrationAwsCloudWatchRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*providerContext)

	var diags diag.Diagnostics

	projectId := d.Get("project_id").(string)
	integrationId := d.Id()

	resp, err := c.client.GetIntegration(ctx, c.organizationId, projectId, integrationId)
	if err != nil || resp.Integration.Status == client.StateDeleted {
		d.SetId("")
		return diags
	}

	if err := d.Set("description", resp.Integration.Description); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("project_id", resp.Integration.ProjectId); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	setVal := func(tfName, dataKey string) {
		if val, ok := resp.Integration.Data[dataKey]; ok {
			if err := d.Set(tfName, val); err != nil {
				diags = append(diags, diag.FromErr(err)...)
			}
		} else {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("returned API data value missing key %q", dataKey),
			})
		}
	}
	setVal("group_name", "groupName")
	setVal("region", "region")
	setVal("source", "source")

	return diags
}

func resourceIntegrationAwsCloudWatchDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*providerContext)

	var diags diag.Diagnostics

	projectId := d.Get("project_id").(string)
	integrationId := d.Id()

	if err := c.client.DeleteIntegration(ctx, c.organizationId, projectId, integrationId); err != nil {
		return err
	}

	start := time.Now()
	for {
		resp, err := c.client.GetIntegration(ctx, c.organizationId, projectId, integrationId)
		if err != nil {
			return diag.Errorf("error polling integration %q (%q) to see if it actually got deleted", integrationId, d.Get("description"))
		}
		elapsed := time.Since(start)
		if elapsed.Seconds() > 30.0 {
			return diag.Errorf("integration %q (%q) does not seem to be deleting", integrationId, d.Get("description"))
		}
		if resp.Integration.Status == "deleted" {
			return diags
		}
		time.Sleep(1.0)
	}
}

func resourceIntegrationAwsCloudWatchUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*providerContext)

	if !d.HasChanges("description", "group_name", "region", "access_key_id", "secret_access_key") {
		return resourceIntegrationAwsCloudWatchRead(ctx, d, meta)
	}

	var desc *string
	desc = nil
	newDesc := d.Get("description").(string)
	if newDesc != "" {
		desc = &newDesc
	}

	data := map[string]interface{}{
		"groupName": d.Get("group_name").(string),
		"source":    d.Get("source").(string),
		"sink":      "awsCloudWatch",
		"region":    d.Get("region").(string),
	}
	if d.HasChange("access_key_id") {
		newVal := d.Get("access_key_id").(string)
		if newVal != "" {
			data["accessKeyId"] = newVal
		}
	}
	if d.HasChange("secret_access_key") {
		newVal := d.Get("secret_access_key").(string)
		if newVal != "" {
			data["secretAccessKey"] = newVal
		}
	}

	request := client.UpdateIntegrationRequest{
		Data:        &data,
		Description: desc,
	}

	orgId := c.organizationId
	projectId := d.Get("project_id").(string)
	integrationId := d.Id()

	if err := c.client.UpdateIntegration(ctx, orgId, projectId, integrationId, request); err != nil {
		return err
	}

	return resourceIntegrationAwsCloudWatchRead(ctx, d, meta)
}
