package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/workspaces/waiter"
)

func resourceAwsWorkspacesWorkspace() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWorkspacesWorkspaceCreate,
		Read:   resourceAwsWorkspacesWorkspaceRead,
		Update: resourceAwsWorkspacesWorkspaceUpdate,
		Delete: resourceAwsWorkspacesWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"bundle_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"computer_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"root_volume_encryption_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"user_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_volume_encryption_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"volume_encryption_key": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"workspace_properties": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"compute_type_name": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  workspaces.ComputeValue,
							ValidateFunc: validation.StringInSlice([]string{
								workspaces.ComputeValue,
								workspaces.ComputeStandard,
								workspaces.ComputePerformance,
								workspaces.ComputePower,
								workspaces.ComputePowerpro,
								workspaces.ComputeGraphics,
								workspaces.ComputeGraphicspro,
							}, false),
						},
						"root_volume_size_gib": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  80,
							ValidateFunc: validation.Any(
								validation.IntInSlice([]int{80}),
								validation.IntBetween(175, 2000),
							),
						},
						"running_mode": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  workspaces.RunningModeAlwaysOn,
							ValidateFunc: validation.StringInSlice([]string{
								workspaces.RunningModeAlwaysOn,
								workspaces.RunningModeAutoStop,
							}, false),
						},
						"running_mode_auto_stop_timeout_in_minutes": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
								val := v.(int)
								if val%60 != 0 {
									errors = append(errors, fmt.Errorf(
										"%q should be configured in 60-minute intervals, got: %d", k, val))
								}
								return
							},
						},
						"user_volume_size_gib": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  10,
							ValidateFunc: validation.Any(
								validation.IntInSlice([]int{10, 50}),
								validation.IntBetween(100, 2000),
							),
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsWorkspacesWorkspaceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn

	tags := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().WorkspacesTags()

	input := &workspaces.WorkspaceRequest{
		BundleId:                    aws.String(d.Get("bundle_id").(string)),
		DirectoryId:                 aws.String(d.Get("directory_id").(string)),
		UserName:                    aws.String(d.Get("user_name").(string)),
		RootVolumeEncryptionEnabled: aws.Bool(d.Get("root_volume_encryption_enabled").(bool)),
		UserVolumeEncryptionEnabled: aws.Bool(d.Get("user_volume_encryption_enabled").(bool)),
		Tags:                        tags,
	}

	if v, ok := d.GetOk("volume_encryption_key"); ok {
		input.VolumeEncryptionKey = aws.String(v.(string))
	}

	input.WorkspaceProperties = expandWorkspaceProperties(d.Get("workspace_properties").([]interface{}))

	log.Printf("[DEBUG] Creating workspace...\n%#v\n", *input)
	resp, err := conn.CreateWorkspaces(&workspaces.CreateWorkspacesInput{
		Workspaces: []*workspaces.WorkspaceRequest{input},
	})
	if err != nil {
		return err
	}

	wsFail := resp.FailedRequests
	if len(wsFail) > 0 {
		return fmt.Errorf("workspace creation failed: %s", *wsFail[0].ErrorMessage)
	}

	workspaceID := aws.StringValue(resp.PendingRequests[0].WorkspaceId)

	log.Printf("[DEBUG] Waiting for workspace %q to be available...", workspaceID)
	_, err = waiter.WorkspaceAvailable(conn, workspaceID)
	if err != nil {
		return fmt.Errorf("workspace %q is not available: %s", workspaceID, err)
	}

	d.SetId(workspaceID)
	log.Printf("[DEBUG] Workspace %q is available", workspaceID)

	return resourceAwsWorkspacesWorkspaceRead(d, meta)
}

func resourceAwsWorkspacesWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	rawOutput, state, err := waiter.WorkspaceState(conn, d.Id())()
	if err != nil {
		return fmt.Errorf("error reading workspace (%s): %s", d.Id(), err)
	}
	if state == workspaces.WorkspaceStateTerminated {
		log.Printf("[WARN] workspace (%s) is not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	workspace := rawOutput.(*workspaces.Workspace)
	d.Set("bundle_id", workspace.BundleId)
	d.Set("directory_id", workspace.DirectoryId)
	d.Set("ip_address", workspace.IpAddress)
	d.Set("computer_name", workspace.ComputerName)
	d.Set("state", workspace.State)
	d.Set("root_volume_encryption_enabled", workspace.RootVolumeEncryptionEnabled)
	d.Set("user_name", workspace.UserName)
	d.Set("user_volume_encryption_enabled", workspace.UserVolumeEncryptionEnabled)
	d.Set("volume_encryption_key", workspace.VolumeEncryptionKey)
	if err := d.Set("workspace_properties", flattenWorkspaceProperties(workspace.WorkspaceProperties)); err != nil {
		return fmt.Errorf("error setting workspace properties: %s", err)
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

func resourceAwsWorkspacesWorkspaceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn

	// IMPORTANT: Only one workspace property could be changed in a time.
	// I've create AWS Support feature request to allow multiple properties modification in a time.
	// https://docs.aws.amazon.com/workspaces/latest/adminguide/modify-workspaces.html

	if d.HasChange("workspace_properties.0.compute_type_name") {
		if err := workspacePropertyUpdate("compute_type_name", conn, d); err != nil {
			return err
		}
	}

	if d.HasChange("workspace_properties.0.root_volume_size_gib") {
		if err := workspacePropertyUpdate("root_volume_size_gib", conn, d); err != nil {
			return err
		}
	}

	if d.HasChange("workspace_properties.0.running_mode") {
		if err := workspacePropertyUpdate("running_mode", conn, d); err != nil {
			return err
		}
	}

	if d.HasChange("workspace_properties.0.running_mode_auto_stop_timeout_in_minutes") {
		if err := workspacePropertyUpdate("running_mode_auto_stop_timeout_in_minutes", conn, d); err != nil {
			return err
		}
	}

	if d.HasChange("workspace_properties.0.user_volume_size_gib") {
		if err := workspacePropertyUpdate("user_volume_size_gib", conn, d); err != nil {
			return err
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.WorkspacesUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsWorkspacesWorkspaceRead(d, meta)
}

func resourceAwsWorkspacesWorkspaceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn

	err := workspaceDelete(d.Id(), conn)
	if err != nil {
		return err
	}

	return nil
}

func workspaceDelete(id string, conn *workspaces.WorkSpaces) error {
	log.Printf("[DEBUG] Terminating workspace %q", id)
	_, err := conn.TerminateWorkspaces(&workspaces.TerminateWorkspacesInput{
		TerminateWorkspaceRequests: []*workspaces.TerminateRequest{
			{
				WorkspaceId: aws.String(id),
			},
		},
	})
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Waiting for workspace %q to be terminated", id)
	_, err = waiter.WorkspaceTerminated(conn, id)
	if err != nil {
		return fmt.Errorf("workspace %q was not terminated: %s", id, err)
	}
	log.Printf("[DEBUG] Workspace %q is terminated", id)

	return nil
}

func workspacePropertyUpdate(p string, conn *workspaces.WorkSpaces, d *schema.ResourceData) error {
	id := d.Id()

	log.Printf("[DEBUG] Modifying workspace %q %s property...", id, p)

	var wsp *workspaces.WorkspaceProperties

	switch p {
	case "compute_type_name":
		wsp = &workspaces.WorkspaceProperties{
			ComputeTypeName: aws.String(d.Get("workspace_properties.0.compute_type_name").(string)),
		}
	case "root_volume_size_gib":
		wsp = &workspaces.WorkspaceProperties{
			RootVolumeSizeGib: aws.Int64(int64(d.Get("workspace_properties.0.root_volume_size_gib").(int))),
		}
	case "running_mode":
		wsp = &workspaces.WorkspaceProperties{
			RunningMode: aws.String(d.Get("workspace_properties.0.running_mode").(string)),
		}
	case "running_mode_auto_stop_timeout_in_minutes":
		if d.Get("workspace_properties.0.running_mode") != workspaces.RunningModeAutoStop {
			log.Printf("[DEBUG] Property running_mode_auto_stop_timeout_in_minutes makes sense only for AUTO_STOP running mode")
			return nil
		}

		wsp = &workspaces.WorkspaceProperties{
			RunningModeAutoStopTimeoutInMinutes: aws.Int64(int64(d.Get("workspace_properties.0.running_mode_auto_stop_timeout_in_minutes").(int))),
		}
	case "user_volume_size_gib":
		wsp = &workspaces.WorkspaceProperties{
			UserVolumeSizeGib: aws.Int64(int64(d.Get("workspace_properties.0.user_volume_size_gib").(int))),
		}
	}

	_, err := conn.ModifyWorkspaceProperties(&workspaces.ModifyWorkspacePropertiesInput{
		WorkspaceId:         aws.String(id),
		WorkspaceProperties: wsp,
	})
	if err != nil {
		return fmt.Errorf("workspace %q %s property was not modified: %w", d.Id(), p, err)
	}

	log.Printf("[DEBUG] Waiting for workspace %q %s property to be modified...", d.Id(), p)
	_, err = waiter.WorkspaceUpdated(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error modifying workspace %q property %q was not modified: %w", d.Id(), p, err)
	}
	log.Printf("[DEBUG] Workspace %q %s property is modified", d.Id(), p)

	return nil
}

func expandWorkspaceProperties(properties []interface{}) *workspaces.WorkspaceProperties {
	log.Printf("[DEBUG] Expand Workspace properties: %+v ", properties)

	if len(properties) == 0 || properties[0] == nil {
		return nil
	}

	p := properties[0].(map[string]interface{})

	return &workspaces.WorkspaceProperties{
		ComputeTypeName:                     aws.String(p["compute_type_name"].(string)),
		RootVolumeSizeGib:                   aws.Int64(int64(p["root_volume_size_gib"].(int))),
		RunningMode:                         aws.String(p["running_mode"].(string)),
		RunningModeAutoStopTimeoutInMinutes: aws.Int64(int64(p["running_mode_auto_stop_timeout_in_minutes"].(int))),
		UserVolumeSizeGib:                   aws.Int64(int64(p["user_volume_size_gib"].(int))),
	}
}

func flattenWorkspaceProperties(properties *workspaces.WorkspaceProperties) []map[string]interface{} {
	log.Printf("[DEBUG] Flatten workspace properties: %+v ", properties)

	if properties == nil {
		return []map[string]interface{}{}
	}

	return []map[string]interface{}{
		{
			"compute_type_name":                         aws.StringValue(properties.ComputeTypeName),
			"root_volume_size_gib":                      int(aws.Int64Value(properties.RootVolumeSizeGib)),
			"running_mode":                              aws.StringValue(properties.RunningMode),
			"running_mode_auto_stop_timeout_in_minutes": int(aws.Int64Value(properties.RunningModeAutoStopTimeoutInMinutes)),
			"user_volume_size_gib":                      int(aws.Int64Value(properties.UserVolumeSizeGib)),
		},
	}
}
