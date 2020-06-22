package aws

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const dataSourceAwsKmsSecretRemovedMessage = "This data source has been replaced with the `aws_kms_secrets` data source. Upgrade information is available at: https://www.terraform.io/docs/providers/aws/guides/version-2-upgrade.html#data-source-aws_kms_secret"

func dataSourceAwsKmsSecret() *schema.Resource {
	return &schema.Resource{
		Read: func(d *schema.ResourceData, meta interface{}) error {
			return errors.New(dataSourceAwsKmsSecretRemovedMessage)
		},

		Schema: map[string]*schema.Schema{
			"secret": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"payload": {
							Type:     schema.TypeString,
							Required: true,
						},
						"context": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"grant_tokens": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}
