package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/qldb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsQLDBLedger() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsQLDBLedgerRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexp.MustCompile(`^[A-Za-z0-9_-]+`), "must contain only alphanumeric characters, underscores, and hyphens"),
				),
			},

			"deletion_protection": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsQLDBLedgerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).qldbconn

	target := d.Get("name")

	req := &qldb.DescribeLedgerInput{
		Name: aws.String(target.(string)),
	}

	log.Printf("[DEBUG] Reading QLDB Ledger: %s", req)
	resp, err := conn.DescribeLedger(req)

	if err != nil {
		return fmt.Errorf("Error describing ledger: %s", err)
	}

	d.SetId(aws.StringValue(resp.Name))
	d.Set("arn", resp.Arn)
	d.Set("deletion_protection", resp.DeletionProtection)

	return nil
}
