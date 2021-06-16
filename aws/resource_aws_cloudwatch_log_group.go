package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsCloudWatchLogGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudWatchLogGroupCreate,
		Read:   resourceAwsCloudWatchLogGroupRead,
		Update: resourceAwsCloudWatchLogGroupUpdate,
		Delete: resourceAwsCloudWatchLogGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validateLogGroupName,
			},
			"name_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateLogGroupNamePrefix,
			},

			"retention_in_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntInSlice([]int{0, 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, 3653}),
			},

			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsCloudWatchLogGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatchlogsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	var logGroupName string
	if v, ok := d.GetOk("name"); ok {
		logGroupName = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		logGroupName = resource.PrefixedUniqueId(v.(string))
	} else {
		logGroupName = resource.UniqueId()
	}

	log.Printf("[DEBUG] Creating CloudWatch Log Group: %s", logGroupName)

	params := &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(logGroupName),
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		params.KmsKeyId = aws.String(v.(string))
	}

	if len(tags) > 0 {
		params.Tags = tags.IgnoreAws().CloudwatchlogsTags()
	}

	_, err := conn.CreateLogGroup(params)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == cloudwatchlogs.ErrCodeResourceAlreadyExistsException {
			return fmt.Errorf("Creating CloudWatch Log Group failed: %s:  The CloudWatch Log Group '%s' already exists.", err, d.Get("name").(string))
		}
		return fmt.Errorf("Creating CloudWatch Log Group failed: %s '%s'", err, d.Get("name"))
	}

	d.SetId(logGroupName)

	log.Println("[INFO] CloudWatch Log Group created")

	if v, ok := d.GetOk("retention_in_days"); ok {
		input := cloudwatchlogs.PutRetentionPolicyInput{
			LogGroupName:    aws.String(logGroupName),
			RetentionInDays: aws.Int64(int64(v.(int))),
		}
		log.Printf("[DEBUG] Setting retention for CloudWatch Log Group: %q: %s", logGroupName, input)
		_, err = conn.PutRetentionPolicy(&input)

		if err != nil {
			return err
		}
	}

	return resourceAwsCloudWatchLogGroupRead(d, meta)
}

func resourceAwsCloudWatchLogGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatchlogsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading CloudWatch Log Group: %q", d.Get("name").(string))
	lg, err := lookupCloudWatchLogGroup(conn, d.Id())
	if err != nil {
		return err
	}

	if lg == nil {
		log.Printf("[DEBUG] CloudWatch Group %q Not Found", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", strings.TrimSuffix(aws.StringValue(lg.Arn), ":*"))
	d.Set("name", lg.LogGroupName)
	d.Set("kms_key_id", lg.KmsKeyId)
	d.Set("retention_in_days", lg.RetentionInDays)

	tags, err := keyvaluetags.CloudwatchlogsListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for CloudWatch Logs Group (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func lookupCloudWatchLogGroup(conn *cloudwatchlogs.CloudWatchLogs, name string) (*cloudwatchlogs.LogGroup, error) {
	input := &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: aws.String(name),
	}
	var logGroup *cloudwatchlogs.LogGroup
	err := conn.DescribeLogGroupsPages(input, func(page *cloudwatchlogs.DescribeLogGroupsOutput, lastPage bool) bool {
		for _, lg := range page.LogGroups {
			if aws.StringValue(lg.LogGroupName) == name {
				logGroup = lg
				return false
			}
		}
		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	return logGroup, nil
}

func resourceAwsCloudWatchLogGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatchlogsconn

	name := d.Id()
	log.Printf("[DEBUG] Updating CloudWatch Log Group: %q", name)

	if d.HasChange("retention_in_days") {
		var err error

		if v, ok := d.GetOk("retention_in_days"); ok {
			input := cloudwatchlogs.PutRetentionPolicyInput{
				LogGroupName:    aws.String(name),
				RetentionInDays: aws.Int64(int64(v.(int))),
			}
			log.Printf("[DEBUG] Setting retention for CloudWatch Log Group: %q: %s", name, input)
			_, err = conn.PutRetentionPolicy(&input)
		} else {
			log.Printf("[DEBUG] Deleting retention for CloudWatch Log Group: %q", name)
			_, err = conn.DeleteRetentionPolicy(&cloudwatchlogs.DeleteRetentionPolicyInput{
				LogGroupName: aws.String(name),
			})
		}

		if err != nil {
			return err
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.CloudwatchlogsUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating CloudWatch Log Group (%s) tags: %s", d.Id(), err)
		}
	}

	if d.HasChange("kms_key_id") && !d.IsNewResource() {
		_, newKey := d.GetChange("kms_key_id")

		if newKey.(string) == "" {
			_, err := conn.DisassociateKmsKey(&cloudwatchlogs.DisassociateKmsKeyInput{
				LogGroupName: aws.String(name),
			})
			if err != nil {
				return err
			}
		} else {
			_, err := conn.AssociateKmsKey(&cloudwatchlogs.AssociateKmsKeyInput{
				LogGroupName: aws.String(name),
				KmsKeyId:     aws.String(newKey.(string)),
			})
			if err != nil {
				return err
			}
		}
	}

	return resourceAwsCloudWatchLogGroupRead(d, meta)
}

func resourceAwsCloudWatchLogGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatchlogsconn
	log.Printf("[INFO] Deleting CloudWatch Log Group: %s", d.Id())
	_, err := conn.DeleteLogGroup(&cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: aws.String(d.Get("name").(string)),
	})
	if err != nil {
		return fmt.Errorf("Error deleting CloudWatch Log Group: %s", err)
	}
	log.Println("[INFO] CloudWatch Log Group deleted")

	return nil
}
