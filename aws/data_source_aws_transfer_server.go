package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/transfer/finder"
)

func dataSourceAwsTransferServer() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"certificate": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"endpoint_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"identity_provider_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"invocation_role": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"logging_role": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"protocols": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"security_policy_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"server_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		Read: dataSourceAwsTransferServerRead,
	}
}

func dataSourceAwsTransferServerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).transferconn

	serverID := d.Get("server_id").(string)

	output, err := finder.ServerByID(conn, serverID)

	if err != nil {
		return fmt.Errorf("error reading Transfer Server (%s): %w", serverID, err)
	}

	d.SetId(aws.StringValue(output.ServerId))
	d.Set("arn", output.Arn)
	d.Set("certificate", output.Certificate)
	d.Set("endpoint", meta.(*AWSClient).RegionalHostname(fmt.Sprintf("%s.server.transfer", serverID)))
	d.Set("endpoint_type", output.EndpointType)
	d.Set("identity_provider_type", output.IdentityProviderType)
	if output.IdentityProviderDetails != nil {
		d.Set("invocation_role", output.IdentityProviderDetails.InvocationRole)
	} else {
		d.Set("invocation_role", "")
	}
	d.Set("logging_role", output.LoggingRole)
	d.Set("protocols", aws.StringValueSlice(output.Protocols))
	d.Set("security_policy_name", output.SecurityPolicyName)
	if output.IdentityProviderDetails != nil {
		d.Set("url", output.IdentityProviderDetails.Url)
	} else {
		d.Set("url", "")
	}

	return nil
}
