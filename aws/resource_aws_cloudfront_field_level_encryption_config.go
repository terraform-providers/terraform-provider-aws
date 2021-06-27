package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudfront/finder"
)

func resourceAwsCloudfrontFieldLevelEncryptionConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudfrontFieldLevelEncryptionConfigCreate,
		Read:   resourceAwsCloudfrontFieldLevelEncryptionConfigRead,
		Update: resourceAwsCloudfrontFieldLevelEncryptionConfigUpdate,
		Delete: resourceAwsCloudfrontFieldLevelEncryptionConfigDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"content_type_profile_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"forward_when_content_type_is_unknown": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"content_type_profile": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"content_type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"format": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(cloudfront.Format_Values(), false),
									},
									"profile_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"query_arg_profile_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"forward_when_query_arg_is_unknown": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"query_arg_profile": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"profile_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"query_arg": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"caller_reference": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsCloudfrontFieldLevelEncryptionConfigCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	fl := &cloudfront.FieldLevelEncryptionConfig{
		CallerReference:          aws.String(resource.UniqueId()),
		ContentTypeProfileConfig: expandAwsCloudfrontFieldLevelEncryptionConfigContentTypeProfileConfig(d.Get("content_type_profile_config").([]interface{})),
		QueryArgProfileConfig:    expandAwsCloudfrontFieldLevelEncryptionConfigQueryArgProfileConfig(d.Get("query_arg_profile_config").([]interface{})),
	}

	if v, ok := d.GetOk("comment"); ok {
		fl.Comment = aws.String(v.(string))
	}

	input := &cloudfront.CreateFieldLevelEncryptionConfigInput{
		FieldLevelEncryptionConfig: fl,
	}

	resp, err := conn.CreateFieldLevelEncryptionConfig(input)
	if err != nil {
		return fmt.Errorf("error creating Cloudfront Field Level Encryption Config (%s): %w", d.Id(), err)
	}

	d.SetId(aws.StringValue(resp.FieldLevelEncryption.Id))

	return resourceAwsCloudfrontFieldLevelEncryptionConfigRead(d, meta)
}

func resourceAwsCloudfrontFieldLevelEncryptionConfigRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	resp, err := finder.FieldLevelEncryptionConfigByID(conn, d.Id())
	if isAWSErr(err, cloudfront.ErrCodeNoSuchFieldLevelEncryptionConfig, "") {
		log.Printf("[WARN] Cloudfront Field Level Encryption Config %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Cloudfront Field Level Encryption Config (%s): %w", d.Id(), err)
	}
	Config := resp.FieldLevelEncryptionConfig
	d.Set("etag", resp.ETag)
	d.Set("comment", Config.Comment)
	d.Set("caller_reference", Config.CallerReference)

	if err := d.Set("content_type_profile_config", flattenAwsCloudfrontFieldLevelEncryptionConfigContentTypeProfileConfig(Config.ContentTypeProfileConfig)); err != nil {
		return fmt.Errorf("error setting content_type_profile_config %w", err)
	}

	if err := d.Set("query_arg_profile_config", flattenAwsCloudfrontFieldLevelEncryptionConfigQueryArgProfileConfig(Config.QueryArgProfileConfig)); err != nil {
		return fmt.Errorf("error setting query_arg_profile_config %w", err)
	}

	return nil
}

func resourceAwsCloudfrontFieldLevelEncryptionConfigUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	fl := &cloudfront.FieldLevelEncryptionConfig{
		CallerReference:          aws.String(d.Get("caller_reference").(string)),
		ContentTypeProfileConfig: expandAwsCloudfrontFieldLevelEncryptionConfigContentTypeProfileConfig(d.Get("content_type_profile_config").([]interface{})),
		QueryArgProfileConfig:    expandAwsCloudfrontFieldLevelEncryptionConfigQueryArgProfileConfig(d.Get("query_arg_profile_config").([]interface{})),
	}

	if v, ok := d.GetOk("comment"); ok {
		fl.Comment = aws.String(v.(string))
	}

	input := &cloudfront.UpdateFieldLevelEncryptionConfigInput{
		FieldLevelEncryptionConfig: fl,
		Id:                         aws.String(d.Id()),
		IfMatch:                    aws.String(d.Get("etag").(string)),
	}

	_, err := conn.UpdateFieldLevelEncryptionConfig(input)
	if err != nil {
		return fmt.Errorf("error creating Cloudfront Field Level Encryption Config (%s): %w", d.Id(), err)
	}

	return resourceAwsCloudfrontFieldLevelEncryptionConfigRead(d, meta)
}

