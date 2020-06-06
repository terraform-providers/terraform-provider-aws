package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elastictranscoder"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsElasticTranscoderPipeline() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsElasticTranscoderPipelineCreate,
		Read:   resourceAwsElasticTranscoderPipelineRead,
		Update: resourceAwsElasticTranscoderPipelineUpdate,
		Delete: resourceAwsElasticTranscoderPipelineDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"aws_kms_key_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},

			// ContentConfig also requires ThumbnailConfig
			"content_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					// elastictranscoder.PipelineOutputConfig
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Optional: true,
							// AWS may insert the bucket name here taken from output_bucket
							Computed: true,
						},
						"storage_class": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"content_config_permissions": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"grantee": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"grantee_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"input_bucket": {
				Type:     schema.TypeString,
				Required: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if !regexp.MustCompile(`^[.0-9A-Za-z-_]+$`).MatchString(value) {
						errors = append(errors, fmt.Errorf(
							"only alphanumeric characters, hyphens, underscores, and periods allowed in %q", k))
					}
					if len(value) > 40 {
						errors = append(errors, fmt.Errorf("%q cannot be longer than 40 characters", k))
					}
					return
				},
			},

			"notifications": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"completed": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"error": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"progressing": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"warning": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			// The output_bucket must be set, or both of content_config.bucket
			// and thumbnail_config.bucket.
			// This is set as Computed, because the API may or may not return
			// this as set based on the other 2 configurations.
			"output_bucket": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"role": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},

			"thumbnail_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					// elastictranscoder.PipelineOutputConfig
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Optional: true,
							// AWS may insert the bucket name here taken from output_bucket
							Computed: true,
						},
						"storage_class": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"thumbnail_config_permissions": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"grantee": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"grantee_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsElasticTranscoderPipelineCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elastictranscoderconn

	req := &elastictranscoder.CreatePipelineInput{
		AwsKmsKeyArn:    aws.String(d.Get("aws_kms_key_arn").(string)),
		ContentConfig:   expandETPiplineOutputConfig(d, "content_config"),
		InputBucket:     aws.String(d.Get("input_bucket").(string)),
		Notifications:   expandETNotifications(d),
		Role:            aws.String(d.Get("role").(string)),
		ThumbnailConfig: expandETPiplineOutputConfig(d, "thumbnail_config"),
	}

	if v, ok := d.GetOk("output_bucket"); ok && v.(string) != "" {
		req.OutputBucket = aws.String(v.(string))
	}

	if name, ok := d.GetOk("name"); ok {
		req.Name = aws.String(name.(string))
	} else {
		name := resource.PrefixedUniqueId("tf-et-")
		d.Set("name", name)
		req.Name = aws.String(name)
	}

	if (req.OutputBucket == nil && (req.ContentConfig == nil || req.ContentConfig.Bucket == nil)) ||
		(req.OutputBucket != nil && req.ContentConfig != nil && req.ContentConfig.Bucket != nil) {
		return fmt.Errorf("you must specify only one of output_bucket or content_config.bucket")
	}

	log.Printf("[DEBUG] Elastic Transcoder Pipeline create opts: %s", req)
	resp, err := conn.CreatePipeline(req)
	if err != nil {
		return fmt.Errorf("Error creating Elastic Transcoder Pipeline: %s", err)
	}

	d.SetId(*resp.Pipeline.Id)

	for _, w := range resp.Warnings {
		log.Printf("[WARN] Elastic Transcoder Pipeline %v: %v", *w.Code, *w.Message)
	}

	return resourceAwsElasticTranscoderPipelineRead(d, meta)
}

func expandETNotifications(d *schema.ResourceData) *elastictranscoder.Notifications {
	list, ok := d.GetOk("notifications")
	if !ok {
		return nil
	}

	l := list.([]interface{})
	if len(l) == 0 {
		return nil
	}

	if l[0] == nil {
		log.Printf("[ERR] First element of Notifications list is nil")
		return nil
	}

	rN := l[0].(map[string]interface{})

	return &elastictranscoder.Notifications{
		Completed:   aws.String(rN["completed"].(string)),
		Error:       aws.String(rN["error"].(string)),
		Progressing: aws.String(rN["progressing"].(string)),
		Warning:     aws.String(rN["warning"].(string)),
	}
}

func flattenETNotifications(n *elastictranscoder.Notifications) []map[string]interface{} {
	if n == nil {
		return nil
	}

	allEmpty := func(s ...*string) bool {
		for _, s := range s {
			if s != nil && *s != "" {
				return false
			}
		}
		return true
	}

	// the API always returns a Notifications value, even when all fields are nil
	if allEmpty(n.Completed, n.Error, n.Progressing, n.Warning) {
		return nil
	}

	result := map[string]interface{}{
		"completed":   aws.StringValue(n.Completed),
		"error":       aws.StringValue(n.Error),
		"progressing": aws.StringValue(n.Progressing),
		"warning":     aws.StringValue(n.Warning),
	}

	return []map[string]interface{}{result}
}

func expandETPiplineOutputConfig(d *schema.ResourceData, key string) *elastictranscoder.PipelineOutputConfig {
	list, ok := d.GetOk(key)
	if !ok {
		return nil
	}

	l := list.([]interface{})
	if len(l) == 0 {
		return nil
	}

	cc := l[0].(map[string]interface{})

	cfg := &elastictranscoder.PipelineOutputConfig{
		Bucket:       aws.String(cc["bucket"].(string)),
		StorageClass: aws.String(cc["storage_class"].(string)),
	}

	switch key {
	case "content_config":
		cfg.Permissions = expandETPermList(d.Get("content_config_permissions").(*schema.Set))
	case "thumbnail_config":
		cfg.Permissions = expandETPermList(d.Get("thumbnail_config_permissions").(*schema.Set))
	}

	return cfg
}

