package esc

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/EventStore/terraform-provider-eventstorecloud/client"
)

func resourceManagedCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Manages EventStoreDB instances and clusters in Event Store Cloud",

		CreateContext: resourceManagedClusterCreate,
		ReadContext:   resourceManagedClusterRead,
		UpdateContext: resourceManagedClusterUpdate,
		DeleteContext: resourceManagedClusterDelete,

		CustomizeDiff: resourceManagedClusterCustomizeDiff,

		Importer: &schema.ResourceImporter{
			StateContext: resourceImport,
		},

		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project in which the managed cluster exists",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"network_id": {
				Description: "ID of the network in which the managed cluster exists",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"name": {
				Description: "Name of the managed cluster",
				Required:    true,
				ForceNew:    false,
				Type:        schema.TypeString,
			},
			"topology": {
				Description:  "Topology of the managed cluster (`single-node` or `three-node-multi-zone`)",
				Required:     true,
				ForceNew:     true,
				Type:         schema.TypeString,
				ValidateFunc: ValidateWithByPass(validation.StringInSlice(validTopologies, true)),
			},
			"instance_type": {
				Description:  "Instance type of the managed cluster (find the list of valid values below)",
				Required:     true,
				ForceNew:     true,
				Type:         schema.TypeString,
				ValidateFunc: ValidateWithByPass(validation.StringInSlice(validInstanceTypes, true)),
				StateFunc: func(val interface{}) string {
					// Normalize to lower case
					return strings.ToLower(val.(string))
				},
			},
			"disk_size": {
				Description:  "Size of the data disks, in gigabytes",
				Required:     true,
				Type:         schema.TypeInt,
				ValidateFunc: ValidateWithByPass(validation.IntBetween(8, 4096)),
			},
			"disk_type": {
				Description:  "Storage class of the data disks (find the list of valid values below)",
				Required:     true,
				ForceNew:     false,
				Type:         schema.TypeString,
				ValidateFunc: ValidateWithByPass(validation.StringInSlice(validDiskTypes, true)),
				StateFunc: func(val interface{}) string {
					// Normalize to lower case
					return strings.ToLower(val.(string))
				},
			},
			"disk_iops": {
				Description: "Number of IOPS for storage, required if disk_type is `gp3`",
				Optional:    true,
				Type:        schema.TypeInt,
			},
			"disk_throughput": {
				Description: "Throughput in MB/s for storage, required if disk_type is `gp3`",
				Optional:    true,
				Type:        schema.TypeInt,
			},
			"server_version": {
				Description: "Server version to provision (find the list of valid values below)",
				Required:    true,
				ForceNew:    false,
				Type:        schema.TypeString,
				StateFunc: func(val interface{}) string {
					// Normalize to lower case
					return strings.ToLower(val.(string))
				},
			},
			"server_version_tag": {
				Description: "Server version tag to provision (find the list of valid values below). A higher server_version_tag will prompt an upgrade.",
				Optional:    true,
				ForceNew:    false,
				Computed:    true,
				Type:        schema.TypeString,
				StateFunc: func(val interface{}) string {
					// Normalize to lower case
					return strings.ToLower(val.(string))
				},
			},
			"projection_level": {
				Description:  "Determines whether to run no projections, system projections only, or system and user projections (find the list of valid values below)",
				Optional:     true,
				ForceNew:     true,
				Default:      "off",
				Type:         schema.TypeString,
				ValidateFunc: ValidateWithByPass(validation.StringInSlice(validProjectionLevels, true)),
				StateFunc: func(val interface{}) string {
					// Normalize to lower case
					return strings.ToLower(val.(string))
				},
			},
			"resource_provider": {
				Description: "Provider in which the cluster was created. Determined by the provider of the Network.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"region": {
				Description: "Region in which the cluster was created. Determined by the region of the Network",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"dns_name": {
				Description: "DNS address of the cluster",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"protected": {
				Description: "Protection from an accidental cluster deletion",
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
			},
		},
	}
}

func resourceManagedClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*providerContext)

	projectId := d.Get("project_id").(string)

	request := &client.CreateManagedClusterRequest{
		OrganizationID:  c.organizationId,
		ProjectID:       projectId,
		NetworkId:       d.Get("network_id").(string),
		Name:            d.Get("name").(string),
		Topology:        strings.ToLower(d.Get("topology").(string)),
		InstanceType:    strings.ToLower(d.Get("instance_type").(string)),
		DiskSizeGB:      int32(d.Get("disk_size").(int)),
		DiskType:        strings.ToLower(d.Get("disk_type").(string)),
		DiskIops:        int32(d.Get("disk_iops").(int)),
		DiskThroughput:  int32(d.Get("disk_throughput").(int)),
		ServerVersion:   strings.ToLower(d.Get("server_version").(string)),
		ProjectionLevel: strings.ToLower(d.Get("projection_level").(string)),
		Protected:       d.Get("protected").(bool),
	}

	resp, err := c.client.ManagedClusterCreate(ctx, request)
	if err != nil {
		return err
	}

	d.SetId(resp.ClusterID)

	if err := c.client.ManagedClusterWaitForState(ctx, &client.WaitForManagedClusterStateRequest{
		OrganizationID: c.organizationId,
		ProjectID:      projectId,
		ClusterID:      resp.ClusterID,
		State:          "available",
	}); err != nil {
		return err
	}

	return resourceManagedClusterRead(ctx, d, meta)
}

func resourceManagedClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*providerContext)

	var diags diag.Diagnostics

	projectId := d.Get("project_id").(string)
	clusterId := d.Id()

	request := &client.GetManagedClusterRequest{
		OrganizationID: c.organizationId,
		ProjectID:      projectId,
		ClusterID:      clusterId,
	}

	resp, err := c.client.ManagedClusterGet(ctx, request)
	if err != nil {
		return diag.Errorf("Internal Server Error, try again later")
	}

	if resp.ManagedCluster.Status == client.StateDeleted {
		d.SetId("")
		return nil
	}

	if err := d.Set("project_id", resp.ManagedCluster.ProjectID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("network_id", resp.ManagedCluster.NetworkID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("name", resp.ManagedCluster.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("topology", resp.ManagedCluster.Topology); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("instance_type", resp.ManagedCluster.InstanceType); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("disk_size", int(resp.ManagedCluster.DiskSizeGB)); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("disk_type", resp.ManagedCluster.DiskType); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("disk_iops", resp.ManagedCluster.DiskIops); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("disk_throughput", resp.ManagedCluster.DiskThroughput); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("server_version", resp.ManagedCluster.ServerVersion); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("server_version_tag", resp.ManagedCluster.ServerVersionTag); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("projection_level", resp.ManagedCluster.ProjectionLevel); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("resource_provider", resp.ManagedCluster.Provider); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("region", resp.ManagedCluster.Region); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("dns_name", fmt.Sprintf("%s.mesdb.eventstore.cloud", resp.ManagedCluster.ClusterID)); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("protected", resp.ManagedCluster.Protected); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceManagedClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*providerContext)

	projectId := d.Get("project_id").(string)
	clusterId := d.Id()

	if d.HasChange("name") || d.HasChange("protected") {
		request := &client.ManagedClusterUpdateRequest{
			OrganizationID: c.organizationId,
			ProjectID:      projectId,
			ClusterID:      clusterId,
			Description:    d.Get("name").(string),
			Protected:      d.Get("protected").(bool),
		}

		if err := c.client.ManagedClusterUpdate(ctx, request); err != nil {
			return err
		}
	}

	serverVersionTag, serverVersionTagFound := d.GetOk("server_version_tag")

	if serverVersionTagFound && d.HasChange("server_version_tag") {

		serverVersion := d.Get("server_version").(string)

		if !strings.HasPrefix(serverVersionTag.(string), serverVersion) {
			return diag.FromErr(fmt.Errorf("invalid server_version_tag: tag \"%s\" must begin with version \"%s\"", serverVersionTag.(string), serverVersion))

		}

		request := &client.ManagedClusterUpgradeRequest{
			OrganizationID: c.organizationId,
			ProjectID:      projectId,
			ClusterID:      clusterId,
			TargetTag:      serverVersionTag.(string),
		}
		if err := c.client.ManagedClusterUpgrade(ctx, request); err != nil {
			return err
		}
		if err := c.client.ManagedClusterWaitForState(ctx, &client.WaitForManagedClusterStateRequest{
			OrganizationID: c.organizationId,
			ProjectID:      projectId,
			ClusterID:      clusterId,
			State:          "available",
		}); err != nil {
			return err
		}
	}

	if d.HasChange("disk_size") || d.HasChange("disk_type") || d.HasChange("disk_iops") || d.HasChange("disk_throughput") {
		oldI, newI := d.GetChange("disk_size")
		oldSize, newSize := oldI.(int), newI.(int)
		if newSize < oldSize {
			return diag.FromErr(fmt.Errorf("Disks cannot be made smaller - must be %dGB or larger.", oldSize))
		}

		request := &client.ExpandManagedClusterDiskRequest{
			OrganizationID: c.organizationId,
			ProjectID:      projectId,
			ClusterID:      clusterId,
			DiskIops:       int32(d.Get("disk_iops").(int)),
			DiskSizeGB:     int32(newSize),
			DiskThroughput: int32(d.Get("disk_throughput").(int)),
			DiskType:       d.Get("disk_type").(string),
		}
		if err := c.client.ManagedClusterExpandDisk(ctx, request); err != nil {
			return err
		}

		if err := c.client.ManagedClusterWaitForState(ctx, &client.WaitForManagedClusterStateRequest{
			OrganizationID: c.organizationId,
			ProjectID:      projectId,
			ClusterID:      clusterId,
			State:          "available",
		}); err != nil {
			return err
		}
	}

	return resourceManagedClusterRead(ctx, d, meta)
}

func resourceManagedClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*providerContext)

	projectId := d.Get("project_id").(string)
	clusterId := d.Id()

	request := &client.DeleteManagedClusterRequest{
		OrganizationID: c.organizationId,
		ProjectID:      projectId,
		ClusterID:      clusterId,
	}

	if err := c.client.ManagedClusterDelete(ctx, request); err != nil {
		return err
	}

	return c.client.ManagedClusterWaitForState(ctx, &client.WaitForManagedClusterStateRequest{
		OrganizationID: c.organizationId,
		ProjectID:      projectId,
		ClusterID:      clusterId,
		State:          "deleted",
	})
}

func resourceManagedClusterCustomizeDiff(ctx context.Context, diff *schema.ResourceDiff, v interface{}) error {

	disk_type := diff.Get("disk_type").(string)
	disk_iops := diff.Get("disk_iops").(int)
	disk_throughput := diff.Get("disk_throughput").(int)

	switch disk_type {
	case "GP3", "gp3":
		if disk_iops == 0 {
			return fmt.Errorf("'iops' must be set when 'type' is '%s'", disk_type)
		}
		if disk_throughput == 0 {
			return fmt.Errorf("'throughput' must be set when 'type' is '%s'", disk_type)
		}
		if disk_iops < 3000 || disk_iops > 16000 {
			return fmt.Errorf("'iops' must be set between 3000 and 16000")
		}
		if disk_throughput < 125 || disk_throughput > 1000 {
			return fmt.Errorf("'throughput' must be set between 125 and 1000")
		}
	default:
		if disk_iops != 0 {
			return fmt.Errorf("'iops' must not be set when 'type' is '%s'", disk_type)
		}

		if disk_throughput != 0 {
			return fmt.Errorf("'throughput' must not be set when 'type' is '%s'", disk_type)
		}
	}

	return nil
}
