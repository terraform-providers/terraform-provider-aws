package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsGlueDataCatalogEncryptionSettings() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAwsGlueDataCatalogEncryptionSettingsRead,
		Schema: map[string]*schema.Schema{
			"catalog_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"data_catalog_encryption_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connection_password_encryption": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"aws_kms_key_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"return_connection_password_encrypted": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
						"encryption_at_rest": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"catalog_encryption_mode": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"sse_aws_kms_key_id": {
										Type:     schema.TypeString,
										Computed: true,
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

func dataSourceAwsGlueDataCatalogEncryptionSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).glueconn
	id := d.Get("catalog_id").(string)
	input := &glue.GetDataCatalogEncryptionSettingsInput{
		CatalogId: aws.String(id),
	}
	out, err := conn.GetDataCatalogEncryptionSettings(input)
	if err != nil {
		return diag.Errorf("Error reading Glue Data Catalog Encryption Settings: %s", err)
	}
	d.SetId(id)
	d.Set("catalog_id", d.Id())

	if err := d.Set("data_catalog_encryption_settings", flattenGlueDataCatalogEncryptionSettings(out.DataCatalogEncryptionSettings)); err != nil {
		return diag.Errorf("error setting data_catalog_encryption_settings: %s", err)
	}
	return nil
}
