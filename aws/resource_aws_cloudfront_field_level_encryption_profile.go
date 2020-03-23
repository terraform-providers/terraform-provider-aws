package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsCloudfrontFieldLevelEncryptionProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudfrontFieldLevelEncryptionProfileCreate,
		Read:   resourceAwsCloudfrontFieldLevelEncryptionProfileRead,
		Update: resourceAwsCloudfrontFieldLevelEncryptionProfileUpdate,
		Delete: resourceAwsCloudfrontFieldLevelEncryptionProfileDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"encryption_entities": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"public_key_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"provider_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"field_patterns": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
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

func resourceAwsCloudfrontFieldLevelEncryptionProfileCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	fl := &cloudfront.FieldLevelEncryptionProfileConfig{
		CallerReference:    aws.String(resource.UniqueId()),
		Name:               aws.String(d.Get("name").(string)),
		EncryptionEntities: expandAwsCloudfrontFieldLevelEncryptionProfileConfig(d.Get("encryption_entities").([]interface{})),
	}

	if v, ok := d.GetOk("comment"); ok {
		fl.Comment = aws.String(v.(string))
	}

	input := &cloudfront.CreateFieldLevelEncryptionProfileInput{
		FieldLevelEncryptionProfileConfig: fl,
	}

	resp, err := conn.CreateFieldLevelEncryptionProfile(input)
	if err != nil {
		return fmt.Errorf("error creating Cloudfront Field Level Encryption Profile (%s): %s", d.Id(), err)
	}

	d.SetId(*resp.FieldLevelEncryptionProfile.Id)

	return resourceAwsCloudfrontFieldLevelEncryptionProfileRead(d, meta)
}

func resourceAwsCloudfrontFieldLevelEncryptionProfileRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	input := &cloudfront.GetFieldLevelEncryptionProfileInput{
		Id: aws.String(d.Id()),
	}

	resp, err := conn.GetFieldLevelEncryptionProfile(input)
	if isAWSErr(err, cloudfront.ErrCodeNoSuchFieldLevelEncryptionProfile, "") {
		log.Printf("[WARN] Cloudfront Field Level Encryption Profile %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Cloudfront Field Level Encryption Profile (%s): %s", d.Id(), err)
	}
	profile := resp.FieldLevelEncryptionProfile.FieldLevelEncryptionProfileConfig
	d.Set("etag", resp.ETag)
	d.Set("comment", profile.Comment)
	d.Set("name", profile.Name)
	d.Set("caller_reference", profile.CallerReference)

	if err := d.Set("encryption_entities", flattenAwsCloudfrontFieldLevelEncryptionProfileEncryptionEntitiesConfig(profile.EncryptionEntities)); err != nil {
		return fmt.Errorf("error setting encryption_entities %s", err)
	}

	return nil
}

func resourceAwsCloudfrontFieldLevelEncryptionProfileUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	fl := &cloudfront.FieldLevelEncryptionProfileConfig{
		CallerReference:    aws.String(d.Get("caller_reference").(string)),
		Name:               aws.String(d.Get("name").(string)),
		EncryptionEntities: expandAwsCloudfrontFieldLevelEncryptionProfileConfig(d.Get("encryption_entities").([]interface{})),
	}

	if v, ok := d.GetOk("comment"); ok {
		fl.Comment = aws.String(v.(string))
	}

	input := &cloudfront.UpdateFieldLevelEncryptionProfileInput{
		FieldLevelEncryptionProfileConfig: fl,
		Id:                                aws.String(d.Id()),
		IfMatch:                           aws.String(d.Get("etag").(string)),
	}

	_, err := conn.UpdateFieldLevelEncryptionProfile(input)
	if err != nil {
		return fmt.Errorf("error creating Cloudfront Field Level Encryption Profile (%s): %s", d.Id(), err)
	}

	return resourceAwsCloudfrontFieldLevelEncryptionProfileRead(d, meta)
}

func resourceAwsCloudfrontFieldLevelEncryptionProfileDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	input := &cloudfront.DeleteFieldLevelEncryptionProfileInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	}

	_, err := conn.DeleteFieldLevelEncryptionProfile(input)
	if err != nil {
		return fmt.Errorf("error deleting Cloudfront Field Level Encryption Profile (%s): %s", d.Id(), err)
	}

	return nil
}
func expandAwsCloudfrontFieldLevelEncryptionProfileConfig(config []interface{}) *cloudfront.EncryptionEntities {
	entities := make([]*cloudfront.EncryptionEntity, 0)

	for _, raw := range config {
		m := raw.(map[string]interface{})

		entity := &cloudfront.EncryptionEntity{
			PublicKeyId:   aws.String(m["public_key_id"].(string)),
			ProviderId:    aws.String(m["provider_id"].(string)),
			FieldPatterns: expandAwsCloudfrontFieldLevelEncryptionProfileFieldPatternsConfig(m["field_patterns"].(*schema.Set)),
		}

		entities = append(entities, entity)
	}

	contentTypeProfiles := &cloudfront.EncryptionEntities{
		Quantity: aws.Int64(int64(len(config))),
		Items:    entities,
	}

	return contentTypeProfiles
}

func expandAwsCloudfrontFieldLevelEncryptionProfileFieldPatternsConfig(config *schema.Set) *cloudfront.FieldPatterns {
	contentTypeProfiles := &cloudfront.FieldPatterns{
		Quantity: aws.Int64(int64(config.Len())),
		Items:    expandStringSet(config),
	}

	return contentTypeProfiles
}

func flattenAwsCloudfrontFieldLevelEncryptionProfileEncryptionEntitiesConfig(config *cloudfront.EncryptionEntities) []map[string]interface{} {
	result := make([]map[string]interface{}, len(config.Items))

	for i, s := range config.Items {
		m := make(map[string]interface{})
		m["provider_id"] = aws.StringValue(s.ProviderId)
		m["public_key_id"] = aws.StringValue(s.PublicKeyId)
		m["field_patterns"] = flattenStringSet(s.FieldPatterns.Items)
		result[i] = m
	}

	return result
}
