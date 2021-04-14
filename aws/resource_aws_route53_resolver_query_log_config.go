package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53resolver/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53resolver/waiter"
)

func resourceAwsRoute53ResolverQueryLogConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53ResolverQueryLogConfigCreate,
		Read:   resourceAwsRoute53ResolverQueryLogConfigRead,
		Update: resourceAwsRoute53ResolverQueryLogConfigUpdate,
		Delete: resourceAwsRoute53ResolverQueryLogConfigDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"destination_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateRoute53ResolverName,
			},

			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"share_status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsRoute53ResolverQueryLogConfigCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	input := &route53resolver.CreateResolverQueryLogConfigInput{
		CreatorRequestId: aws.String(resource.PrefixedUniqueId("tf-r53-resolver-query-log-config-")),
		DestinationArn:   aws.String(d.Get("destination_arn").(string)),
		Name:             aws.String(d.Get("name").(string)),
	}
	if v, ok := d.GetOk("tags"); ok && len(v.(map[string]interface{})) > 0 {
		input.Tags = keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().Route53resolverTags()
	}

	log.Printf("[DEBUG] Creating Route53 Resolver Query Log Config: %s", input)
	output, err := conn.CreateResolverQueryLogConfig(input)

	if err != nil {
		return fmt.Errorf("error creating Route53 Resolver Query Log Config: %w", err)
	}

	d.SetId(aws.StringValue(output.ResolverQueryLogConfig.Id))

	_, err = waiter.QueryLogConfigCreated(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for Route53 Resolver Query Log Config (%s) to become available: %w", d.Id(), err)
	}

	return resourceAwsRoute53ResolverQueryLogConfigRead(d, meta)
}

func resourceAwsRoute53ResolverQueryLogConfigRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	queryLogConfig, err := finder.ResolverQueryLogConfigByID(conn, d.Id())

	if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Route53 Resolver Query Log Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Route53 Resolver Query Log Config (%s): %w", d.Id(), err)
	}

	if queryLogConfig == nil {
		log.Printf("[WARN] Route53 Resolver Query Log Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := aws.StringValue(queryLogConfig.Arn)
	d.Set("arn", arn)
	d.Set("destination_arn", queryLogConfig.DestinationArn)
	d.Set("name", queryLogConfig.Name)
	d.Set("owner_id", queryLogConfig.OwnerId)
	d.Set("share_status", queryLogConfig.ShareStatus)

	tags, err := keyvaluetags.Route53resolverListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("error listing tags for Route53 Resolver Query Log Config (%s): %w", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsRoute53ResolverQueryLogConfigUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Route53resolverUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Route53 Resolver Query Log Config (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return resourceAwsRoute53ResolverQueryLogConfigRead(d, meta)
}

func resourceAwsRoute53ResolverQueryLogConfigDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	log.Printf("[DEBUG] Deleting Route53 Resolver Query Log Config (%s)", d.Id())
	_, err := conn.DeleteResolverQueryLogConfig(&route53resolver.DeleteResolverQueryLogConfigInput{
		ResolverQueryLogConfigId: aws.String(d.Id()),
	})

	if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Route53 Resolver Query Log Config (%s): %w", d.Id(), err)
	}

	_, err = waiter.QueryLogConfigDeleted(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for Route53 Resolver Query Log Config (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}
