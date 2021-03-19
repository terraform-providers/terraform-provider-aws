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

func resourceAwsLightsailDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLightsailDomainCreate,
		Read:   resourceAwsLightsailDomainRead,
		Update: resourceAwsLightsailDomainUpdate,
		Delete: resourceAwsLightsailDomainDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsLightsailDomainCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn

	req := lightsail.CreateDomainInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		req.Tags = keyvaluetags.New(v).IgnoreAws().LightsailTags()
	}

	_, err := conn.CreateDomain(&req)

	if err != nil {
		return err
	}

	d.SetId(d.Get("domain_name").(string))

	return resourceAwsLightsailDomainRead(d, meta)
}

func resourceAwsLightsailDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.GetDomain(&lightsail.GetDomainInput{
		DomainName: aws.String(d.Id()),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NotFoundException" {
				log.Printf("[WARN] Lightsail Domain (%s) not found, removing from state", d.Id())
				d.SetId("")
				return nil
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

func resourceAwsLightsailDomainDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn
	_, err := conn.DeleteDomain(&lightsail.DeleteDomainInput{
		DomainName: aws.String(d.Id()),
	})

	return err
}

func resourceAwsLightsailDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.LightsailUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Lightsail domain (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsLightsailDomainRead(d, meta)
}
