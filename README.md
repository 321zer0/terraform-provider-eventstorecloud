# Terraform Provider for Event Store Cloud

This repository contains a [Terraform][terraform] provider for provisioning resources in [Event Store Cloud][esc].

# Installation

## Terraform 0.12

We provide binary releases for macOS, Windows and Linux via GitHub releases. In order for Terraform to find the plugin, the appropriate binary must be placed into the Terraform third-party plugin directory, the location of which varies by operating system:

- `%APPDATA%\terraform.d\plugins` on Windows
- `~/.terraform.d/plugins` on macOS or Linux

Alternatively, the binary can be placed alongside the main `terraform` binary.

On macOS and Linux, you can download the provider using the following commands:

- macOS: `curl -o ./terraform-provider-eventstorecloud.zip -L https://github.com/EventStore/terraform-provider-eventstorecloud/releases/download/v1.5.3/terraform-provider-eventstorecloud_1.5.3_darwin_amd64.zip && unzip ./terraform-provider-eventstorecloud.zip && mv ./terraform-provider-eventstorecloud_v1.5.3 ~/.terraform.d/plugins/terraform-provider-eventstorecloud_v1.5.3`
- Linux: `curl -o ./terraform-provider-eventstorecloud.zip -L https://github.com/EventStore/terraform-provider-eventstorecloud/releases/download/v1.5.3/terraform-provider-eventstorecloud_1.5.3_linux_amd64.zip && unzip ./terraform-provider-eventstorecloud.zip && mv ./terraform-provider-eventstorecloud_v1.5.3 ~/.terraform.d/plugins/terraform-provider-eventstorecloud_v1.5.3`

If you prefer to install from source, use the `make install` target in this repository. You'll need a Go 1.13+
development environment.

## Terraform 0.13+

Terraform now supports third party modules installed via the plugin registry. Add the following to your terraform module
configuration.

```
terraform {
  required_providers {
    eventstorecloud = {
      source = "EventStore/eventstorecloud"
      version = "1.5.3"
    }
  }
}
```

# Provider Configuration

The Event Store Cloud provider must be configured with an access token, however there are several additional
options which may be useful.

Provider configuration options are:

