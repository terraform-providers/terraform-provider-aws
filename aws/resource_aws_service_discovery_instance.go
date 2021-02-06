package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicediscovery/waiter"
	"log"
	"strings"
)

func resourceAwsServiceDiscoveryInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsServiceDiscoveryInstanceCreate,
		ReadContext:   resourceAwsServiceDiscoveryInstanceRead,
		UpdateContext: resourceAwsServiceDiscoveryInstanceUpdate,
		DeleteContext: resourceAwsServiceDiscoveryInstanceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceAwsServiceDiscoveryInstanceImport,
		},

		Schema: map[string]*schema.Schema{
			"service_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"attributes": {
				Type:     schema.TypeMap,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"creator_request_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
		},
	}
}

func resourceAwsServiceDiscoveryInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	conn := meta.(*AWSClient).sdconn

	input := &servicediscovery.RegisterInstanceInput{
		ServiceId:  aws.String(d.Get("service_id").(string)),
		InstanceId: aws.String(d.Get("instance_id").(string)),
		Attributes: stringMapToPointers(d.Get("attributes").(map[string]interface{})),
	}

	if v, ok := d.GetOk("creator_request_id"); ok {
		input.CreatorRequestId = aws.String(v.(string))
	}

	resp, err := conn.RegisterInstance(input)
	if err != nil {
		return diag.FromErr(err)
	}

	if resp != nil && resp.OperationId != nil {
		if _, err := waiter.OperationSuccess(conn, aws.StringValue(resp.OperationId)); err != nil {
			return diag.FromErr(fmt.Errorf("error waiting for Service Discovery Service Instance (%s) create: %w", d.Id(), err))
		}
	}

	d.SetId(d.Get("instance_id").(string))

	return resourceAwsServiceDiscoveryInstanceRead(ctx, d, meta)
}

func resourceAwsServiceDiscoveryInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).sdconn

	input := &servicediscovery.GetInstanceInput{
		ServiceId:  aws.String(d.Get("service_id").(string)),
		InstanceId: aws.String(d.Get("instance_id").(string)),
	}

	resp, err := conn.GetInstanceWithContext(ctx, input)
	if err != nil {
		if isAWSErr(err, servicediscovery.ErrCodeInstanceNotFound, "") {
			log.Printf("[WARN] Service Discovery Instance (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	attributes := resp.Instance.Attributes
	if _, ok := attributes["AWS_EC2_INSTANCE_ID"]; ok {
		delete(attributes, "AWS_INSTANCE_IPV4")
	}

	err = d.Set("attributes", aws.StringValueMap(attributes))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAwsServiceDiscoveryInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceAwsServiceDiscoveryInstanceCreate(ctx, d, meta)
}

func resourceAwsServiceDiscoveryInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).sdconn

	input := &servicediscovery.DeregisterInstanceInput{
		ServiceId:  aws.String(d.Get("service_id").(string)),
		InstanceId: aws.String(d.Get("instance_id").(string)),
	}

	resp, err := conn.DeregisterInstanceWithContext(ctx, input)

	if isAWSErr(err, servicediscovery.ErrCodeInstanceNotFound, "") {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("Error deregistering Service Discovery Instance (%s): %w", d.Id(), err))
	}

	if resp != nil && resp.OperationId != nil {
		if _, err := waiter.OperationSuccess(conn, aws.StringValue(resp.OperationId)); err != nil {
			return diag.FromErr(fmt.Errorf("Error waiting for Service Discovery Service Instance (%s) delete: %w", d.Id(), err))
		}
	}

	return nil
}

func resourceAwsServiceDiscoveryInstanceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format (%q), expected <service-id>/<instance-id>", d.Id())
	}

	serviceId := idParts[0]
	instanceId := idParts[1]

	d.Set("service_id", serviceId)
	d.Set("instance_id", instanceId)
	d.SetId(instanceId)

	return []*schema.ResourceData{d}, nil
}