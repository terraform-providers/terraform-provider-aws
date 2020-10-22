package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/workspaces/waiter"
)

func resourceAwsWorkspacesDirectory() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWorkspacesDirectoryCreate,
		Read:   resourceAwsWorkspacesDirectoryRead,
		Update: resourceAwsWorkspacesDirectoryUpdate,
		Delete: resourceAwsWorkspacesDirectoryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
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
				ForceNew: true,
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
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"iam_role_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_group_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"registration_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"self_service_permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"change_compute_type": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"increase_volume_size": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"rebuild_workspace": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"restart_workspace": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"switch_running_mode": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tagsSchema(),
			"workspace_creation_properties": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_security_group_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"default_ou": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"enable_internet_access": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"enable_maintenance_mode": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"user_enabled_as_local_administrator": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
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

func resourceAwsWorkspacesDirectoryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn
	directoryId := d.Get("directory_id").(string)

	tags := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().WorkspacesTags()

	input := &workspaces.RegisterWorkspaceDirectoryInput{
		DirectoryId:       aws.String(directoryId),
		EnableSelfService: aws.Bool(false), // this is handled separately below
		EnableWorkDocs:    aws.Bool(false),
		Tenancy:           aws.String(workspaces.TenancyShared),
		Tags:              tags,
	}

	if v, ok := d.GetOk("subnet_ids"); ok {
		input.SubnetIds = expandStringSet(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Regestering WorkSpaces Directory...\n%#v\n", *input)
	_, err := conn.RegisterWorkspaceDirectory(input)
	if err != nil {
		return err
	}
	d.SetId(directoryId)

	log.Printf("[DEBUG] Waiting for WorkSpaces Directory %q to become registered...", directoryId)
	_, err = waiter.DirectoryRegistered(conn, directoryId)
	if err != nil {
		return fmt.Errorf("error registering directory: %s", err)
	}
	log.Printf("[DEBUG] WorkSpaces Directory %q is registered", directoryId)

	if v, ok := d.GetOk("self_service_permissions"); ok {
		log.Printf("[DEBUG] Modifying WorkSpaces Directory %q self-service permissions...", directoryId)
		_, err := conn.ModifySelfservicePermissions(&workspaces.ModifySelfservicePermissionsInput{
			ResourceId:             aws.String(directoryId),
			SelfservicePermissions: expandSelfServicePermissions(v.([]interface{})),
		})
		if err != nil {
			return fmt.Errorf("error setting self service permissions: %s", err)
		}
		log.Printf("[DEBUG] WorkSpaces Directory %q self-service permissions are set", directoryId)
	}

	if v, ok := d.GetOk("workspace_creation_properties"); ok {
		log.Printf("[DEBUG] Modifying WorkSpaces Directory %q creation properties...", directoryId)
		_, err := conn.ModifyWorkspaceCreationProperties(&workspaces.ModifyWorkspaceCreationPropertiesInput{
			ResourceId:                  aws.String(directoryId),
			WorkspaceCreationProperties: expandWorkspaceCreationProperties(v.([]interface{})),
		})
		if err != nil {
			return fmt.Errorf("error setting creation properties: %s", err)
		}
		log.Printf("[DEBUG] WorkSpaces Directory %q creation properties are set", directoryId)
	}

	return resourceAwsWorkspacesDirectoryRead(d, meta)
}

func resourceAwsWorkspacesDirectoryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	rawOutput, state, err := waiter.DirectoryState(conn, d.Id())()
	if err != nil {
		return fmt.Errorf("error getting WorkSpaces Directory (%s): %s", d.Id(), err)
	}
	if state == workspaces.WorkspaceDirectoryStateDeregistered {
		log.Printf("[WARN] WorkSpaces Directory (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	directory := rawOutput.(*workspaces.WorkspaceDirectory)
	d.Set("directory_id", directory.DirectoryId)
	if err := d.Set("subnet_ids", flattenStringSet(directory.SubnetIds)); err != nil {
		return fmt.Errorf("error setting subnet_ids: %s", err)
	}
	d.Set("workspace_security_group_id", directory.WorkspaceSecurityGroupId)
	d.Set("iam_role_id", directory.IamRoleId)
	d.Set("registration_code", directory.RegistrationCode)
	d.Set("directory_name", directory.DirectoryName)
	d.Set("directory_type", directory.DirectoryType)
	d.Set("alias", directory.Alias)

	if err := d.Set("self_service_permissions", flattenSelfServicePermissions(directory.SelfservicePermissions)); err != nil {
		return fmt.Errorf("error setting self_service_permissions: %s", err)
	}

	if err := d.Set("workspace_creation_properties", flattenWorkspaceCreationProperties(directory.WorkspaceCreationProperties)); err != nil {
		return fmt.Errorf("error setting workspace_creation_properties: %s", err)
	}

	if err := d.Set("ip_group_ids", flattenStringSet(directory.IpGroupIds)); err != nil {
		return fmt.Errorf("error setting ip_group_ids: %s", err)
	}

	if err := d.Set("dns_ip_addresses", flattenStringSet(directory.DnsIpAddresses)); err != nil {
		return fmt.Errorf("error setting dns_ip_addresses: %s", err)
	}

	tags, err := keyvaluetags.WorkspacesListTags(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error listing tags: %s", err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsWorkspacesDirectoryUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn

	if d.HasChange("self_service_permissions") {
		log.Printf("[DEBUG] Modifying WorkSpaces Directory %q self-service permissions...", d.Id())
		permissions := d.Get("self_service_permissions").([]interface{})

		_, err := conn.ModifySelfservicePermissions(&workspaces.ModifySelfservicePermissionsInput{
			ResourceId:             aws.String(d.Id()),
			SelfservicePermissions: expandSelfServicePermissions(permissions),
		})
		if err != nil {
			return fmt.Errorf("error updating self service permissions: %s", err)
		}
		log.Printf("[DEBUG] WorkSpaces Directory %q self-service permissions are set", d.Id())
	}

	if d.HasChange("workspace_creation_properties") {
		log.Printf("[DEBUG] Modifying WorkSpaces Directory %q creation properties...", d.Id())
		properties := d.Get("workspace_creation_properties").([]interface{})

		_, err := conn.ModifyWorkspaceCreationProperties(&workspaces.ModifyWorkspaceCreationPropertiesInput{
			ResourceId:                  aws.String(d.Id()),
			WorkspaceCreationProperties: expandWorkspaceCreationProperties(properties),
		})
		if err != nil {
			return fmt.Errorf("error updating creation properties: %s", err)
		}
		log.Printf("[DEBUG] WorkSpaces Directory %q creation properties are set", d.Id())
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.WorkspacesUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsWorkspacesDirectoryRead(d, meta)
}

func resourceAwsWorkspacesDirectoryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn

	err := workspacesDirectoryDelete(d.Id(), conn)
	if err != nil {
		return fmt.Errorf("error deleting WorkSpaces Directory (%s): %s", d.Id(), err)
	}

	return nil
}

func workspacesDirectoryDelete(id string, conn *workspaces.WorkSpaces) error {
	log.Printf("[DEBUG] Deregistering WorkSpaces Directory %q", id)
	_, err := conn.DeregisterWorkspaceDirectory(&workspaces.DeregisterWorkspaceDirectoryInput{
		DirectoryId: aws.String(id),
	})
	if err != nil {
		return fmt.Errorf("error deregistering WorkSpaces Directory %q: %w", id, err)
	}

	log.Printf("[DEBUG] Waiting for WorkSpaces Directory %q to be deregistered", id)
	_, err = waiter.DirectoryDeregistered(conn, id)
	if err != nil {
		return fmt.Errorf("error waiting for WorkSpaces Directory %q to be deregistered: %w", id, err)
	}
	log.Printf("[DEBUG] WorkSpaces Directory %q is deregistered", id)

	return nil
}

func expandSelfServicePermissions(permissions []interface{}) *workspaces.SelfservicePermissions {
	if len(permissions) == 0 || permissions[0] == nil {
		return nil
	}

	result := &workspaces.SelfservicePermissions{}

	p := permissions[0].(map[string]interface{})

	if p["change_compute_type"].(bool) {
		result.ChangeComputeType = aws.String(workspaces.ReconnectEnumEnabled)
	} else {
		result.ChangeComputeType = aws.String(workspaces.ReconnectEnumDisabled)
	}

	if p["increase_volume_size"].(bool) {
		result.IncreaseVolumeSize = aws.String(workspaces.ReconnectEnumEnabled)
	} else {
		result.IncreaseVolumeSize = aws.String(workspaces.ReconnectEnumDisabled)
	}

	if p["rebuild_workspace"].(bool) {
		result.RebuildWorkspace = aws.String(workspaces.ReconnectEnumEnabled)
	} else {
		result.RebuildWorkspace = aws.String(workspaces.ReconnectEnumDisabled)
	}

	if p["restart_workspace"].(bool) {
		result.RestartWorkspace = aws.String(workspaces.ReconnectEnumEnabled)
	} else {
		result.RestartWorkspace = aws.String(workspaces.ReconnectEnumDisabled)
	}

	if p["switch_running_mode"].(bool) {
		result.SwitchRunningMode = aws.String(workspaces.ReconnectEnumEnabled)
	} else {
		result.SwitchRunningMode = aws.String(workspaces.ReconnectEnumDisabled)
	}

	return result
}

func expandWorkspaceCreationProperties(properties []interface{}) *workspaces.WorkspaceCreationProperties {
	if len(properties) == 0 || properties[0] == nil {
		return nil
	}

	p := properties[0].(map[string]interface{})

	return &workspaces.WorkspaceCreationProperties{
		CustomSecurityGroupId:           aws.String(p["custom_security_group_id"].(string)),
		DefaultOu:                       aws.String(p["default_ou"].(string)),
		EnableInternetAccess:            aws.Bool(p["enable_internet_access"].(bool)),
		EnableMaintenanceMode:           aws.Bool(p["enable_maintenance_mode"].(bool)),
		UserEnabledAsLocalAdministrator: aws.Bool(p["user_enabled_as_local_administrator"].(bool)),
	}
}

func flattenSelfServicePermissions(permissions *workspaces.SelfservicePermissions) []interface{} {
	if permissions == nil {
		return []interface{}{}
	}

	result := map[string]interface{}{}

	switch *permissions.ChangeComputeType {
	case workspaces.ReconnectEnumEnabled:
		result["change_compute_type"] = true
	case workspaces.ReconnectEnumDisabled:
		result["change_compute_type"] = false
	default:
		result["change_compute_type"] = nil
	}

	switch *permissions.IncreaseVolumeSize {
	case workspaces.ReconnectEnumEnabled:
		result["increase_volume_size"] = true
	case workspaces.ReconnectEnumDisabled:
		result["increase_volume_size"] = false
	default:
		result["increase_volume_size"] = nil
	}

	switch *permissions.RebuildWorkspace {
	case workspaces.ReconnectEnumEnabled:
		result["rebuild_workspace"] = true
	case workspaces.ReconnectEnumDisabled:
		result["rebuild_workspace"] = false
	default:
		result["rebuild_workspace"] = nil
	}

	switch *permissions.RestartWorkspace {
	case workspaces.ReconnectEnumEnabled:
		result["restart_workspace"] = true
	case workspaces.ReconnectEnumDisabled:
		result["restart_workspace"] = false
	default:
		result["restart_workspace"] = nil
	}

	switch *permissions.SwitchRunningMode {
	case workspaces.ReconnectEnumEnabled:
		result["switch_running_mode"] = true
	case workspaces.ReconnectEnumDisabled:
		result["switch_running_mode"] = false
	default:
		result["switch_running_mode"] = nil
	}

	return []interface{}{result}
}

func flattenWorkspaceCreationProperties(properties *workspaces.DefaultWorkspaceCreationProperties) []interface{} {
	if properties == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"custom_security_group_id":            aws.StringValue(properties.CustomSecurityGroupId),
			"default_ou":                          aws.StringValue(properties.DefaultOu),
			"enable_internet_access":              aws.BoolValue(properties.EnableInternetAccess),
			"enable_maintenance_mode":             aws.BoolValue(properties.EnableMaintenanceMode),
			"user_enabled_as_local_administrator": aws.BoolValue(properties.UserEnabledAsLocalAdministrator),
		},
	}
}
