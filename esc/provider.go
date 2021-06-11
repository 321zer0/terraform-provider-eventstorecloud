package esc

import (
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/EventStore/terraform-provider-eventstorecloud/client"
)

var defaultTokenStore = filepath.Join(os.Getenv("HOME"), ".esctf", "tokens")

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ESC_URL", "https://api.eventstore.cloud"),
			},

			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ESC_TOKEN", ""),
			},

			"organization_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ESC_ORG_ID", ""),
			},

			"token_store": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ESC_TOKEN_STORE", defaultTokenStore),
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"eventstorecloud_project": dataSourceProject(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"eventstorecloud_project":          resourceProject(),
			"eventstorecloud_network":          resourceNetwork(),
			"eventstorecloud_peering":          resourcePeering(),
			"eventstorecloud_managed_cluster":  resourceManagedCluster(),
			"eventstorecloud_scheduled_backup": resourceScheduledBackup(),
			"eventstorecloud_integration":      resourceIntegration(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := &client.Config{
		URL:          d.Get("url").(string),
		RefreshToken: d.Get("token").(string),
		TokenStore:   d.Get("token_store").(string),
	}

	c, err := client.New(config)
	if err != nil {
		return nil, err
	}

	return &providerContext{
		organizationId: d.Get("organization_id").(string),
		client:         c,
	}, nil
}

type providerContext struct {
	organizationId string
	client         *client.Client
}

// Networks, Peerings
var fastResourceTimeout = &schema.ResourceTimeout{
	Create: schema.DefaultTimeout(3 * time.Minute),
	Read:   schema.DefaultTimeout(30 * time.Second),
	Update: schema.DefaultTimeout(3 * time.Minute),
	Delete: schema.DefaultTimeout(3 * time.Minute),
}

// Clusters
var slowResourceTimeout = &schema.ResourceTimeout{
	Create: schema.DefaultTimeout(10 * time.Minute),
	Read:   schema.DefaultTimeout(30 * time.Second),
	Update: schema.DefaultTimeout(10 * time.Minute),
	Delete: schema.DefaultTimeout(10 * time.Minute),
}

var validProviders = []string{"aws", "gcp", "azure"}
var validServerVersions = []string{"20.6", "20.10"}
var validTopologies = []string{"single-node", "three-node-multi-zone"}
var validInstanceTypes = []string{"F1", "C4", "M8", "M16", "M32", "M64", "M128"}
var validDiskTypes = []string{"GP2", "SSD", "PREMIUM-SSD-LRS"}
var validProjectionLevels = []string{"off", "system", "user"}
