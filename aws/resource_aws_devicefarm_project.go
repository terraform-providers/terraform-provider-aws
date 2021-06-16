package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/devicefarm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsDevicefarmProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDevicefarmProjectCreate,
		Read:   resourceAwsDevicefarmProjectRead,
		Update: resourceAwsDevicefarmProjectUpdate,
		Delete: resourceAwsDevicefarmProjectDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"default_job_timeout_minutes": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},
		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsDevicefarmProjectCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).devicefarmconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &devicefarm.CreateProjectInput{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk("default_job_timeout_minutes"); ok {
		input.DefaultJobTimeoutMinutes = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Creating DeviceFarm Project: %s", name)
	out, err := conn.CreateProject(input)
	if err != nil {
		return fmt.Errorf("Error creating DeviceFarm Project: %w", err)
	}

	arn := aws.StringValue(out.Project.Arn)
	log.Printf("[DEBUG] Successsfully Created DeviceFarm Project: %s", arn)
	d.SetId(arn)

	if len(tags) > 0 {
		if err := keyvaluetags.DevicefarmUpdateTags(conn, arn, nil, tags); err != nil {
			return fmt.Errorf("error updating DeviceFarm Project (%s) tags: %w", arn, err)
		}
	}

	return resourceAwsDevicefarmProjectRead(d, meta)
}

func resourceAwsDevicefarmProjectRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).devicefarmconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &devicefarm.GetProjectInput{
		Arn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading DeviceFarm Project: %s", d.Id())
	out, err := conn.GetProject(input)
	if err != nil {
		if isAWSErr(err, devicefarm.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] DeviceFarm Project (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading DeviceFarm Project: %w", err)
	}

	project := out.Project
	arn := aws.StringValue(project.Arn)
	d.Set("name", project.Name)
	d.Set("arn", arn)
	d.Set("default_job_timeout_minutes", project.DefaultJobTimeoutMinutes)

	tags, err := keyvaluetags.DevicefarmListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for DeviceFarm Project (%s): %w", arn, err)
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

func resourceAwsDevicefarmProjectUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).devicefarmconn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &devicefarm.UpdateProjectInput{
			Arn: aws.String(d.Id()),
		}

		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChange("default_job_timeout_minutes") {
			input.DefaultJobTimeoutMinutes = aws.Int64(int64(d.Get("default_job_timeout_minutes").(int)))
		}

		log.Printf("[DEBUG] Updating DeviceFarm Project: %s", d.Id())
		_, err := conn.UpdateProject(input)
		if err != nil {
			return fmt.Errorf("Error Updating DeviceFarm Project: %w", err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.DevicefarmUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating DeviceFarm Project (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceAwsDevicefarmProjectRead(d, meta)
}

func resourceAwsDevicefarmProjectDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).devicefarmconn

	input := &devicefarm.DeleteProjectInput{
		Arn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DeviceFarm Project: %s", d.Id())
	_, err := conn.DeleteProject(input)
	if err != nil {
		if isAWSErr(err, devicefarm.ErrCodeNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting DeviceFarm Project: %w", err)
	}

	return nil
}