- `token` - (`ESC_TOKEN` via the environment) - *Required* - a refresh token for Event Store Cloud. This token can be created and displayed with the esc cli tool [esc cli](https://github.com/EventStore/esc), or via the "request refresh token" button on the [Authentification Tokens page](https://console.eventstore.cloud/authentication-tokens)  in the console. The token id displayed in the cloud console is not a valid token.
- `organization_id` - (`ESC_ORG_ID` via the environment) - *Required* - the identifier of the Event Store Cloud
  organization into which to provision resources.

- `url` - (`ESC_URL` via the environment) - *Optional* - the URL of the Event Store Cloud API. This defaults
  to the public cloud instance of Event Store Cloud, but may be overridden to provision resources in another
  instance.
- `token_store` - (`ESC_TOKEN_STORE` via the environment) - *Optional* - the location on the local filesystem
  of the token cache. This is shared with the Event Store Cloud CLI.

# Resources

All resources in Event Store Cloud can be provisioned using the Terraform provider, and existing projects
can be queried using a data source in the provider.

## Data Source `eventstorecloud_project`

Looks up a project in the organization with which the provider is configured by name.

### Example

```hcl
# This assumes a project with the name "Example Project" exists
data "eventstorecloud_project" "example" {
  name = "Example Project"
}

output "project_id" {
  value = data.eventstorecloud_project.example.id
}
```

### Arguments

- `name` - (`string`, Required) - the name of the project.

### Attributes

As well as the input arguments, the following properties are exposed:

- `id` - (`string`) - the ID of the project.


## Resource `eventstorecloud_project`

Creates a project in the organization with which the provider is configured.

### Example

```hcl
resource "eventstorecloud_project" "example" {
  name = "Example Project"
}
```

### Arguments

- `name` - (`string`, Required) - the name of the project.

### Attributes

As well as the input arguments, the following properties are exposed:

- `id` - (`string`) - the ID of the project.


## Resource `eventstorecloud_network`

Creates a new network in a cloud project. Networks are specific to a given cloud provider and region.

### Example

```hcl
resource "eventstorecloud_project" "example" {
  name = "Example Project"
}

resource "eventstorecloud_network" "example" {
  name = "Example Network"

  project_id = data.eventstorecloud_project.example.id

  resource_provider = "aws"
  region = "us-west-2"
  cidr_block = "172.21.0.0/16"
}
```

### Arguments

- `name` - (`string`, Required) - the name of the network.
- `project_id` - (`string`, Required) - the ID of the project in which the network should be created.
- `resource_provider` - (`string`, Required) - the name of the cloud in which the network should be created. Valid
  values for `resource_provider` are `aws`, `gcp` and `azure`.
- `region` - (`string`, Required) - the name of the region in which the network should be created. Region names must
  be in the format used by the resource provider, for example `us-west-2` for AWS, or `East US` for Azure.
- `cidr_block` - (`string`, Required) - the address space of the network. The maximum prefix length is `/9` however what
  is allowed is provider dependent, and the minimum is `/24`. Smaller networks can hold fewer managed clusters, but may 
  be easier to peer to infrastructure hosting applications.

### Attributes

As well as the input arguments, the following properties are exposed:

- `id` - (`string`) - the ID of the network.

## Resource `eventstorecloud_peering`

Creates a new peering in a cloud network. Peerings can be created from customer-managed networks in the same
cloud and region as the network with which the peering exists. A number of arguments have a format dependent
on the cloud provider.

### Example (AWS)

```hcl
resource "eventstorecloud_project" "example" {
  name = "Example Project"
}

resource "eventstorecloud_network" "example" {
  name = "Example Network"

  project_id = data.eventstorecloud_project.example.id

  resource_provider = "aws"
  region = "us-west-2"
  cidr_block = "172.21.0.0/16"
}

resource "eventstorecloud_peering" "example" {
	name = "Peering from AWS into Example Network"

	project_id = eventstorecloud_network.example.project_id
	network_id = eventstorecloud_network.example.id

	peer_resource_provider = eventstorecloud_network.example.resource_provider
	peer_network_region = eventstorecloud_network.example.region

	peer_account_id = "<Customer AWS Account ID>"
	peer_network_id = "<Customer VPC ID>"
	routes = ["<Address space of the customer VPC>"]
}
```

### Arguments

- `name` - (`string`, Required) - the name of the peering.
- `project_id` - (`string`, Required) - the ID of the project in which the peering should be created.
- `network_id` - (`string`, Required) - the ID of the network into which the peering should be created.
- `peer_resource_provider` - (`string`, Required) - the name of the cloud in which the customer managed network exists.
  Currently this must be the same as the resource provider of the Event Store Cloud network.
- `peer_network_region` - (`string`, Required) - the name of the region in which the customer managed network exists.
  Currently this must be the same as the region of the Event Store Cloud network, and specified in the format used by the
  customer cloud - for example `us-west-2` in AWS, and `westus2` for Azure.
- `peer_account_id` - (`string`, Required) - the account identifier for the account in which the customer-managed network
  exists. The required format is dependent on the target cloud:
    - AWS: The Account ID (12 digit numeric value available from the account drop down in the AWS console)
- `peer_network_id` - (`string`, Required) - the network identifier for the customer-managed network. The required format
  is dependent on the target cloud:
    - AWS: The VPC ID (`vpc-XXXXXXXX`, available via the VPC section of the AWS console)
- `routes` - (`string`, Required) - CIDR Blocks representing routes to be created from the Event Store Cloud network into
  the customer-managed network. Typically this consists of one element, the address space of the customer-managed network.

### Attributes

As well as the input arguments, the following properties are exposed:

- `id` - (`string`) - the ID of the peering.
- `provider_metadata` - metadata supplied by the resource provider cloud about the peering link:
   - `aws_peering_link_id` - (`string`) - the ID of the peering link in AWS. Empty if the resource provider is not AWS.
   - `gcp_project_id` - (`string`) - the project ID of the peering link in GCP. Empty if the resource provider is not GCP.
   - `gcp_network_name` - (`string`) - the network name for the peering link in GCP. Empty if the resource provider is not GCP.
   - `gcp_network_id` - (`string`) - GCP Network ID in URL format which can be passed to `google_compute_network_peering` resources. Empty if the peering Provider is not GCP.

## Resource `eventstorecloud_managed_cluster`

Creates a new Managed Event StoreDB cluster in a network.

### Example

```hcl
data "eventstorecloud_project" "example" {
  name = "Example Project"
}

resource "eventstorecloud_network" "example" {
  name = "Example Network"

  project_id = data.eventstorecloud_project.example.id

  resource_provider = "aws"
  region = "us-west-2"
  cidr_block = "172.21.0.0/16"
}

resource "eventstorecloud_managed_cluster" "example" {
	name = "Example Cluster"

	project_id = eventstorecloud_network.example.project_id
	network_id = eventstorecloud_network.example.id

	topology = "three-node-multi-zone"
	instance_type = "F1"
	disk_size = 24
	disk_type = "gp2"
	server_version = "20.6"
}
```

### Arguments

- `name` - (`string`, Required) - the name of the managed cluster.
- `project_id` - (`string`, Required) - the ID of the project in which the managed cluster should be created.
- `network_id` - (`string`, Required) - the ID of the network in which the managed cluster should be created. This
  determines which cloud and region the managed cluster will be hosted in.
- `topology` - (`string`, Required) - the topology of the managed cluster. This determines the fault tolerance of
  the cluster. Valid values are `single-node` and `three-node-multi-zone`. The actual implementation of each topology
  is specific to the resource provider.
- `instance_type` - (`string`, Required) - the size of the instances to use in the managed cluster. This determines
  the performance of the cluster. Valid values are `F1` and `C4`.
- `disk_size` - (`int`, Required) - the size of the data disks in gigabytes. This determines how much data can be
  stored by the cluster. The minimum size is 8GB for a cluster in AWS and 10GB for a cluster in GCP.
- `disk_type` - (`string`, Required) - the type of data disks, which helps determine the performance profile of the
  cluster. Currently `gp2` is the only option.
- `server_version` - (`string`, Required) - the version of Event Store Server with which the cluster is provisioned.
  Currently `20.6` and `20.10` are available.
- `projection_level` - (`string`, Optional) - the mode in which to enable projections. Valid values are `off`, `system`
  and `user`. Defaults to `off`.

### Attributes

As well as the input arguments, the following properties are exposed:

- `id` - (`string`) - the ID of the network.
- `dns_name` (`string`) - the DNS name at which the cluster can be found from networks peered into cluster network.
- `resource_provider` (`string`) - the resource provider into which the cluster was provisioned. This is controlled by
  the network in which the cluster is created.
- `region` (`string`) - the region in which the cluster was provisioned. This is controlled by the network in which
  the cluster is created.

## Resource `eventstorecloud_scheduled_backup`

Creates a new scheduled backup.

### Example

```hcl
resource "eventstorecloud_scheduled_backup" "daily" {
    project_id = eventstorecloud_project.example.id
    schedule = "0 12 * * */1"
    description = "Creates a backup once a day at 12:00"
    
    source_cluster_id = eventstorecloud_managed_cluster.example.id
    backup_description = "{cluster} Daily Backup {datetime:RFC3339}"
    max_backup_count = 3
}
```

### Arguments

- `description` - (`string`, Required) - a description of the backup
- `project_id` - (`string`, Required) - the ID of the project in which the managed cluster should be created
- `schedule` - (`string`, Required) - Schedule for the backup, defined using restricted subset of cron
- `source_cluster_id` - (`string`, Required) - the ID of the cluster to back up
- `backup_description` - (`string`, Required) - a description to give each backup. This can include template parameters, such as the time (for example, "{datetime:RFC3339}")
- `max_backup_count` - (`int`, Required) - the maximum number of backups to keep for this job
  
### Attributes

As well as the input arguments, the following properties are exposed:

- `id` - (`string`) - the ID of the created job. This may change if the job is updated, as doing so will force the job to be recreated.

## Resource `eventstorecloud_integrations`

Creates a new integration resource.

### Example

```hcl
resource "eventstorecloud_integration" "opsgenie_issues" {
    project_id = var.project_id
    description = "create OpsGenie alerts from issues"
    data = {
        sink = "opsGenie"
        api_key = "<secret OpsGenie key here>"
        source = "issues"
    }
}


resource "eventstorecloud_integration" "slack_notifications" {
    project_id = var.project_id
    description = "send Slack a message when a notification happens"
    data = {
        sink = "slack"
        token = "<secret token here>"        
        channel_id = "#esc-cluster-notifications"
        source = "notifications"
    }
}
```

### Arguments

- `description` - (`string`, Required) - a description of the integration
- `project_id` - (`string`, Required) - the ID of the project to which the integration applies. Only events coming from clusters or other resources under this project will interact with this integration
- `data` - (`map`, Required) - Defines the integration, see below

#### Data Properties

- `source` - Common to all integrations. Can be either `issues` or `notifications`. See [here](https://developers.eventstore.com/cloud/integrations/#integration-sources) for information on the types of integration sources
- `sinks` - Common to all integrations. The type of this value determines the acceptable settings for `source` as well as what other properties are allowed in this field. See [here](https://developers.eventstore.com/cloud/integrations/#integration-sinks) for information on integration sinks

##### OpsGenie Specific Data Files

OpsGenie currently only supports the "issues" source. See [here](https://developers.eventstore.com/cloud/integrations/opsgenie.html#how-to-create-an-api-key) for more info on creating OpsGenie integrations.

- `api_key` - A secret key required to integrate with OpsGenie.  *NOTE:* This value is considered private and so is NOT ever returned by the Event Store Cloud API. For that reason it is impossible to display it as a Terraform output value.

##### Slack Specific Data Files

Slack currently only supports the "issues" and "notification" sources. See [here](https://developers.eventstore.com/cloud/integrations/slack.html) for more info on creating Slack integrations

- `channel_id` - The channel in Slack that messages will be sent. See [here](https://developers.eventstore.com/cloud/integrations/slack.html) for more info. Can be retrieved as an output value.
- `token` - A secret token required to integrate with Slack. *NOTE:* This value is considered private and so is NOT ever returned by the Event Store Cloud API. For that reason it is impossible to display it as a Terraform output value.


## Contributing

The Event Store Cloud Terraform provider is released under the Mozilla Public License version 2, like most Terraform
providers. We welcome pull requests and issues! We adhere to the [Contributor Covenant][codeofconduct] Code of Conduct,
and ask that contributors familiarize themselves with it. We also have a set of [Contributing Guidelines][contributing].

[terraform]: (https://terraform.io)
[esc]: https://eventstore.com/event-store-cloud/
[codeofconduct]: https://github.com/EventStore/terraform-provider-eventstorecloud/tree/trunk/CODE-OF-CONDUCT.md
[contributing]: https://github.com/EventStore/terraform-provider-eventstorecloud/tree/trunk/CONTRIBUTING.md
