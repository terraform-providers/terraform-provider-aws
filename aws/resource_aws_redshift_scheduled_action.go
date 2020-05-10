package aws

import (
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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
				ForceNew: true,
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
