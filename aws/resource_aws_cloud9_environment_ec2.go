package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloud9"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloud9/waiter"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
)

func resourceAwsCloud9EnvironmentEc2() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloud9EnvironmentEc2Create,
		Read:   resourceAwsCloud9EnvironmentEc2Read,
		Update: resourceAwsCloud9EnvironmentEc2Update,
		Delete: resourceAwsCloud9EnvironmentEc2Delete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 60),
			},
			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"automatic_stop_time_minutes": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtMost(20160),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 200),
			},
			"owner_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsCloud9EnvironmentEc2Create(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloud9conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	params := &cloud9.CreateEnvironmentEC2Input{
		InstanceType:       aws.String(d.Get("instance_type").(string)),
		Name:               aws.String(d.Get("name").(string)),
		ClientRequestToken: aws.String(resource.UniqueId()),
		Tags:               tags.IgnoreAws().Cloud9Tags(),
	}

	if v, ok := d.GetOk("automatic_stop_time_minutes"); ok {
		params.AutomaticStopTimeMinutes = aws.Int64(int64(v.(int)))
	}
	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("owner_arn"); ok {
		params.OwnerArn = aws.String(v.(string))
	}
	if v, ok := d.GetOk("subnet_id"); ok {
		params.SubnetId = aws.String(v.(string))
	}

	var out *cloud9.CreateEnvironmentEC2Output
	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		var err error
		out, err = conn.CreateEnvironmentEC2(params)
		if err != nil {
			// NotFoundException: User arn:aws:iam::*******:user/****** does not exist.
			if isAWSErr(err, cloud9.ErrCodeNotFoundException, "User") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		out, err = conn.CreateEnvironmentEC2(params)
	}

	if err != nil {
		return fmt.Errorf("Error creating Cloud9 EC2 Environment: %w", err)
	}
	d.SetId(aws.StringValue(out.EnvironmentId))

	_, err = waiter.EnvironmentReady(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error waiting for Cloud9 EC2 Environment %q to be ready: %w", d.Id(), err)
	}

	return resourceAwsCloud9EnvironmentEc2Read(d, meta)
}

func resourceAwsCloud9EnvironmentEc2Read(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloud9conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[INFO] Reading Cloud9 Environment EC2 %s", d.Id())

	out, err := conn.DescribeEnvironments(&cloud9.DescribeEnvironmentsInput{
		EnvironmentIds: []*string{aws.String(d.Id())},
	})
	if err != nil {
		if isAWSErr(err, cloud9.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] Cloud9 Environment EC2 (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}
	if len(out.Environments) == 0 {
		log.Printf("[WARN] Cloud9 Environment EC2 (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	env := out.Environments[0]

	arn := aws.StringValue(env.Arn)
	d.Set("arn", arn)
	d.Set("description", env.Description)
	d.Set("name", env.Name)
	d.Set("owner_arn", env.OwnerArn)
	d.Set("type", env.Type)

	tags, err := keyvaluetags.Cloud9ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Cloud9 EC2 Environment (%s): %w", arn, err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	log.Printf("[DEBUG] Received Cloud9 Environment EC2: %s", env)

	return nil
}

func resourceAwsCloud9EnvironmentEc2Update(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloud9conn

	if d.HasChangeExcept("tags") {
		input := cloud9.UpdateEnvironmentInput{
			Description:   aws.String(d.Get("description").(string)),
			EnvironmentId: aws.String(d.Id()),
			Name:          aws.String(d.Get("name").(string)),
		}

		log.Printf("[INFO] Updating Cloud9 Environment EC2: %s", input)

		out, err := conn.UpdateEnvironment(&input)
		if err != nil {
			return fmt.Errorf("error updating Cloud9 EC2 Environment (%s): %w", d.Id(), err)
		}

		log.Printf("[DEBUG] Cloud9 Environment EC2 updated: %s", out)
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		arn := d.Get("arn").(string)

		if err := keyvaluetags.Cloud9UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Cloud9 EC2 Environment (%s) tags: %w", arn, err)
		}
	}

	return resourceAwsCloud9EnvironmentEc2Read(d, meta)
}

func resourceAwsCloud9EnvironmentEc2Delete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloud9conn

	_, err := conn.DeleteEnvironment(&cloud9.DeleteEnvironmentInput{
		EnvironmentId: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, cloud9.ErrCodeNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting Cloud9 EC2 Environment: %w", err)
	}

	_, err = waiter.EnvironmentDeleted(conn, d.Id())
	if err != nil {
		if isAWSErr(err, cloud9.ErrCodeNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error waiting for Cloud9 EC2 Environment %q to be deleted: %w", d.Id(), err)
	}

	return nil
}
