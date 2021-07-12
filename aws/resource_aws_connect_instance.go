package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/connect/waiter"
)

var resourceConnectInstanceAttributesMapping = map[string]string{
	connect.InstanceAttributeTypeInboundCalls:          "inbound_calls_enabled",
	connect.InstanceAttributeTypeOutboundCalls:         "outbound_calls_enabled",
	connect.InstanceAttributeTypeContactflowLogs:       "contact_flow_logs_enabled",
	connect.InstanceAttributeTypeContactLens:           "contact_lens_enabled",
	connect.InstanceAttributeTypeAutoResolveBestVoices: "auto_resolve_best_voices",
	connect.InstanceAttributeTypeUseCustomTtsVoices:    "use_custom_tts_voices",
	connect.InstanceAttributeTypeEarlyMedia:            "early_media_enabled",
}

func resourceAwsConnectInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsConnectInstanceCreate,
		ReadContext:   resourceAwsConnectInstanceRead,
		UpdateContext: resourceAwsConnectInstanceUpdate,
		DeleteContext: resourceAwsConnectInstanceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"identity_management_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      connect.DirectoryTypeConnectManaged,
				ValidateFunc: validation.StringInSlice(connect.DirectoryType_Values(), false),
			},
			"directory_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(12, 12),
			},
			"instance_alias": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^([\da-zA-Z]+)([\da-zA-Z-]+)$`), "must contain only alphanumeric hyphen and underscore characters"),
					validation.StringDoesNotMatch(regexp.MustCompile(`^(d-).+$`), "can not start with d-"),
				),
			},
			"inbound_calls_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"outbound_calls_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"early_media_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"contact_flow_logs_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"contact_lens_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"auto_resolve_best_voices": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"use_custom_tts_voices": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_role": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsConnectInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn

	input := &connect.CreateInstanceInput{
		ClientToken:            aws.String(resource.UniqueId()),
		IdentityManagementType: aws.String(d.Get("identity_management_type").(string)),
		InstanceAlias:          aws.String(d.Get("instance_alias").(string)),
		InboundCallsEnabled:    aws.Bool(d.Get("inbound_calls_enabled").(bool)),
		OutboundCallsEnabled:   aws.Bool(d.Get("outbound_calls_enabled").(bool)),
	}

	if _, ok := d.GetOk("directory_id"); ok {
		input.DirectoryId = aws.String(d.Get("directory_id").(string))
	}

	log.Printf("[DEBUG] Creating Connect Instance %s", input)

	output, err := conn.CreateInstanceWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Connect Instance (%s): %s", d.Id(), err))
	}

	d.SetId(aws.StringValue(output.Id))

	if _, err := waiter.InstanceCreated(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Connect instance creation (%s): %s", d.Id(), err))
	}

	for att := range resourceConnectInstanceAttributesMapping {
		rKey := resourceConnectInstanceAttributesMapping[att]
		val := d.Get(rKey)
		err := resourceAwsConnectInstanceUpdateAttribute(ctx, conn, d.Id(), att, strconv.FormatBool(val.(bool)))
		if err != nil {
			return diag.FromErr(fmt.Errorf("error setting Connect instance (%s) attribute (%s): %s", d.Id(), att, err))
		}
	}

	return resourceAwsConnectInstanceRead(ctx, d, meta)
}

func resourceAwsConnectInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn

	for att := range resourceConnectInstanceAttributesMapping {
		rKey := resourceConnectInstanceAttributesMapping[att]
		if d.HasChange(rKey) {
			_, n := d.GetChange(rKey)
			err := resourceAwsConnectInstanceUpdateAttribute(ctx, conn, d.Id(), att, strconv.FormatBool(n.(bool)))
			if err != nil {
				return diag.FromErr(fmt.Errorf("error updating Connect instance (%s) attribute (%s): %s", d.Id(), att, err))
			}
		}
	}

	return nil
}
func resourceAwsConnectInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn

	input := connect.DescribeInstanceInput{
		InstanceId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Connect Instance %s", d.Id())

	output, err := conn.DescribeInstanceWithContext(ctx, &input)

	if isAWSErr(err, connect.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Connect Instance (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Connect Instance (%s): %s", d.Id(), err))
	}

	instance := output.Instance

	d.SetId(aws.StringValue(instance.Id))

	d.Set("arn", instance.Arn)
	d.Set("created_time", instance.CreatedTime.Format(time.RFC3339))
	d.Set("identity_management_type", instance.IdentityManagementType)
	d.Set("instance_alias", instance.InstanceAlias)
	d.Set("inbound_calls_enabled", instance.InboundCallsEnabled)
	d.Set("outbound_calls_enabled", instance.OutboundCallsEnabled)
	d.Set("status", instance.InstanceStatus)
	d.Set("service_role", instance.ServiceRole)

	for att := range resourceConnectInstanceAttributesMapping {
		value, err := resourceAwsConnectInstanceReadAttribute(ctx, conn, d.Id(), att)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error reading Connect instance (%s) attribute (%s): %s", d.Id(), att, err))
		}
		d.Set(resourceConnectInstanceAttributesMapping[att], value)
	}

	return nil
}

func resourceAwsConnectInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn

	input := &connect.DeleteInstanceInput{
		InstanceId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Connect Instance %s", d.Id())

	_, err := conn.DeleteInstance(input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Connect Instance (%s): %s", d.Id(), err))
	}

	// For using an existing directory,
	// There is no proper way. If the Connect instance was unregistered from the attached directory.
	// We don't have a PENDING_DELETION or DELETED for the Connect instance.
	// Deleting the directory immediately after removing the connect instance
	// will cause an error because it is still has authorized applications.
	// There is no specific wait time for this check, so we'll wait a few seconds
	// to make sure the Connect instance will deregister.

	imt := d.Get("identity_management_type")

	if imt.(string) == connect.DirectoryTypeExistingDirectory {
		log.Print("[INFO] Waiting for Connect to deregister from the Directory Service")
		time.Sleep(30 * time.Second)
	}

	return nil
}

func resourceAwsConnectInstanceUpdateAttribute(ctx context.Context, conn *connect.Connect, instanceID string, attributeType string, value string) error {
	input := &connect.UpdateInstanceAttributeInput{
		InstanceId:    aws.String(instanceID),
		AttributeType: aws.String(attributeType),
		Value:         aws.String(value),
	}

	_, err := conn.UpdateInstanceAttributeWithContext(ctx, input)

	if err != nil {
		return err
	}

	return nil
}

func resourceAwsConnectInstanceReadAttribute(ctx context.Context, conn *connect.Connect, instanceID string, attributeType string) (bool, error) {
	input := &connect.DescribeInstanceAttributeInput{
		InstanceId:    aws.String(instanceID),
		AttributeType: aws.String(attributeType),
	}

	output, err := conn.DescribeInstanceAttributeWithContext(ctx, input)

	if err != nil {
		return false, err
	}

	result, parseerr := strconv.ParseBool(*output.Attribute.Value)
	return result, parseerr
}