func flattenETPipelineOutputConfig(cfg *elastictranscoder.PipelineOutputConfig) []map[string]interface{} {
	if cfg == nil {
		return nil
	}

	result := map[string]interface{}{
		"bucket":        aws.StringValue(cfg.Bucket),
		"storage_class": aws.StringValue(cfg.StorageClass),
	}

	return []map[string]interface{}{result}
}

func expandETPermList(permissions *schema.Set) []*elastictranscoder.Permission {
	var perms []*elastictranscoder.Permission

	for _, p := range permissions.List() {
		if p == nil {
			continue
		}

		m := p.(map[string]interface{})

		perm := &elastictranscoder.Permission{
			Access:      expandStringList(m["access"].([]interface{})),
			Grantee:     aws.String(m["grantee"].(string)),
			GranteeType: aws.String(m["grantee_type"].(string)),
		}

		perms = append(perms, perm)
	}
	return perms
}

func flattenETPermList(perms []*elastictranscoder.Permission) []map[string]interface{} {
	var set []map[string]interface{}

	for _, p := range perms {
		result := map[string]interface{}{
			"access":       flattenStringList(p.Access),
			"grantee":      aws.StringValue(p.Grantee),
			"grantee_type": aws.StringValue(p.GranteeType),
		}

		set = append(set, result)
	}
	return set
}

func resourceAwsElasticTranscoderPipelineUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elastictranscoderconn

	req := &elastictranscoder.UpdatePipelineInput{
		Id: aws.String(d.Id()),
	}

	if d.HasChange("aws_kms_key_arn") {
		req.AwsKmsKeyArn = aws.String(d.Get("aws_kms_key_arn").(string))
	}

	if d.HasChange("content_config") {
		req.ContentConfig = expandETPiplineOutputConfig(d, "content_config")
	}

	if d.HasChange("input_bucket") {
		req.InputBucket = aws.String(d.Get("input_bucket").(string))
	}

	if d.HasChange("name") {
		req.Name = aws.String(d.Get("name").(string))
	}

	if d.HasChange("notifications") {
		req.Notifications = expandETNotifications(d)
	}

	if d.HasChange("role") {
		req.Role = aws.String(d.Get("role").(string))
	}

	if d.HasChange("thumbnail_config") {
		req.ThumbnailConfig = expandETPiplineOutputConfig(d, "thumbnail_config")
	}

	log.Printf("[DEBUG] Updating Elastic Transcoder Pipeline: %#v", req)
	output, err := conn.UpdatePipeline(req)
	if err != nil {
		return fmt.Errorf("Error updating Elastic Transcoder pipeline: %s", err)
	}

	for _, w := range output.Warnings {
		log.Printf("[WARN] Elastic Transcoder Pipeline %v: %v", *w.Code, *w.Message)
	}

	return resourceAwsElasticTranscoderPipelineRead(d, meta)
}

func resourceAwsElasticTranscoderPipelineRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elastictranscoderconn

	resp, err := conn.ReadPipeline(&elastictranscoder.ReadPipelineInput{
		Id: aws.String(d.Id()),
	})

	if err != nil {
		if isAWSErr(err, elastictranscoder.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] No such resource found for Elastic Transcoder Pipeline (%s)", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	log.Printf("[DEBUG] Elastic Transcoder Pipeline Read response: %#v", resp)

	pipeline := resp.Pipeline

	d.Set("arn", pipeline.Arn)

	if arn := pipeline.AwsKmsKeyArn; arn != nil {
		d.Set("aws_kms_key_arn", arn)
	}

	if pipeline.ContentConfig != nil {
		err := d.Set("content_config", flattenETPipelineOutputConfig(pipeline.ContentConfig))
		if err != nil {
			return fmt.Errorf("error setting content_config: %s", err)
		}

		if pipeline.ContentConfig.Permissions != nil {
			err := d.Set("content_config_permissions", flattenETPermList(pipeline.ContentConfig.Permissions))
			if err != nil {
				return fmt.Errorf("error setting content_config_permissions: %s", err)
			}
		}
	}

	d.Set("input_bucket", pipeline.InputBucket)
	d.Set("name", pipeline.Name)

	notifications := flattenETNotifications(pipeline.Notifications)
	if notifications != nil {
		if err := d.Set("notifications", notifications); err != nil {
			return fmt.Errorf("error setting notifications: %s", err)
		}
	}

	d.Set("role", pipeline.Role)

	if pipeline.ThumbnailConfig != nil {
		err := d.Set("thumbnail_config", flattenETPipelineOutputConfig(pipeline.ThumbnailConfig))
		if err != nil {
			return fmt.Errorf("error setting thumbnail_config: %s", err)
		}

		if pipeline.ThumbnailConfig.Permissions != nil {
			err := d.Set("thumbnail_config_permissions", flattenETPermList(pipeline.ThumbnailConfig.Permissions))
			if err != nil {
				return fmt.Errorf("error setting thumbnail_config_permissions: %s", err)
			}
		}
	}

	if pipeline.OutputBucket != nil {
		d.Set("output_bucket", pipeline.OutputBucket)
	}

	return nil
}

func resourceAwsElasticTranscoderPipelineDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elastictranscoderconn

	log.Printf("[DEBUG] Elastic Transcoder Delete Pipeline: %s", d.Id())
	_, err := conn.DeletePipeline(&elastictranscoder.DeletePipelineInput{
		Id: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error deleting Elastic Transcoder Pipeline: %s", err)
	}
	return nil
}
