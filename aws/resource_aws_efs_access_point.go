package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsEfsAccessPoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEfsAccessPointCreate,
		Read:   resourceAwsEfsAccessPointRead,
		Update: resourceAwsEfsAccessPointUpdate,
		Delete: resourceAwsEfsAccessPointDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"file_system_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_system_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"posix_user": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gid": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"uid": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"secondary_gids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeInt},
							Set:      schema.HashInt,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"root_directory": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"path": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
						},
						"creation_info": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"owner_gid": {
										Type:     schema.TypeInt,
										Required: true,
										ForceNew: true,
									},
									"owner_uid": {
										Type:     schema.TypeInt,
										Required: true,
										ForceNew: true,
									},
									"permissions": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
								},
							},
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsEfsAccessPointCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn

	fsId := d.Get("file_system_id").(string)

	input := efs.CreateAccessPointInput{
		FileSystemId: aws.String(fsId),
		Tags:         keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().EfsTags(),
	}

	if v, ok := d.GetOk("posix_user"); ok {
		input.PosixUser = expandEfsAccessPointPosixUser(v.([]interface{}))
	}

	if v, ok := d.GetOk("root_directory"); ok {
		input.RootDirectory = expandEfsAccessPointRootDirectory(v.([]interface{}))
	}

	log.Printf("[DEBUG] Creating EFS Access Point: %#v", input)

	ap, err := conn.CreateAccessPoint(&input)
	if err != nil {
		return fmt.Errorf("error creating EFS Access Point for File System (%s): %w", fsId, err)
	}

	d.SetId(*ap.AccessPointId)
	log.Printf("[INFO] EFS access point ID: %s", d.Id())

	stateConf := &resource.StateChangeConf{
		Pending: []string{efs.LifeCycleStateCreating},
		Target:  []string{efs.LifeCycleStateAvailable},
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DescribeAccessPoints(&efs.DescribeAccessPointsInput{
				AccessPointId: aws.String(d.Id()),
			})
			if err != nil {
				return nil, "error", err
			}

			if hasEmptyAccessPoints(resp) {
				return nil, "error", fmt.Errorf("EFS access point %q could not be found.", d.Id())
			}

			mt := resp.AccessPoints[0]

			log.Printf("[DEBUG] Current status of %q: %q", aws.StringValue(mt.AccessPointId), aws.StringValue(mt.LifeCycleState))
			return mt, aws.StringValue(mt.LifeCycleState), nil
		},
		Timeout:    10 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for EFS access point (%s) to create: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] EFS access point created: %s", *ap.AccessPointId)

	return resourceAwsEfsAccessPointRead(d, meta)
}

func resourceAwsEfsAccessPointUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.EfsUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EFS file system (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsEfsAccessPointRead(d, meta)
}

func resourceAwsEfsAccessPointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeAccessPoints(&efs.DescribeAccessPointsInput{
		AccessPointId: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, efs.ErrCodeAccessPointNotFound, "") {
			log.Printf("[WARN] EFS access point %q could not be found.", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading EFS access point %s: %s", d.Id(), err)
	}

	if hasEmptyAccessPoints(resp) {
		return fmt.Errorf("EFS access point %q could not be found.", d.Id())
	}

	ap := resp.AccessPoints[0]

	log.Printf("[DEBUG] Found EFS access point: %#v", ap)

	d.SetId(*ap.AccessPointId)

	fsARN := arn.ARN{
		AccountID: meta.(*AWSClient).accountid,
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("file-system/%s", aws.StringValue(ap.FileSystemId)),
		Service:   "elasticfilesystem",
	}.String()

	d.Set("file_system_arn", fsARN)
	d.Set("file_system_id", ap.FileSystemId)
	d.Set("arn", ap.AccessPointArn)
	d.Set("owner_id", ap.OwnerId)

	if err := d.Set("posix_user", flattenEfsAccessPointPosixUser(ap.PosixUser)); err != nil {
		return fmt.Errorf("error setting posix user: %s", err)
	}

	if err := d.Set("root_directory", flattenEfsAccessPointRootDirectory(ap.RootDirectory)); err != nil {
		return fmt.Errorf("error setting root directory: %s", err)
	}

	if err := d.Set("tags", keyvaluetags.EfsKeyValueTags(ap.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsEfsAccessPointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn

	log.Printf("[DEBUG] Deleting EFS access point %q", d.Id())
	_, err := conn.DeleteAccessPoint(&efs.DeleteAccessPointInput{
		AccessPointId: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error deleting EFS Access Point (%s): %w", d.Id(), err)
	}

	err = waitForDeleteEfsAccessPoint(conn, d.Id(), 10*time.Minute)
	if err != nil {
		return fmt.Errorf("Error waiting for EFS access point (%q) to delete: %s", d.Id(), err.Error())
	}

	log.Printf("[DEBUG] EFS access point %q deleted.", d.Id())

	return nil
}

func waitForDeleteEfsAccessPoint(conn *efs.EFS, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{efs.LifeCycleStateAvailable, efs.LifeCycleStateDeleting, efs.LifeCycleStateDeleted},
		Target:  []string{},
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DescribeAccessPoints(&efs.DescribeAccessPointsInput{
				AccessPointId: aws.String(id),
			})
			if err != nil {
				if isAWSErr(err, efs.ErrCodeAccessPointNotFound, "") {
					return nil, "", nil
				}

				return nil, "error", err
			}

			if hasEmptyAccessPoints(resp) {
				return nil, "", nil
			}

			mt := resp.AccessPoints[0]

			log.Printf("[DEBUG] Current status of %q: %q", aws.StringValue(mt.AccessPointId), aws.StringValue(mt.LifeCycleState))
			return mt, aws.StringValue(mt.LifeCycleState), nil
		},
		Timeout:    timeout,
		Delay:      2 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err := stateConf.WaitForState()
	return err
}

func hasEmptyAccessPoints(aps *efs.DescribeAccessPointsOutput) bool {
	if aps != nil && len(aps.AccessPoints) > 0 {
		return false
	}
	return true
}

func expandEfsAccessPointPosixUser(pUser []interface{}) *efs.PosixUser {
	if len(pUser) < 1 || pUser[0] == nil {
		return nil
	}

	m := pUser[0].(map[string]interface{})

	posixUser := &efs.PosixUser{
		Gid: aws.Int64(int64(m["gid"].(int))),
		Uid: aws.Int64(int64(m["uid"].(int))),
	}

	if v, ok := m["secondary_gids"].(*schema.Set); ok && len(v.List()) > 0 {
		posixUser.SecondaryGids = expandInt64Set(v)
	}

	return posixUser
}

func expandEfsAccessPointRootDirectory(rDir []interface{}) *efs.RootDirectory {
	if len(rDir) < 1 || rDir[0] == nil {
		return nil
	}

	m := rDir[0].(map[string]interface{})

	rootDir := &efs.RootDirectory{}

	if v, ok := m["path"]; ok {
		rootDir.Path = aws.String(v.(string))
	}

	if v, ok := m["creation_info"]; ok {
		rootDir.CreationInfo = expandEfsAccessPointRootDirectoryCreationInfo(v.([]interface{}))
	}

	return rootDir
}

func expandEfsAccessPointRootDirectoryCreationInfo(cInfo []interface{}) *efs.CreationInfo {
	if len(cInfo) < 1 || cInfo[0] == nil {
		return nil
	}

	m := cInfo[0].(map[string]interface{})

	creationInfo := &efs.CreationInfo{
		OwnerGid:    aws.Int64(int64(m["owner_gid"].(int))),
		OwnerUid:    aws.Int64(int64(m["owner_uid"].(int))),
		Permissions: aws.String(m["permissions"].(string)),
	}

	return creationInfo
}

func flattenEfsAccessPointPosixUser(posixUser *efs.PosixUser) []interface{} {
	if posixUser == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"gid":            aws.Int64Value(posixUser.Gid),
		"uid":            aws.Int64Value(posixUser.Uid),
		"secondary_gids": aws.Int64ValueSlice(posixUser.SecondaryGids),
	}

	return []interface{}{m}
}

func flattenEfsAccessPointRootDirectory(rDir *efs.RootDirectory) []interface{} {
	if rDir == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"path":          aws.StringValue(rDir.Path),
		"creation_info": flattenEfsAccessPointRootDirectoryCreationInfo(rDir.CreationInfo),
	}

	return []interface{}{m}
}

func flattenEfsAccessPointRootDirectoryCreationInfo(cInfo *efs.CreationInfo) []interface{} {
	if cInfo == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"owner_gid":   aws.Int64Value(cInfo.OwnerGid),
		"owner_uid":   aws.Int64Value(cInfo.OwnerUid),
		"permissions": aws.StringValue(cInfo.Permissions),
	}

	return []interface{}{m}
}
