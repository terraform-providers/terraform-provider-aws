package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

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
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"storage_location": {
				Type:     schema.TypeList,
				Required: true,
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

func resourceAwsGameliftScriptCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	sl := expandGameliftStorageLocation(d.Get("storage_location").([]interface{}))
	input := gamelift.CreateScriptInput{
		Name:            aws.String(d.Get("name").(string)),
		StorageLocation: sl,
		Tags:            keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().GameliftTags(),
	}
	if v, ok := d.GetOk("version"); ok {
		input.Version = aws.String(v.(string))
	}
	log.Printf("[INFO] Creating Gamelift Script: %s", input)
	out, err := conn.CreateScript(&input)
	if err != nil {
		return fmt.Errorf("Error creating Gamelift Script: %s", err)
	}
	d.SetId(*out.Script.ScriptId)

	return resourceAwsGameliftScriptRead(d, meta)
}

func resourceAwsGameliftScriptRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

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
		return fmt.Errorf("error listing tags for Game Lift Script (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsGameliftScriptUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	log.Printf("[INFO] Updating Gamelift Script: %s", d.Id())
	input := gamelift.UpdateScriptInput{
		ScriptId:        aws.String(d.Id()),
		Name:            aws.String(d.Get("name").(string)),
		StorageLocation: expandGameliftStorageLocation(d.Get("storage_location").([]interface{})),
	}

	if v, ok := d.GetOk("version"); ok {
		input.Version = aws.String(v.(string))
	}

	_, err := conn.UpdateScript(&input)
	if err != nil {
		return err
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.GameliftUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Game Lift Script (%s) tags: %s", arn, err)
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
	return err
}

func flattenStorageLocation(sl *gamelift.S3Location) []interface{} {
	if sl == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"bucket":   aws.StringValue(sl.Bucket),
		"key":      aws.StringValue(sl.Key),
		"role_arn": aws.StringValue(sl.RoleArn),
	}

	return []interface{}{m}
}
