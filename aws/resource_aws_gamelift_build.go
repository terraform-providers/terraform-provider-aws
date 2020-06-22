package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsGameliftBuild() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGameliftBuildCreate,
		Read:   resourceAwsGameliftBuildRead,
		Update: resourceAwsGameliftBuildUpdate,
		Delete: resourceAwsGameliftBuildDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"operating_system": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					gamelift.OperatingSystemAmazonLinux,
					gamelift.OperatingSystemWindows2012,
				}, false),
			},
			"storage_location": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"key": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
			"version": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsGameliftBuildCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	sl := expandGameliftStorageLocation(d.Get("storage_location").([]interface{}))
	input := gamelift.CreateBuildInput{
		Name:            aws.String(d.Get("name").(string)),
		OperatingSystem: aws.String(d.Get("operating_system").(string)),
		StorageLocation: sl,
		Tags:            keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().GameliftTags(),
	}
	if v, ok := d.GetOk("version"); ok {
		input.Version = aws.String(v.(string))
	}
	log.Printf("[INFO] Creating Gamelift Build: %s", input)
	var out *gamelift.CreateBuildOutput
	err := resource.Retry(30*time.Second, func() *resource.RetryError {
		var err error
		out, err = conn.CreateBuild(&input)
		if err != nil {
			if isAWSErr(err, gamelift.ErrCodeInvalidRequestException, "Provided build is not accessible.") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		out, err = conn.CreateBuild(&input)
	}
	if err != nil {
		return fmt.Errorf("Error creating Gamelift build client: %s", err)
	}

	d.SetId(*out.Build.BuildId)

	stateConf := resource.StateChangeConf{
		Pending: []string{gamelift.BuildStatusInitialized},
		Target:  []string{gamelift.BuildStatusReady},
		Timeout: 1 * time.Minute,
		Refresh: func() (interface{}, string, error) {
			out, err := conn.DescribeBuild(&gamelift.DescribeBuildInput{
				BuildId: aws.String(d.Id()),
			})
			if err != nil {
				return 42, "", err
			}

			return out, *out.Build.Status, nil
		},
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}

	return resourceAwsGameliftBuildRead(d, meta)
}

func resourceAwsGameliftBuildRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[INFO] Reading Gamelift Build: %s", d.Id())
	out, err := conn.DescribeBuild(&gamelift.DescribeBuildInput{
		BuildId: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, gamelift.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] Gamelift Build (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}
	b := out.Build

	d.Set("name", b.Name)
	d.Set("operating_system", b.OperatingSystem)
	d.Set("version", b.Version)

	arn := aws.StringValue(b.BuildArn)
	d.Set("arn", arn)
	tags, err := keyvaluetags.GameliftListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Game Lift Build (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsGameliftBuildUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	log.Printf("[INFO] Updating Gamelift Build: %s", d.Id())
	input := gamelift.UpdateBuildInput{
		BuildId: aws.String(d.Id()),
		Name:    aws.String(d.Get("name").(string)),
	}
	if v, ok := d.GetOk("version"); ok {
		input.Version = aws.String(v.(string))
	}

	_, err := conn.UpdateBuild(&input)
	if err != nil {
		return err
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.GameliftUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Game Lift Build (%s) tags: %s", arn, err)
		}
	}

	return resourceAwsGameliftBuildRead(d, meta)
}

func resourceAwsGameliftBuildDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	log.Printf("[INFO] Deleting Gamelift Build: %s", d.Id())
	_, err := conn.DeleteBuild(&gamelift.DeleteBuildInput{
		BuildId: aws.String(d.Id()),
	})
	return err
}

func expandGameliftStorageLocation(cfg []interface{}) *gamelift.S3Location {
	loc := cfg[0].(map[string]interface{})
	return &gamelift.S3Location{
		Bucket:  aws.String(loc["bucket"].(string)),
		Key:     aws.String(loc["key"].(string)),
		RoleArn: aws.String(loc["role_arn"].(string)),
	}
}