func resourceAwsCloudfrontFieldLevelEncryptionConfigDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	input := &cloudfront.DeleteFieldLevelEncryptionConfigInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	}

	_, err := conn.DeleteFieldLevelEncryptionConfig(input)
	if err != nil {
		return fmt.Errorf("error deleting Cloudfront Field Level Encryption Config (%s): %w", d.Id(), err)
	}

	return nil
}
func expandAwsCloudfrontFieldLevelEncryptionConfigContentTypeProfileConfig(config []interface{}) *cloudfront.ContentTypeProfileConfig {
	if len(config) == 0 || config[0] == nil {
		return nil
	}

	m := config[0].(map[string]interface{})

	profileConf := &cloudfront.ContentTypeProfileConfig{
		ForwardWhenContentTypeIsUnknown: aws.Bool(m["forward_when_content_type_is_unknown"].(bool)),
		ContentTypeProfiles:             expandAwsCloudfrontFieldLevelEncryptionConfigContentTypeProfiles(m["content_type_profile"].([]interface{})),
	}

	return profileConf
}

func expandAwsCloudfrontFieldLevelEncryptionConfigContentTypeProfiles(config []interface{}) *cloudfront.ContentTypeProfiles {
	profiles := make([]*cloudfront.ContentTypeProfile, 0)

	for _, raw := range config {
		m := raw.(map[string]interface{})

		profile := &cloudfront.ContentTypeProfile{
			ContentType: aws.String(m["content_type"].(string)),
			Format:      aws.String(m["format"].(string)),
		}

		if v, ok := m["profile_id"].(string); ok && v != "" {
			profile.ProfileId = aws.String(v)
		}

		profiles = append(profiles, profile)
	}

	contentTypeProfiles := &cloudfront.ContentTypeProfiles{
		Quantity: aws.Int64(int64(len(config))),
		Items:    profiles,
	}

	return contentTypeProfiles
}

func expandAwsCloudfrontFieldLevelEncryptionConfigQueryArgProfileConfig(config []interface{}) *cloudfront.QueryArgProfileConfig {
	if len(config) == 0 || config[0] == nil {
		return nil
	}

	m := config[0].(map[string]interface{})

	profileConf := &cloudfront.QueryArgProfileConfig{
		ForwardWhenQueryArgProfileIsUnknown: aws.Bool(m["forward_when_query_arg_is_unknown"].(bool)),
		QueryArgProfiles:                    expandAwsCloudfrontFieldLevelEncryptionConfigQueryArgProfiles(m["query_arg_profile"].([]interface{})),
	}

	return profileConf
}

func expandAwsCloudfrontFieldLevelEncryptionConfigQueryArgProfiles(config []interface{}) *cloudfront.QueryArgProfiles {
	profiles := make([]*cloudfront.QueryArgProfile, 0)

	for _, raw := range config {
		m := raw.(map[string]interface{})

		profile := &cloudfront.QueryArgProfile{
			ProfileId: aws.String(m["profile_id"].(string)),
			QueryArg:  aws.String(m["query_arg"].(string)),
		}

		profiles = append(profiles, profile)
	}

	contentTypeProfiles := &cloudfront.QueryArgProfiles{
		Quantity: aws.Int64(int64(len(config))),
		Items:    profiles,
	}

	return contentTypeProfiles
}

func flattenAwsCloudfrontFieldLevelEncryptionConfigContentTypeProfileConfig(config *cloudfront.ContentTypeProfileConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := make(map[string]interface{})

	m["forward_when_content_type_is_unknown"] = aws.BoolValue(config.ForwardWhenContentTypeIsUnknown)
	if config.ContentTypeProfiles != nil {
		m["content_type_profile"] = flattenAwsCloudfrontFieldLevelEncryptionConfigContentTypeProfiles(config.ContentTypeProfiles.Items)
	}

	return []map[string]interface{}{m}
}

func flattenAwsCloudfrontFieldLevelEncryptionConfigContentTypeProfiles(profiles []*cloudfront.ContentTypeProfile) []interface{} {
	if len(profiles) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, profile := range profiles {
		tfMap := map[string]interface{}{
			"content_type": aws.StringValue(profile.ContentType),
			"format":       aws.StringValue(profile.Format),
		}

		if aws.StringValue(profile.ProfileId) != "" {
			tfMap["profile_id"] = aws.StringValue(profile.ProfileId)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenAwsCloudfrontFieldLevelEncryptionConfigQueryArgProfileConfig(config *cloudfront.QueryArgProfileConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := make(map[string]interface{})

	m["forward_when_query_arg_is_unknown"] = aws.BoolValue(config.ForwardWhenQueryArgProfileIsUnknown)
	if config.QueryArgProfiles != nil {
		m["query_arg_profile"] = flattenAwsCloudfrontFieldLevelEncryptionConfigQueryArgProfiles(config.QueryArgProfiles.Items)
	}

	return []map[string]interface{}{m}
}

func flattenAwsCloudfrontFieldLevelEncryptionConfigQueryArgProfiles(profiles []*cloudfront.QueryArgProfile) []interface{} {
	if len(profiles) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, profile := range profiles {
		tfMap := map[string]interface{}{
			"query_arg":  aws.StringValue(profile.QueryArg),
			"profile_id": aws.StringValue(profile.ProfileId),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
