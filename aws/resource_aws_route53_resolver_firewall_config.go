package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53resolver/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsRoute53ResolverFirewallConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53ResolverFirewallConfigCreate,
		Read:   resourceAwsRoute53ResolverFirewallConfigRead,
		Update: resourceAwsRoute53ResolverFirewallConfigUpdate,
		Delete: resourceAwsRoute53ResolverFirewallConfigDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"resource_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"firewall_fail_open": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(route53resolver.FirewallFailOpenStatus_Values(), false),
			},
		},
	}
}

func resourceAwsRoute53ResolverFirewallConfigCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	input := &route53resolver.UpdateFirewallConfigInput{
		ResourceId: aws.String(d.Get("resource_id").(string)),
	}

	if v, ok := d.GetOk("firewall_fail_open"); ok {
		input.FirewallFailOpen = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Route 53 Resolver DNS Firewall config: %#v", input)
	output, err := conn.UpdateFirewallConfig(input)
	if err != nil {
		return fmt.Errorf("error creating Route 53 Resolver DNS Firewall config: %w", err)
	}

	d.SetId(aws.StringValue(output.FirewallConfig.Id))

	return resourceAwsRoute53ResolverFirewallConfigRead(d, meta)
}

func resourceAwsRoute53ResolverFirewallConfigRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	config, err := finder.FirewallConfigByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route 53 Resolver DNS Firewall config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Route 53 Resolver DNS Firewall config (%s): %w", d.Id(), err)
	}

	d.Set("owner_id", config.OwnerId)
	d.Set("resource_id", config.ResourceId)
	d.Set("firewall_fail_open", config.FirewallFailOpen)

	return nil
}

func resourceAwsRoute53ResolverFirewallConfigUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	input := &route53resolver.UpdateFirewallConfigInput{
		ResourceId: aws.String(d.Get("resource_id").(string)),
	}

	if v, ok := d.GetOk("firewall_fail_open"); ok {
		input.FirewallFailOpen = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Updating Route 53 Resolver DNS Firewall config: %#v", input)
	_, err := conn.UpdateFirewallConfig(input)
	if err != nil {
		return fmt.Errorf("error creating Route 53 Resolver DNS Firewall config: %w", err)
	}

	return resourceAwsRoute53ResolverFirewallConfigRead(d, meta)
}

func resourceAwsRoute53ResolverFirewallConfigDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	log.Printf("[DEBUG] Deleting Route 53 Resolver DNS Firewall config")
	_, err := conn.UpdateFirewallConfig(&route53resolver.UpdateFirewallConfigInput{
		ResourceId:       aws.String(d.Get("resource_id").(string)),
		FirewallFailOpen: aws.String(route53resolver.FirewallFailOpenStatusDisabled),
	})

	if err != nil {
		return fmt.Errorf("error deleting Route 53 Resolver DNS Firewall config (%s): %w", d.Id(), err)
	}

	return nil
}
