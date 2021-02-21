package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/signer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsSignerSigningProfile() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsSignerSigningProfileRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform_display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"revocation_record": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"revocation_effective_from": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"revoked_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"revoked_by": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"signature_validity_period": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsSignerSigningProfileRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).signerconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	profileName := d.Get("name").(string)
	signingProfileOutput, err := conn.GetSigningProfile(&signer.GetSigningProfileInput{
		ProfileName: aws.String(profileName),
	})

	if err != nil {
		return fmt.Errorf("error reading Signer signing profile (%s): %w", d.Id(), err)
	}

	if err := d.Set("platform_id", signingProfileOutput.PlatformId); err != nil {
		return fmt.Errorf("error setting signer signing profile platform id: %w", err)
	}

	if err := d.Set("signature_validity_period", []interface{}{
		map[string]interface{}{
			"value": signingProfileOutput.SignatureValidityPeriod.Value,
			"type":  signingProfileOutput.SignatureValidityPeriod.Type,
		},
	}); err != nil {
		return fmt.Errorf("error setting signer signing profile signature validity period: %w", err)
	}

	if err := d.Set("platform_display_name", signingProfileOutput.PlatformDisplayName); err != nil {
		return fmt.Errorf("error setting signer signing profile platform display name: %w", err)
	}

	if err := d.Set("arn", signingProfileOutput.Arn); err != nil {
		return fmt.Errorf("error setting signer signing profile arn: %w", err)
	}

	if err := d.Set("version", signingProfileOutput.ProfileVersion); err != nil {
		return fmt.Errorf("error setting signer signing profile version: %w", err)
	}

	if err := d.Set("version_arn", signingProfileOutput.ProfileVersionArn); err != nil {
		return fmt.Errorf("error setting signer signing profile version arn: %w", err)
	}

	if err := d.Set("status", signingProfileOutput.Status); err != nil {
		return fmt.Errorf("error setting signer signing profile status: %w", err)
	}

	if err := d.Set("tags", keyvaluetags.SignerKeyValueTags(signingProfileOutput.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting signer signing profile tags: %w", err)
	}

	if err := d.Set("revocation_record", flattenSignerSigningProfileRevocationRecord(signingProfileOutput.RevocationRecord)); err != nil {
		return fmt.Errorf("error setting signer signing profile revocation record: %w", err)
	}

	d.SetId(aws.StringValue(signingProfileOutput.ProfileName))

	return nil
}
