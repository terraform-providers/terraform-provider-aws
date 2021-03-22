package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/workspaces/waiter"
)

func dataSourceAwsWorkspacesDirectory() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsWorkspacesDirectoryRead,

		Schema: map[string]*schema.Schema{
			"alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"customer_user_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"directory_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_ip_addresses": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"iam_role_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_group_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"registration_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"self_service_permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"change_compute_type": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"increase_volume_size": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"rebuild_workspace": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"restart_workspace": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"switch_running_mode": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tagsSchema(),
			"workspace_access_properties": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device_type_android": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device_type_chromeos": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device_type_ios": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device_type_osx": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device_type_web": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device_type_windows": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device_type_zeroclient": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"workspace_creation_properties": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_security_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"default_ou": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enable_internet_access": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"enable_maintenance_mode": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"user_enabled_as_local_administrator": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"workspace_security_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsWorkspacesDirectoryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	directoryID := d.Get("directory_id").(string)

	rawOutput, state, err := waiter.DirectoryState(conn, directoryID)()
	if err != nil {
		return fmt.Errorf("error getting WorkSpaces Directory (%s): %w", directoryID, err)
	}
	if state == workspaces.WorkspaceDirectoryStateDeregistered {
		return fmt.Errorf("WorkSpaces directory %s was not found", directoryID)
	}

	d.SetId(directoryID)

	directory := rawOutput.(*workspaces.WorkspaceDirectory)
	d.Set("directory_id", directory.DirectoryId)
	d.Set("workspace_security_group_id", directory.WorkspaceSecurityGroupId)
	d.Set("iam_role_id", directory.IamRoleId)
	d.Set("registration_code", directory.RegistrationCode)
	d.Set("directory_name", directory.DirectoryName)
	d.Set("directory_type", directory.DirectoryType)
	d.Set("alias", directory.Alias)

	if err := d.Set("subnet_ids", flattenStringSet(directory.SubnetIds)); err != nil {
		return fmt.Errorf("error setting subnet_ids: %w", err)
	}

	if err := d.Set("self_service_permissions", flattenSelfServicePermissions(directory.SelfservicePermissions)); err != nil {
		return fmt.Errorf("error setting self_service_permissions: %w", err)
	}

	if err := d.Set("workspace_access_properties", flattenWorkspaceAccessProperties(directory.WorkspaceAccessProperties)); err != nil {
		return fmt.Errorf("error setting workspace_access_properties: %w", err)
	}

	if err := d.Set("workspace_creation_properties", flattenWorkspaceCreationProperties(directory.WorkspaceCreationProperties)); err != nil {
		return fmt.Errorf("error setting workspace_creation_properties: %w", err)
	}

	if err := d.Set("ip_group_ids", flattenStringSet(directory.IpGroupIds)); err != nil {
		return fmt.Errorf("error setting ip_group_ids: %w", err)
	}

	if err := d.Set("dns_ip_addresses", flattenStringSet(directory.DnsIpAddresses)); err != nil {
		return fmt.Errorf("error setting dns_ip_addresses: %w", err)
	}

	tags, err := keyvaluetags.WorkspacesListTags(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error listing tags: %w", err)
	}
	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}
