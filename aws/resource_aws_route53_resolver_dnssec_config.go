package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53resolver/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53resolver/waiter"
)

func resourceAwsRoute53ResolverDnssecConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53ResolverDnssecConfigCreate,
		Read:   resourceAwsRoute53ResolverDnssecConfigRead,
		Delete: resourceAwsRoute53ResolverDnssecConfigDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"resource_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"validation_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsRoute53ResolverDnssecConfigCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	req := &route53resolver.UpdateResolverDnssecConfigInput{
		ResourceId: aws.String(d.Get("resource_id").(string)),
		Validation: aws.String(route53resolver.ValidationEnable),
	}

	log.Printf("[DEBUG] Creating Route53 Resolver DNSSEC config: %#v", req)
	resp, err := conn.UpdateResolverDnssecConfig(req)
	if err != nil {
		return fmt.Errorf("error creating Route53 Resolver DNSSEC config: %w", err)
	}

	d.SetId(aws.StringValue(resp.ResolverDNSSECConfig.Id))

	_, err = waiter.DnssecConfigCreated(conn, d.Id())
	if err != nil {
		return err
	}

	return resourceAwsRoute53ResolverDnssecConfigRead(d, meta)
}

func resourceAwsRoute53ResolverDnssecConfigRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	config, err := finder.ResolverDnssecConfigByID(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error getting Route53 Resolver DNSSEC config (%s): %w", d.Id(), err)
	}

	if config == nil || aws.StringValue(config.ValidationStatus) == route53resolver.ResolverDNSSECValidationStatusDisabled {
		log.Printf("[WARN] Route53 Resolver DNSSEC config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("id", config.Id)
	d.Set("owner_id", config.OwnerId)
	d.Set("resource_id", config.ResourceId)
	d.Set("validation_status", config.ValidationStatus)

	configArn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "route53resolver",
		Region:    meta.(*AWSClient).region,
		AccountID: aws.StringValue(config.OwnerId),
		Resource:  fmt.Sprintf("resolver-dnssec-config/%s", aws.StringValue(config.ResourceId)),
	}.String()
	d.Set("arn", configArn)

	return nil
}

func resourceAwsRoute53ResolverDnssecConfigDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	log.Printf("[DEBUG] Deleting Route53 Resolver DNSSEC config: %s", d.Id())
	_, err := conn.UpdateResolverDnssecConfig(&route53resolver.UpdateResolverDnssecConfigInput{
		ResourceId: aws.String(d.Get("resource_id").(string)),
		Validation: aws.String(route53resolver.ValidationDisable),
	})
	if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting Route53 Resolver DNSSEC config (%s): %w", d.Id(), err)
	}

	_, err = waiter.DnssecConfigDeleted(conn, d.Id())
	if err != nil {
		return err
	}

	return nil
}
