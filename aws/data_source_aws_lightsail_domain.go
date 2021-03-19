package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsLightsailDomain() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLightsailDomainRead,

		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func dataSourceAwsLightsailDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.GetDomain(&lightsail.GetDomainInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NotFoundException" {
				log.Printf("[WARN] Lightsail Domain (%s) not found, removing from state", d.Get("domain_name"))
				return fmt.Errorf("no matching Lightsail Domain found")
			}
			return err
		}
		return err
	}

	domain := resp.Domain

	d.Set("arn", domain.Arn)
	d.Set("domain_name", domain.Name)
	d.SetId(d.Get("domain_name").(string))
	if err := d.Set("tags", keyvaluetags.LightsailKeyValueTags(domain.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}
