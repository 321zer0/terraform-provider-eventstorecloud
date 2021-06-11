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
		Update: resourceIntegrationUpdate,

		Schema: map[string]*schema.Schema{
			"description": {
				Description: "Human readable description of the integration",
				Required:    true,
				ForceNew:    false,
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
				ForceNew:    false,
				Type:        schema.TypeMap,
			},
		},
	}
}

type ModifyMapArgs struct {
	RenameNames []struct{ from, to string }
	RemoveNames []string
}

func modifyMap(args ModifyMapArgs, data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range data {
		result[k] = v
	}

	for _, e := range args.RenameNames {
		if value, ok := result[e.from]; ok {
			result[e.to] = value
			delete(result, e.from)
		}
	}
	for _, name := range args.RemoveNames {
		if _, ok := result[name]; ok {
			delete(result, name)
		}
	}
	return result
}

func translateTfDataToApi(data map[string]interface{}) map[string]interface{} {
	return modifyMap(ModifyMapArgs{
		RenameNames: []struct{ from, to string }{
			{from: "api_key", to: "apiKey"},
			{from: "channel_id", to: "channelId"},
		},
		RemoveNames: []string{},
	}, data)
}

func translateApiDataToTf(data map[string]interface{}) map[string]interface{} {
	// We rename the read only fields the API returns on GET back to their
	// writable counterparts seen in the POST call.
	// Allowing them to be different here violates terraform's constructs and
	// makes them impossible to retrieve individually, although oddly enough
	// you can see them if you set the entire "data" map to an output variable.

	return modifyMap(ModifyMapArgs{
		RenameNames: []struct{ from, to string }{
			{from: "channelId", to: "channel_id"},
		},
		RemoveNames: []string{
			"apiKeyDisplay",
			"tokenDisplay",
		},
	}, data)
}

func resourceIntegrationCreate(d *schema.ResourceData, meta interface{}) error {
	log.Println("[BESPIN] in the hood creating up good!")

	log.Printf("[BESPIN-q] resourceIntegrationCreate data=%q\n", d.Get("data"))

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

	log.Printf("[BESPIN-q] resourceIntegrationExists data=%q\n", d.Get("data"))
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

	log.Printf("[BESPIN-q] resourceIntegrationRead data=%q\n", d.Get("data"))

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

	log.Printf("[BESPIN-q] resourceIntegrationDelete data=%q\n", d.Get("data"))

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

func resourceIntegrationUpdate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*providerContext)

	log.Printf("[BESPIN-q] resourceIntegrationUpdate data=%q\n", d.Get("data"))
	log.Println("[BESPIN-z] hiiii")

	if !d.HasChanges("description", "data") {
		log.Println("[BESPIN-z] nopeys!")
		return nil
	}

	var desc *string
	if !d.HasChange("description") {
		log.Printf("[BESPIN-z] ooh I'm digging that new DESCRIPTION!! :%q\n", d.Get("description"))
		newDesc := d.Get("description").(string)
		desc = &newDesc
	} else {
		desc = nil
	}

	var data *map[string]interface{}
	if d.HasChange("data") {
		switch v := d.Get("data").(type) {
		case nil:
			data = nil
		case map[string]interface{}:
			newData := translateTfDataToApi(v)
			log.Println("[BESPIN-z] ooh I'm digging that new DATA!!!!!!")
			data = &newData
		default:
			log.Println("[BESPIN-z] what is DATA?!!!")
		}
	} else {
		data = nil
	}

	request := client.UpdateIntegrationRequest{
		Data:        data,
		Description: desc,
	}

	orgId := c.organizationId
	log.Printf("[BESPIN-z] getting project...%q\n", d.Get("project_id"))
	projectId := d.Get("project_id").(string)
	log.Printf("[BESPIN-z] getting id...%q\n", d.Id())
	integrationId := d.Id()

	log.Println("[BESPIN-z] calling UPDATE INTEGTIUEWDF")
	return c.client.UpdateIntegration(context.Background(), orgId, projectId, integrationId, request)
}
