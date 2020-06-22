package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsTransferServer() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsTransferServerRead,
		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"invocation_role": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identity_provider_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"logging_role": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsTransferServerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).transferconn

	serverID := d.Get("server_id").(string)
	input := &transfer.DescribeServerInput{
		ServerId: aws.String(serverID),
	}

	log.Printf("[DEBUG] Describe Transfer Server Option: %#v", input)

	resp, err := conn.DescribeServer(input)
	if err != nil {
		return fmt.Errorf("error describing Transfer Server (%s): %s", serverID, err)
	}

	endpoint := meta.(*AWSClient).RegionalHostname(fmt.Sprintf("%s.server.transfer", serverID))

	d.SetId(serverID)
	d.Set("arn", resp.Server.Arn)
	d.Set("endpoint", endpoint)
	if resp.Server.IdentityProviderDetails != nil {
		d.Set("invocation_role", aws.StringValue(resp.Server.IdentityProviderDetails.InvocationRole))
		d.Set("url", aws.StringValue(resp.Server.IdentityProviderDetails.Url))
	}
	d.Set("identity_provider_type", resp.Server.IdentityProviderType)
	d.Set("logging_role", resp.Server.LoggingRole)

	return nil
}
