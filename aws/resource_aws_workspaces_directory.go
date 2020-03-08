package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
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
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
		for _, id := range v.(*schema.Set).List() {
			input.SubnetIds = append(input.SubnetIds, aws.String(id.(string)))
		}
	}

	log.Printf("[DEBUG] Regestering workspaces directory...\n%#v\n", *input)
	_, err := conn.RegisterWorkspaceDirectory(input)
	if err != nil {
		return err
	}
	d.SetId(directoryId)

	log.Printf("[DEBUG] Waiting for workspaces directory %q to become registered...", d.Id())
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			workspaces.WorkspaceDirectoryStateRegistering,
		},
		Target:       []string{workspaces.WorkspaceDirectoryStateRegistered},
		Refresh:      workspacesDirectoryRefreshStateFunc(conn, directoryId),
		PollInterval: 30 * time.Second,
		Timeout:      10 * time.Minute,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error registering directory: %s", err)
	}
	log.Printf("[DEBUG] Workspaces directory %q is registered", d.Id())

	log.Printf("[DEBUG] Modifying workspaces directory %q self-service permissions...", d.Id())
	if v, ok := d.GetOk("self_service_permissions"); ok {
		_, err := conn.ModifySelfservicePermissions(&workspaces.ModifySelfservicePermissionsInput{
			ResourceId:             aws.String(directoryId),
			SelfservicePermissions: expandSelfServicePermissions(v.([]interface{})),
		})
		if err != nil {
			return fmt.Errorf("error setting self service permissions: %s", err)
		}
	}
	log.Printf("[DEBUG] Workspaces directory %q self-service permissions are set", d.Id())

	return resourceAwsWorkspacesDirectoryRead(d, meta)
}

func resourceAwsWorkspacesDirectoryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn

	raw, state, err := workspacesDirectoryRefreshStateFunc(conn, d.Id())()
	if err != nil {
		return fmt.Errorf("error getting workspaces directory (%s): %s", d.Id(), err)
	}
	if state == workspaces.WorkspaceDirectoryStateDeregistered {
		log.Printf("[WARN] workspaces directory (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	dir := raw.(*workspaces.WorkspaceDirectory)
	d.Set("directory_id", dir.DirectoryId)
	d.Set("subnet_ids", dir.SubnetIds)
	if err := d.Set("self_service_permissions", flattenSelfServicePermissions(dir.SelfservicePermissions)); err != nil {
		return fmt.Errorf("error setting self_service_permissions: %s", err)
	}

	tags, err := keyvaluetags.WorkspacesListTags(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error listing tags: %s", err)
	}

	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsWorkspacesDirectoryUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn

	if d.HasChange("self_service_permissions") {
		log.Printf("[DEBUG] Modifying workspaces directory %q self-service permissions...", d.Id())
		permissions := d.Get("self_service_permissions").([]interface{})

		_, err := conn.ModifySelfservicePermissions(&workspaces.ModifySelfservicePermissionsInput{
			ResourceId:             aws.String(d.Id()),
			SelfservicePermissions: expandSelfServicePermissions(permissions),
		})
		if err != nil {
			return fmt.Errorf("error updating self service permissions: %s", err)
		}
		log.Printf("[DEBUG] Workspaces directory %q self-service permissions are set", d.Id())
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
		return fmt.Errorf("error deleting workspaces directory (%s): %s", d.Id(), err)
	}
	log.Printf("[DEBUG] Workspaces directory %q is deregistered", d.Id())

	return nil
}

func workspacesDirectoryDelete(id string, conn *workspaces.WorkSpaces) error {
	input := &workspaces.DeregisterWorkspaceDirectoryInput{
		DirectoryId: aws.String(id),
	}

	log.Printf("[DEBUG] Deregistering Workspace Directory %q", id)
	_, err := conn.DeregisterWorkspaceDirectory(input)
	if err != nil {
		return fmt.Errorf("error deregistering Workspace Directory %q: %w", id, err)
	}

	log.Printf("[DEBUG] Waiting for Workspace Directory %q to be deregistered", id)
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			workspaces.WorkspaceDirectoryStateRegistering,
			workspaces.WorkspaceDirectoryStateRegistered,
			workspaces.WorkspaceDirectoryStateDeregistering,
		},
		Target: []string{
			workspaces.WorkspaceDirectoryStateDeregistered,
		},
		Refresh:      workspacesDirectoryRefreshStateFunc(conn, id),
		PollInterval: 30 * time.Second,
		Timeout:      10 * time.Minute,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for Workspace Directory %q to be deregistered: %w", id, err)
	}
	return nil
}

func workspacesDirectoryRefreshStateFunc(conn *workspaces.WorkSpaces, directoryID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeWorkspaceDirectories(&workspaces.DescribeWorkspaceDirectoriesInput{
			DirectoryIds: []*string{aws.String(directoryID)},
		})
		if err != nil {
			return nil, workspaces.WorkspaceDirectoryStateError, err
		}
		if len(resp.Directories) == 0 {
			return resp, workspaces.WorkspaceDirectoryStateDeregistered, nil
		}
		directory := resp.Directories[0]
		return directory, aws.StringValue(directory.State), nil
	}
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
