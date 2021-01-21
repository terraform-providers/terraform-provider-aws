package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsIamAccountAlias() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsIamAccountAliasRead,

		Schema: map[string]*schema.Schema{
			"account_alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsIamAccountAliasRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	log.Printf("[DEBUG] Reading IAM Account Aliases.")

	req := &iam.ListAccountAliasesInput{}
	resp, err := conn.ListAccountAliases(req)
	if err != nil {
		return err
	}

	// 'AccountAliases': [] if there is no alias.
	if resp == nil || len(resp.AccountAliases) == 0 {
		return fmt.Errorf("no IAM account alias found")
	}

	alias := aws.StringValue(resp.AccountAliases[0])
	d.SetId(alias)
	log.Printf("[DEBUG] Setting AWS IAM Account Alias to %s.", alias)
	d.Set("account_alias", alias)

	return nil
}
