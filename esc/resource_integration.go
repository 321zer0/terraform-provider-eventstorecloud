package esc

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"

	"github.com/EventStore/terraform-provider-eventstorecloud/client"
)

func resourceIntegration() *schema.Resource {

	return &schema.Resource{
		Create: resourceIntegrationCreate,
		Exists: resourceIntegrationExists,
		Read:   resourceIntegrationRead,
		Delete: resourceIntegrationDelete,

		Schema: map[string]*schema.Schema{
			"description": {
				Description: "Human readable description of the integration",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"project_id": {
				Description: "ID of the project to which the integration applies",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"data": {
				Description: "Data for the integration",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeMap,
			},
		},
	}
}

func resourceIntegrationCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*providerContext)

	projectId := d.Get("project_id").(string)

	request := client.CreateIntegrationRequest{
		Data:        d.Get("data").(map[string]interface{}),
		Description: d.Get("description").(string),
	}

	resp, err := c.client.CreateIntegration(context.Background(), c.organizationId, projectId, request)

	if err != nil {
		return err
	}

	d.SetId(resp.Id)

	return resourceIntegrationRead(d, meta)
}

func resourceIntegrationExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	c := meta.(*providerContext)

	projectId := d.Get("project_id").(string)
	integrationId := d.Id()

	cluster, err := c.client.GetIntegration(context.Background(), c.organizationId, projectId, integrationId)
	if err != nil {
		return false, nil
	}
	if cluster.Integration.Status == client.StateDeleted {
		return false, nil
	}

	return true, nil
}

func resourceIntegrationRead(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*providerContext)
	projectId := d.Get("project_id").(string)
	integrationId := d.Id()

	resp, err := c.client.GetIntegration(context.Background(), c.organizationId, projectId, integrationId)
	if err != nil {
		return err
	}
	if err := d.Set("description", resp.Integration.Description); err != nil {
		return err
	}
	if err := d.Set("project_id", resp.Integration.ProjectId); err != nil {
		return err
	}
	if err := d.Set("data", resp.Integration.Data); err != nil {
		return err
	}

	return nil
}

func resourceIntegrationDelete(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*providerContext)

	projectId := d.Get("project_id").(string)
	integrationId := d.Id()

	if err := c.client.DeleteIntegration(context.Background(), c.organizationId, projectId, integrationId); err != nil {
		return err
	}

	start := time.Now()
	for {
		resp, err := c.client.GetIntegration(context.Background(), c.organizationId, projectId, integrationId)
		if err != nil {
			return fmt.Errorf("error polling integration %q (%q) to see if it actually got deleted", integrationId, d.Get("description"))
		}
		elapsed := time.Since(start)
		if elapsed.Seconds() > 30.0 {
			return errors.Errorf("integration %q (%q) does not seem to be deleting", integrationId, d.Get("description"))
		}
		if resp.Integration.Status == "deleted" {
			return nil
		}
		time.Sleep(1.0)
	}
}
