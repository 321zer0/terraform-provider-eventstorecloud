package esc

import (
	"context"
	"fmt"
	"log"
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

func swapMapNames(names []struct{ from, to string }, data map[string]interface{}) map[string]interface{} {
	for _, e := range names {
		if value, ok := data[e.from]; ok {
			data[e.to] = value
			delete(data, e.from)
		}
	}
	return data
}

func translateTfDataToApi(data map[string]interface{}) map[string]interface{} {
	values := []struct{ from, to string }{
		{from: "api_key", to: "apiKey"},
		{from: "channel_id", to: "channelId"},
	}
	return swapMapNames(values, data)
}

func translateApiDataToTf(data map[string]interface{}) map[string]interface{} {
	// We rename the read only fields the API returns on GET back to their
	// writable counterparts seen in the POST call.
	// Allowing them to be different here violates terraform's constructs and
	// makes them impossible to retrieve individually, although oddly enough
	// you can see them if you set the entire "data" map to an output variable.
	values := []struct{ from, to string }{
		{from: "apiKeyDisplay", to: "api_key"},
		{from: "channelId", to: "channel_id"},
		{from: "tokenDisplay", to: "token"},
	}
	return swapMapNames(values, data)
}

func resourceIntegrationCreate(d *schema.ResourceData, meta interface{}) error {
	log.Println("[BESPIN] in the hood creating up good!")

	c := meta.(*providerContext)

	projectId := d.Get("project_id").(string)

	log.Println("[BESPIN] 2")
	request := client.CreateIntegrationRequest{
		Data:        translateTfDataToApi(d.Get("data").(map[string]interface{})),
		Description: d.Get("description").(string),
	}

	log.Println("[BESPIN] 3")
	resp, err := c.client.CreateIntegration(context.Background(), c.organizationId, projectId, request)

	log.Println("[BESPIN] 4")
	if err != nil {
		log.Println("[BESPIN] 5")
		return err
	}

	log.Println("[BESPIN] 6")
	d.SetId(resp.Id)

	log.Println("[BESPIN] 7")
	return resourceIntegrationRead(d, meta)
}

func resourceIntegrationExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	c := meta.(*providerContext)

	projectId := d.Get("project_id").(string)
	integrationId := d.Id()

	integration, err := c.client.GetIntegration(context.Background(), c.organizationId, projectId, integrationId)
	if err != nil {
		return false, nil
	}
	if integration.Integration.Status == client.StateDeleted {
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
	log.Printf("[BESPIN-get convert] 1 data=%q\n", resp.Integration.Data)
	log.Printf("[BESPIN-get convert] 1 data 2=%q\n", translateApiDataToTf(resp.Integration.Data))
	if err := d.Set("data", translateApiDataToTf(resp.Integration.Data)); err != nil {
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
