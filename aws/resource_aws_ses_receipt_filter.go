package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsSesReceiptFilter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSesReceiptFilterCreate,
		Read:   resourceAwsSesReceiptFilterRead,
		Delete: resourceAwsSesReceiptFilterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-zA-Z_-]+$`), "must contain only alphanumeric, underscore, and hyphen characters"),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-zA-Z]`), "must begin with a alphanumeric character"),
					validation.StringMatch(regexp.MustCompile(`[0-9a-zA-Z]$`), "must end with a alphanumeric character"),
				),
			},

			"cidr": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					validation.IsCIDR,
					validation.IsIPv4Address,
				),
			},

			"policy": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					ses.ReceiptFilterPolicyBlock,
					ses.ReceiptFilterPolicyAllow,
				}, false),
			},
		},
	}
}

func resourceAwsSesReceiptFilterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	name := d.Get("name").(string)

	createOpts := &ses.CreateReceiptFilterInput{
		Filter: &ses.ReceiptFilter{
			Name: aws.String(name),
			IpFilter: &ses.ReceiptIpFilter{
				Cidr:   aws.String(d.Get("cidr").(string)),
				Policy: aws.String(d.Get("policy").(string)),
			},
		},
	}

	_, err := conn.CreateReceiptFilter(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating SES receipt filter: %s", err)
	}

	d.SetId(name)

	return resourceAwsSesReceiptFilterRead(d, meta)
}

func resourceAwsSesReceiptFilterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	listOpts := &ses.ListReceiptFiltersInput{}

	response, err := conn.ListReceiptFilters(listOpts)
	if err != nil {
		return err
	}

	var filter *ses.ReceiptFilter

	for _, responseFilter := range response.Filters {
		if aws.StringValue(responseFilter.Name) == d.Id() {
			filter = responseFilter
			break
		}
	}

	if filter == nil {
		log.Printf("[WARN] SES Receipt Filter (%s) not found", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("cidr", filter.IpFilter.Cidr)
	d.Set("policy", filter.IpFilter.Policy)
	d.Set("name", filter.Name)

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "ses",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("receipt-filter/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	return nil
}

func resourceAwsSesReceiptFilterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	deleteOpts := &ses.DeleteReceiptFilterInput{
		FilterName: aws.String(d.Id()),
	}

	_, err := conn.DeleteReceiptFilter(deleteOpts)
	if err != nil {
		return fmt.Errorf("Error deleting SES receipt filter: %s", err)
	}

	return nil
}
