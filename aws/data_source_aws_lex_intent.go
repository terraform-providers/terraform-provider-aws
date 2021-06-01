package aws

import (
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsLexIntent() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLexIntentRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexp.MustCompile(`^([A-Za-z]_?)+$`), ""),
				),
			},
			"parent_intent_signature": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  LexIntentVersionLatest,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`\$LATEST|[0-9]+`), ""),
				),
			},
		},
	}
}

func dataSourceAwsLexIntentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	intentName := d.Get("name").(string)
	resp, err := conn.GetIntent(&lexmodelbuildingservice.GetIntentInput{
		Name:    aws.String(intentName),
		Version: aws.String(d.Get("version").(string)),
	})
	if err != nil {
		return fmt.Errorf("error getting intent %s: %w", intentName, err)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "lex",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("intent:%s", d.Get("name").(string)),
	}
	d.Set("arn", arn.String())

	d.Set("checksum", resp.Checksum)
	d.Set("created_date", resp.CreatedDate.Format(time.RFC3339))
	d.Set("description", resp.Description)
	d.Set("last_updated_date", resp.LastUpdatedDate.Format(time.RFC3339))
	d.Set("name", resp.Name)
	d.Set("parent_intent_signature", resp.ParentIntentSignature)
	d.Set("version", resp.Version)

	d.SetId(intentName)

	return nil
}
