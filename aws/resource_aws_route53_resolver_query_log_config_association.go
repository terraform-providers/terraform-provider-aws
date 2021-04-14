package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53resolver/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53resolver/waiter"
)

func resourceAwsRoute53ResolverQueryLogConfigAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53ResolverQueryLogConfigAssociationCreate,
		Read:   resourceAwsRoute53ResolverQueryLogConfigAssociationRead,
		Delete: resourceAwsRoute53ResolverQueryLogConfigAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"resolver_query_log_config_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsRoute53ResolverQueryLogConfigAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	input := &route53resolver.AssociateResolverQueryLogConfigInput{
		ResolverQueryLogConfigId: aws.String(d.Get("resolver_query_log_config_id").(string)),
		ResourceId:               aws.String(d.Get("resource_id").(string)),
	}

	log.Printf("[DEBUG] Creating Route53 Resolver Query Log Config Association: %s", input)
	output, err := conn.AssociateResolverQueryLogConfig(input)

	if err != nil {
		return fmt.Errorf("error creating Route53 Resolver Query Log Config Association: %w", err)
	}

	d.SetId(aws.StringValue(output.ResolverQueryLogConfigAssociation.Id))

	_, err = waiter.QueryLogConfigAssociationCreated(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for Route53 Resolver Query Log Config Association (%s) to become available: %w", d.Id(), err)
	}

	return resourceAwsRoute53ResolverQueryLogConfigAssociationRead(d, meta)
}

func resourceAwsRoute53ResolverQueryLogConfigAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	queryLogConfigAssociation, err := finder.ResolverQueryLogConfigAssociationByID(conn, d.Id())

	if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Route53 Resolver Query Log Config Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Route53 Resolver Query Log Config Association (%s): %w", d.Id(), err)
	}

	if queryLogConfigAssociation == nil {
		log.Printf("[WARN] Route53 Resolver Query Log Config Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("resolver_query_log_config_id", queryLogConfigAssociation.ResolverQueryLogConfigId)
	d.Set("resource_id", queryLogConfigAssociation.ResourceId)

	return nil
}

func resourceAwsRoute53ResolverQueryLogConfigAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	log.Printf("[DEBUG] Deleting Route53 Resolver Query Log Config Association (%s)", d.Id())
	_, err := conn.DisassociateResolverQueryLogConfig(&route53resolver.DisassociateResolverQueryLogConfigInput{
		ResolverQueryLogConfigId: aws.String(d.Get("resolver_query_log_config_id").(string)),
		ResourceId:               aws.String(d.Get("resource_id").(string)),
	})

	if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Route53 Resolver Query Log Config Association (%s): %w", d.Id(), err)
	}

	_, err = waiter.QueryLogConfigAssociationDeleted(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for Route53 Resolver Query Log Config Association (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}
