package aws

import (
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/swf"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsSwfDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSwfDomainCreate,
		Read:   resourceAwsSwfDomainRead,
		Update: resourceAwsSwfDomainUpdate,
		Delete: resourceAwsSwfDomainDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"workflow_execution_retention_period_in_days": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
					value, err := strconv.Atoi(v.(string))
					if err != nil || value > 90 || value < 0 {
						es = append(es, fmt.Errorf(
							"%q must be between 0 and 90 days inclusive", k))
					}
					return
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSwfDomainCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).swfconn

	var name string

	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		name = resource.PrefixedUniqueId(v.(string))
	} else {
		name = resource.UniqueId()
	}

	input := &swf.RegisterDomainInput{
		Name:                                   aws.String(name),
		WorkflowExecutionRetentionPeriodInDays: aws.String(d.Get("workflow_execution_retention_period_in_days").(string)),
		Tags:                                   keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().SwfTags(),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	_, err := conn.RegisterDomain(input)
	if err != nil {
		return err
	}

	d.SetId(name)

	return resourceAwsSwfDomainRead(d, meta)
}

func resourceAwsSwfDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).swfconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &swf.DescribeDomainInput{
		Name: aws.String(d.Id()),
	}

	resp, err := conn.DescribeDomain(input)
	if err != nil {
		if isAWSErr(err, swf.ErrCodeUnknownResourceFault, "") {
			log.Printf("[WARN] SWF Domain %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading SWF Domain: %s", err)
	}

	if resp == nil || resp.Configuration == nil || resp.DomainInfo == nil {
		log.Printf("[WARN] SWF Domain %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := *resp.DomainInfo.Arn
	tags, err := keyvaluetags.SwfListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for SWF Domain (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("arn", resp.DomainInfo.Arn)
	d.Set("name", resp.DomainInfo.Name)
	d.Set("description", resp.DomainInfo.Description)
	d.Set("workflow_execution_retention_period_in_days", resp.Configuration.WorkflowExecutionRetentionPeriodInDays)

	return nil
}

func resourceAwsSwfDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).swfconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.SwfUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating SWF Domain (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsSwfDomainRead(d, meta)
}

func resourceAwsSwfDomainDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).swfconn

	input := &swf.DeprecateDomainInput{
		Name: aws.String(d.Get("name").(string)),
	}

	_, err := conn.DeprecateDomain(input)
	if err != nil {
		if isAWSErr(err, swf.ErrCodeDomainDeprecatedFault, "") {
			return nil
		}
		if isAWSErr(err, swf.ErrCodeUnknownResourceFault, "") {
			return nil
		}
		return fmt.Errorf("error deleting SWF Domain: %s", err)
	}

	return nil
}
