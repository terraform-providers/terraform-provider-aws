package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

const awsMutexGameliftScript = `aws_gamelift_script`

func resourceAwsGameliftScript() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGameliftScriptCreate,
		Read:   resourceAwsGameliftScriptRead,
		Update: resourceAwsGameliftScriptUpdate,
		Delete: resourceAwsGameliftScriptDelete,
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
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"storage_location": {
				Type:         schema.TypeList,
				Optional:     true,
				Computed:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"zip_file", "storage_location"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Required: true,
						},
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"object_version": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
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
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"zip_file": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"zip_file", "storage_location"},
			},
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsGameliftScriptCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := gamelift.CreateScriptInput{
		Name: aws.String(d.Get("name").(string)),
		Tags: tags.IgnoreAws().GameliftTags(),
	}

	if v, ok := d.GetOk("version"); ok {
		input.Version = aws.String(v.(string))
	}

	if v, ok := d.GetOk("storage_location"); ok {
		input.StorageLocation = expandGameliftStorageLocation(v.([]interface{}))
	}

	if v, ok := d.GetOk("zip_file"); ok {
		awsMutexKV.Lock(awsMutexGameliftScript)
		defer awsMutexKV.Unlock(awsMutexGameliftScript)
		file, err := loadFileContent(v.(string))
		if err != nil {
			return fmt.Errorf("unable to load %q: %w", v.(string), err)
		}
		input.ZipFile = file
	}

	log.Printf("[INFO] Creating Gamelift Script: %s", input)
	var out *gamelift.CreateScriptOutput
	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		var err error
		out, err = conn.CreateScript(&input)
		if err != nil {
			if isAWSErr(err, gamelift.ErrCodeInvalidRequestException, "Provided resource is not accessible") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		out, err = conn.CreateScript(&input)
	}
	if err != nil {
		return fmt.Errorf("Error creating Gamelift Script: %w", err)
	}
	d.SetId(aws.StringValue(out.Script.ScriptId))

	return resourceAwsGameliftScriptRead(d, meta)
}

func resourceAwsGameliftScriptRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[INFO] Reading Gamelift Script: %s", d.Id())
	out, err := conn.DescribeScript(&gamelift.DescribeScriptInput{
		ScriptId: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, gamelift.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] Gamelift Script (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}
	sc := out.Script

	d.Set("name", sc.Name)
	d.Set("version", sc.Version)
	d.Set("storage_location", flattenStorageLocation(sc.StorageLocation))

	arn := aws.StringValue(sc.ScriptArn)
	d.Set("arn", arn)
	tags, err := keyvaluetags.GameliftListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("error listing tags for Game Lift Script (%s): %w", arn, err)
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

func resourceAwsGameliftScriptUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	log.Printf("[INFO] Updating Gamelift Script: %s", d.Id())
	if d.HasChangesExcept("tags", "tags_all") {
		input := gamelift.UpdateScriptInput{
			ScriptId: aws.String(d.Id()),
			Name:     aws.String(d.Get("name").(string)),
		}

		if d.HasChange("version") {
			if v, ok := d.GetOk("version"); ok {
				input.Version = aws.String(v.(string))
			}
		}

		if d.HasChange("storage_location") {
			if v, ok := d.GetOk("storage_location"); ok {
				input.StorageLocation = expandGameliftStorageLocation(v.([]interface{}))
			}
		}

		if d.HasChange("zip_file") {
			if v, ok := d.GetOk("zip_file"); ok {
				awsMutexKV.Lock(awsMutexGameliftScript)
				defer awsMutexKV.Unlock(awsMutexGameliftScript)
				file, err := loadFileContent(v.(string))
				if err != nil {
					return fmt.Errorf("unable to load %q: %w", v.(string), err)
				}
				input.ZipFile = file
			}
		}

		_, err := conn.UpdateScript(&input)
		if err != nil {
			return fmt.Errorf("error updating Game Lift Script: %w", err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		arn := d.Get("arn").(string)

		if err := keyvaluetags.GameliftUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Game Lift Script (%s) tags: %w", arn, err)
		}
	}

	return resourceAwsGameliftScriptRead(d, meta)
}

func resourceAwsGameliftScriptDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	log.Printf("[INFO] Deleting Gamelift Script: %s", d.Id())
	_, err := conn.DeleteScript(&gamelift.DeleteScriptInput{
		ScriptId: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, gamelift.ErrCodeNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting Gamelift script: %w", err)
	}
	return nil
}

func flattenStorageLocation(sl *gamelift.S3Location) []interface{} {
	if sl == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"bucket":         aws.StringValue(sl.Bucket),
		"key":            aws.StringValue(sl.Key),
		"role_arn":       aws.StringValue(sl.RoleArn),
		"object_version": aws.StringValue(sl.ObjectVersion),
	}

	return []interface{}{m}
}
