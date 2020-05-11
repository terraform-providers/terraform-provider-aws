package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"time"
)

func resourceAwsRedshiftScheduledAction() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRedshiftScheduledActionCreate,
		Read:   resourceAwsRedshiftScheduledActionRead,
		Update: resourceAwsRedshiftScheduledActionUpdate,
		Delete: resourceAwsRedshiftScheduledActionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"active": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"start_time": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"end_time": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"schedule": {
				Type:     schema.TypeString,
				Required: true,
			},
			"iam_role": {
				Type:     schema.TypeString,
				Required: true,
			},
			"target_action": {
				Type:     schema.TypeMap,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"PauseClusterMessage",
								"ResizeCluster",
								"ResumeCluster",
							}, false),
						},
						"cluster_identifier": {
							Type:     schema.TypeString,
							Required: true,
						},
						"classic": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"cluster_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"node_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"number_of_nodes": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsRedshiftScheduledActionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).redshiftconn
	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		name = resource.UniqueId()
	}
	createOpts := &redshift.CreateScheduledActionInput{
		ScheduledActionName: aws.String(name),
		Schedule:            aws.String(d.Get("schedule").(string)),
		IamRole:             aws.String(d.Get("iam_role").(string)),
	}
	// TODO: add target_action
	if attr, ok := d.GetOk("description"); ok {
		createOpts.ScheduledActionDescription = aws.String(attr.(string))
	}
	if attr, ok := d.GetOk("active"); ok {
		createOpts.Enable = aws.Bool(attr.(bool))
	}
	if attr, ok := d.GetOk("start_time"); ok {
		createOpts.StartTime = aws.Time(attr.(time.Time))
	}
	if attr, ok := d.GetOk("end_time"); ok {
		createOpts.EndTime = aws.Time(attr.(time.Time))
	}
}
