package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ebs/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsEbsSnapshotImport() *schema.Resource {
	return &schema.Resource{
		Create:        resourceAwsEbsSnapshotImportCreate,
		Read:          resourceAwsEbsSnapshotImportRead,
		Update:        resourceAwsEbsSnapshotImportUpdate,
		Delete:        resourceAwsEbsSnapshotImportDelete,
		CustomizeDiff: SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_data": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"comment": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"upload_end": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsRFC3339Time,
						},
						"upload_size": {
							Type:     schema.TypeFloat,
							Optional: true,
							ForceNew: true,
						},
						"upload_start": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsRFC3339Time,
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"disk_container": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"format": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(ec2.DiskImageFormat_Values(), false),
						},
						"url": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ExactlyOneOf: []string{"disk_container.0.user_bucket", "disk_container.0.url"},
						},
						"user_bucket": {
							Type:         schema.TypeList,
							Optional:     true,
							ForceNew:     true,
							ExactlyOneOf: []string{"disk_container.0.user_bucket", "disk_container.0.url"},
							MaxItems:     1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"s3_bucket": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"s3_key": {
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
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encrypted": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"volume_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"role_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "vmimport",
			},
			"data_encryption_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},
	}
}

func expandAwsEbsSnapshotClientData(tfMap map[string]interface{}) (*ec2.ClientData, error) {
	clientData := &ec2.ClientData{}

	if v, ok := tfMap["comment"].(string); ok {
		clientData.Comment = aws.String(v)
	}

	if v, ok := tfMap["upload_end"].(string); ok {
		upload_end, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return nil, fmt.Errorf("error parsing upload_end to timestamp: %s", err)
		}
		clientData.UploadEnd = aws.Time(upload_end)
	}

	if v, ok := tfMap["upload_size"].(float64); ok {
		clientData.UploadSize = aws.Float64(v)
	}

	if v, ok := tfMap["upload_start"].(string); ok {
		upload_start, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return nil, fmt.Errorf("error parsing upload_start to timestamp: %s", err)
		}
		clientData.UploadStart = aws.Time(upload_start)
	}

	return clientData, nil
}

func expandAwsEbsSnapshotDiskContainer(tfMap map[string]interface{}) *ec2.SnapshotDiskContainer {
	diskContainer := &ec2.SnapshotDiskContainer{
		Format: aws.String(tfMap["format"].(string)),
	}

	if v, ok := tfMap["description"].(string); ok {
		diskContainer.Description = aws.String(v)
	}

	if v, ok := tfMap["url"].(string); ok && v != "" {
		diskContainer.Url = aws.String(v)
	}

	if v, ok := tfMap["user_bucket"]; ok {
		vL := v.([]interface{})
		for _, v := range vL {
			ub := v.(map[string]interface{})
			diskContainer.UserBucket = &ec2.UserBucket{
				S3Bucket: aws.String(ub["s3_bucket"].(string)),
				S3Key:    aws.String(ub["s3_key"].(string)),
			}
		}
	}

	return diskContainer
}

func resourceAwsEbsSnapshotImportCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig

	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	request := &ec2.ImportSnapshotInput{
		TagSpecifications: ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeImportSnapshotTask),
	}

	if v, ok := d.GetOk("client_data"); ok {
		vL := v.([]interface{})
		for _, v := range vL {
			cdv, ok := v.(map[string]interface{})

			if !ok {
				continue
			}

			clientData, err := expandAwsEbsSnapshotClientData(cdv)

			if err != nil {
				return err
			}

			if clientData == nil {
				continue
			}

			request.ClientData = clientData
		}
	}

	request.ClientToken = aws.String(resource.UniqueId())

	if v, ok := d.GetOk("description"); ok {
		request.Description = aws.String(v.(string))
	}

	v := d.Get("disk_container")
	vL := v.([]interface{})
	for _, v := range vL {
		dcv, ok := v.(map[string]interface{})

		if !ok {
			continue
		}

		diskContainer := expandAwsEbsSnapshotDiskContainer(dcv)

		if diskContainer == nil {
			continue
		}

		request.DiskContainer = diskContainer
	}

	if v, ok := d.GetOk("encrypted"); ok {
		request.Encrypted = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		request.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("role_name"); ok {
		request.RoleName = aws.String(v.(string))
	}

	err := resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		var resp *ec2.ImportSnapshotOutput
		resp, err := conn.ImportSnapshot(request)
		// Error: InvalidParameter: The service role terraform-20201121225356951800000001 provided does not exist or does not have sufficient permissions
		// status code: 400, request id: b0abc3d2-5b59-4e5c-b748-c1cb084020c0

		if isAWSErr(err, "InvalidParameter", "provided does not exist or does not have sufficient permissions") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		importTaskId := aws.StringValue(resp.ImportTaskId)

		res, err := waiter.EbsSnapshotImportCompleted(conn, importTaskId)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error waiting for snapshot (%s) to be imported: %s", d.Id(), err))
		}

		d.SetId(aws.StringValue(res.SnapshotId))

		tags := d.Get("tags").(map[string]interface{})
		if len(tags) > 0 {
			if err := keyvaluetags.Ec2CreateTags(conn, d.Id(), tags); err != nil {
				return resource.NonRetryableError(fmt.Errorf("error setting tags: %s", err))
			}
		}

		err = resourceAwsEbsSnapshotImportRead(d, meta)
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		return fmt.Errorf("timeout error importing EBS Snapshot: %s", err)
	}

	if err != nil {
		return fmt.Errorf("error importing EBS Snapshot: %s", err)
	}

	return nil
}

func resourceAwsEbsSnapshotImportRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	req := &ec2.DescribeSnapshotsInput{
		SnapshotIds: []*string{aws.String(d.Id())},
	}
	res, err := conn.DescribeSnapshots(req)
	if err != nil {
		if isAWSErr(err, "InvalidSnapshot.NotFound", "") {
			log.Printf("[WARN] EBS Snapshot %q Not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if len(res.Snapshots) == 0 {
		log.Printf("[WARN] EBS Snapshot %q Not found - removing from state", d.Id())
		d.SetId("")
		return nil
	}

	snapshot := res.Snapshots[0]

	d.Set("description", snapshot.Description)
	d.Set("owner_id", snapshot.OwnerId)
	d.Set("encrypted", snapshot.Encrypted)
	d.Set("owner_alias", snapshot.OwnerAlias)
	d.Set("data_encryption_key_id", snapshot.DataEncryptionKeyId)
	d.Set("kms_key_id", snapshot.KmsKeyId)
	d.Set("volume_size", snapshot.VolumeSize)

	tags := keyvaluetags.Ec2KeyValueTags(snapshot.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	snapshotArn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("snapshot/%s", d.Id()),
		Service:   "ec2",
	}.String()

	d.Set("arn", snapshotArn)

	return nil
}

func resourceAwsEbsSnapshotImportUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsEbsSnapshotImportRead(d, meta)
}

func resourceAwsEbsSnapshotImportDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	input := &ec2.DeleteSnapshotInput{
		SnapshotId: aws.String(d.Id()),
	}
	err := resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := conn.DeleteSnapshot(input)
		if err == nil {
			return nil
		}
		if isAWSErr(err, "SnapshotInUse", "") {
			return resource.RetryableError(fmt.Errorf("EBS SnapshotInUse - trying again while it detaches"))
		}
		return resource.NonRetryableError(err)
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DeleteSnapshot(input)
	}
	if err != nil {
		return fmt.Errorf("Error deleting EBS snapshot: %s", err)
	}
	return nil
}
