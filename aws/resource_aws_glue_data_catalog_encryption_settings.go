package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsGlueDataCatalogEncryptionSettings() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGlueDataCatalogEncryptionSettingsPut,
		Read:   resourceAwsGlueDataCatalogEncryptionSettingsRead,
		Update: resourceAwsGlueDataCatalogEncryptionSettingsPut,
		Delete: schema.Noop,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"catalog_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"data_catalog_encryption_settings": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connection_password_encryption": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"aws_kms_key_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateArn,
									},
									"return_connection_password_encrypted": {
										Type:     schema.TypeBool,
										Required: true,
									},
								},
							},
						},
						"encryption_at_rest": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"catalog_encryption_mode": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(glue.CatalogEncryptionMode_Values(), false),
									},
									"sse_aws_kms_key_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateArn,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceAwsGlueDataCatalogEncryptionSettingsPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn
	catalogID := createAwsGlueCatalogID(d, meta.(*AWSClient).accountid)

	input := &glue.PutDataCatalogEncryptionSettingsInput{
		CatalogId:                     aws.String(catalogID),
		DataCatalogEncryptionSettings: expandGlueDataCatalogEncryptionSettings(d.Get("data_catalog_encryption_settings").([]interface{})),
	}

	_, err := conn.PutDataCatalogEncryptionSettings(input)
	if err != nil {
		return fmt.Errorf("Error setting Data Catalog Encryption Settings: %w", err)
	}

	d.SetId(catalogID)

	return resourceAwsGlueDataCatalogEncryptionSettingsRead(d, meta)
}

func resourceAwsGlueDataCatalogEncryptionSettingsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	input := &glue.GetDataCatalogEncryptionSettingsInput{
		CatalogId: aws.String(d.Id()),
	}

	out, err := conn.GetDataCatalogEncryptionSettings(input)
	if err != nil {
		return fmt.Errorf("Error reading Glue Data Catalog Encryption Settings: %w", err)
	}

	d.Set("catalog_id", d.Id())

	if err := d.Set("data_catalog_encryption_settings", flattenGlueDataCatalogEncryptionSettings(out.DataCatalogEncryptionSettings)); err != nil {
		return fmt.Errorf("error setting data_catalog_encryption_settings: %w", err)
	}

	return nil
}

func expandGlueDataCatalogEncryptionSettings(settings []interface{}) *glue.DataCatalogEncryptionSettings {
	m := settings[0].(map[string]interface{})

	target := &glue.DataCatalogEncryptionSettings{
		ConnectionPasswordEncryption: expandGlueDataCatalogConnectionPasswordEncryption(m["connection_password_encryption"].([]interface{})),
		EncryptionAtRest:             expandGlueDataCatalogEncryptionAtRest(m["encryption_at_rest"].([]interface{})),
	}

	return target
}

func flattenGlueDataCatalogEncryptionSettings(settings *glue.DataCatalogEncryptionSettings) []map[string]interface{} {
	m := map[string]interface{}{
		"connection_password_encryption": flattenGlueDataCatalogConnectionPasswordEncryption(settings.ConnectionPasswordEncryption),
		"encryption_at_rest":             flattenGlueDataCatalogEncryptionAtRest(settings.EncryptionAtRest),
	}

	return []map[string]interface{}{m}
}

func expandGlueDataCatalogConnectionPasswordEncryption(settings []interface{}) *glue.ConnectionPasswordEncryption {
	m := settings[0].(map[string]interface{})

	target := &glue.ConnectionPasswordEncryption{
		ReturnConnectionPasswordEncrypted: aws.Bool(m["return_connection_password_encrypted"].(bool)),
	}

	if v, ok := m["aws_kms_key_id"].(string); ok && v != "" {
		target.AwsKmsKeyId = aws.String(v)
	}

	return target
}

func flattenGlueDataCatalogConnectionPasswordEncryption(settings *glue.ConnectionPasswordEncryption) []map[string]interface{} {
	m := map[string]interface{}{
		"return_connection_password_encrypted": aws.BoolValue(settings.ReturnConnectionPasswordEncrypted),
	}

	if settings.AwsKmsKeyId != nil {
		m["aws_kms_key_id"] = aws.StringValue(settings.AwsKmsKeyId)
	}

	return []map[string]interface{}{m}
}

func expandGlueDataCatalogEncryptionAtRest(settings []interface{}) *glue.EncryptionAtRest {
	m := settings[0].(map[string]interface{})

	target := &glue.EncryptionAtRest{
		CatalogEncryptionMode: aws.String(m["catalog_encryption_mode"].(string)),
	}

	if v, ok := m["sse_aws_kms_key_id"].(string); ok && v != "" {
		target.SseAwsKmsKeyId = aws.String(v)
	}

	return target
}

func flattenGlueDataCatalogEncryptionAtRest(settings *glue.EncryptionAtRest) []map[string]interface{} {
	m := map[string]interface{}{
		"catalog_encryption_mode": aws.StringValue(settings.CatalogEncryptionMode),
	}

	if settings.SseAwsKmsKeyId != nil {
		m["sse_aws_kms_key_id"] = aws.StringValue(settings.SseAwsKmsKeyId)
	}

	return []map[string]interface{}{m}
}
