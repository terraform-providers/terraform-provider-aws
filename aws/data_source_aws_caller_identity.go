package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsCallerIdentity() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCallerIdentityRead,

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsCallerIdentityRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).stsconn

	log.Printf("[DEBUG] Reading Caller Identity")
	res, err := client.GetCallerIdentity(&sts.GetCallerIdentityInput{})

	if err != nil {
		return fmt.Errorf("Error getting Caller Identity: %w", err)
	}

	log.Printf("[DEBUG] Received Caller Identity: %s", res)

	d.SetId(aws.StringValue(res.Account))
	d.Set("account_id", res.Account)
	d.Set("arn", res.Arn)
	d.Set("user_id", res.UserId)

	return nil
}
