package aws

import (
	"context"
	"log"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsSecurityHubStandardsControl() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsSecurityHubStandardsControlPut,
		ReadContext:   resourceAwsSecurityHubStandardsControlRead,
		UpdateContext: resourceAwsSecurityHubStandardsControlPut,
		DeleteContext: resourceAwsSecurityHubStandardsControlDelete,

		Schema: map[string]*schema.Schema{
			"standards_control_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"control_status": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(securityhub.ControlStatus_Values(), false),
			},
			"disabled_reason": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"control_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"control_status_updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"related_requirements": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"remediation_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"severity_rating": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"title": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsSecurityHubStandardsControlRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).securityhubconn

	a := d.Get("standards_control_arn").(string)
	controlArn, err := arn.Parse(a)
	if err != nil {
		return diag.Errorf("error parsing standards control ARN %q", controlArn)
	}

	subscriptionArn := path.Dir(strings.ReplaceAll(controlArn.String(), "control", "subscription"))

	log.Printf("[DEBUG] Read Security Hub standard control %s", controlArn)

	resp, err := conn.DescribeStandardsControls(&securityhub.DescribeStandardsControlsInput{
		StandardsSubscriptionArn: aws.String(subscriptionArn),
	})
	if err != nil {
		return diag.Errorf("error reading Security Hub standard controls within %q subscription: %s", subscriptionArn, err)
	}

	for _, c := range resp.Controls {
		if aws.StringValue(c.StandardsControlArn) != controlArn.String() {
			continue
		}

		d.Set("control_status", c.ControlStatus)
		d.Set("control_status_updated_at", c.ControlStatusUpdatedAt.Format(time.RFC3339))
		d.Set("description", c.Description)
		d.Set("disabled_reason", c.DisabledReason)
		d.Set("severity_rating", c.SeverityRating)
		d.Set("title", c.Title)

		if err := d.Set("related_requirements", flattenStringList(c.RelatedRequirements)); err != nil {
			return diag.Errorf("error setting related_requirements: %s", err)
		}
	}

	return nil
}

func resourceAwsSecurityHubStandardsControlPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).securityhubconn

	d.SetId(d.Get("standards_control_arn").(string))

	log.Printf("[DEBUG] Update Security Hub standard control %s", d.Id())

	_, err := conn.UpdateStandardsControl(&securityhub.UpdateStandardsControlInput{
		StandardsControlArn: aws.String(d.Get("standards_control_arn").(string)),
		ControlStatus:       aws.String(d.Get("control_status").(string)),
		DisabledReason:      aws.String(d.Get("disabled_reason").(string)),
	})
	if err != nil {
		d.SetId("")
		return diag.Errorf("error updating Security Hub standard control %q: %s", d.Id(), err)
	}

	return resourceAwsSecurityHubStandardsControlRead(ctx, d, meta)
}

func resourceAwsSecurityHubStandardsControlDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[WARN] Cannot delete Security Hub standard control. Terraform will remove this resource from the state.")
	return nil
}
