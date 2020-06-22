package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

const rdsClusterParameterGroupMaxParamsBulkEdit = 20

func resourceAwsRDSClusterParameterGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRDSClusterParameterGroupCreate,
		Read:   resourceAwsRDSClusterParameterGroupRead,
		Update: resourceAwsRDSClusterParameterGroupUpdate,
		Delete: resourceAwsRDSClusterParameterGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validateDbParamGroupName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validateDbParamGroupNamePrefix,
			},
			"family": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "Managed by Terraform",
			},
			"parameter": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"apply_method": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "immediate",
						},
					},
				},
				Set: resourceAwsDbParameterHash,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsRDSClusterParameterGroupCreate(d *schema.ResourceData, meta interface{}) error {
	rdsconn := meta.(*AWSClient).rdsconn

	var groupName string
	if v, ok := d.GetOk("name"); ok {
		groupName = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		groupName = resource.PrefixedUniqueId(v.(string))
	} else {
		groupName = resource.UniqueId()
	}

	createOpts := rds.CreateDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(groupName),
		DBParameterGroupFamily:      aws.String(d.Get("family").(string)),
		Description:                 aws.String(d.Get("description").(string)),
		Tags:                        keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().RdsTags(),
	}

	log.Printf("[DEBUG] Create DB Cluster Parameter Group: %#v", createOpts)
	output, err := rdsconn.CreateDBClusterParameterGroup(&createOpts)
	if err != nil {
		return fmt.Errorf("Error creating DB Cluster Parameter Group: %s", err)
	}

	d.SetId(*createOpts.DBClusterParameterGroupName)
	log.Printf("[INFO] DB Cluster Parameter Group ID: %s", d.Id())

	// Set for update
	d.Set("arn", output.DBClusterParameterGroup.DBClusterParameterGroupArn)

	return resourceAwsRDSClusterParameterGroupUpdate(d, meta)
}

func resourceAwsRDSClusterParameterGroupRead(d *schema.ResourceData, meta interface{}) error {
	rdsconn := meta.(*AWSClient).rdsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	describeOpts := rds.DescribeDBClusterParameterGroupsInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
	}

	describeResp, err := rdsconn.DescribeDBClusterParameterGroups(&describeOpts)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "DBParameterGroupNotFound" {
			log.Printf("[WARN] DB Cluster Parameter Group (%s) not found, error code (404)", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	if len(describeResp.DBClusterParameterGroups) != 1 ||
		aws.StringValue(describeResp.DBClusterParameterGroups[0].DBClusterParameterGroupName) != d.Id() {
		return fmt.Errorf("Unable to find Cluster Parameter Group: %#v", describeResp.DBClusterParameterGroups)
	}

	arn := aws.StringValue(describeResp.DBClusterParameterGroups[0].DBClusterParameterGroupArn)
	d.Set("arn", arn)
	d.Set("description", describeResp.DBClusterParameterGroups[0].Description)
	d.Set("family", describeResp.DBClusterParameterGroups[0].DBParameterGroupFamily)
	d.Set("name", describeResp.DBClusterParameterGroups[0].DBClusterParameterGroupName)

	// Only include user customized parameters as there's hundreds of system/default ones
	describeParametersOpts := rds.DescribeDBClusterParametersInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
		Source:                      aws.String("user"),
	}

	describeParametersResp, err := rdsconn.DescribeDBClusterParameters(&describeParametersOpts)
	if err != nil {
		return err
	}

	if err := d.Set("parameter", flattenParameters(describeParametersResp.Parameters)); err != nil {
		return fmt.Errorf("error setting parameters: %s", err)
	}

	resp, err := rdsconn.ListTagsForResource(&rds.ListTagsForResourceInput{
		ResourceName: aws.String(arn),
	})
	if err != nil {
		log.Printf("[DEBUG] Error retrieving tags for ARN: %s", arn)
	}

	if err := d.Set("tags", keyvaluetags.RdsKeyValueTags(resp.TagList).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsRDSClusterParameterGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	rdsconn := meta.(*AWSClient).rdsconn

	if d.HasChange("parameter") {
		o, n := d.GetChange("parameter")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		// Expand the "parameter" set to aws-sdk-go compat []rds.Parameter
		parameters := expandParameters(ns.Difference(os).List())
		if len(parameters) > 0 {
			// We can only modify 20 parameters at a time, so walk them until
			// we've got them all.
			for parameters != nil {
				var paramsToModify []*rds.Parameter
				if len(parameters) <= rdsClusterParameterGroupMaxParamsBulkEdit {
					paramsToModify, parameters = parameters[:], nil
				} else {
					paramsToModify, parameters = parameters[:rdsClusterParameterGroupMaxParamsBulkEdit], parameters[rdsClusterParameterGroupMaxParamsBulkEdit:]
				}

				modifyOpts := rds.ModifyDBClusterParameterGroupInput{
					DBClusterParameterGroupName: aws.String(d.Id()),
					Parameters:                  paramsToModify,
				}

				log.Printf("[DEBUG] Modify DB Cluster Parameter Group: %s", modifyOpts)
				_, err := rdsconn.ModifyDBClusterParameterGroup(&modifyOpts)
				if err != nil {
					return fmt.Errorf("error modifying DB Cluster Parameter Group: %s", err)
				}
			}
		}

		// Reset parameters that have been removed
		parameters = expandParameters(os.Difference(ns).List())
		if len(parameters) > 0 {
			for parameters != nil {
				parameterGroupName := d.Get("name").(string)
				var paramsToReset []*rds.Parameter
				if len(parameters) <= rdsClusterParameterGroupMaxParamsBulkEdit {
					paramsToReset, parameters = parameters[:], nil
				} else {
					paramsToReset, parameters = parameters[:rdsClusterParameterGroupMaxParamsBulkEdit], parameters[rdsClusterParameterGroupMaxParamsBulkEdit:]
				}

				resetOpts := rds.ResetDBClusterParameterGroupInput{
					DBClusterParameterGroupName: aws.String(parameterGroupName),
					Parameters:                  paramsToReset,
					ResetAllParameters:          aws.Bool(false),
				}

				log.Printf("[DEBUG] Reset DB Cluster Parameter Group: %s", resetOpts)
				err := resource.Retry(3*time.Minute, func() *resource.RetryError {
					_, err := rdsconn.ResetDBClusterParameterGroup(&resetOpts)
					if err != nil {
						if isAWSErr(err, "InvalidDBParameterGroupState", "has pending changes") {
							return resource.RetryableError(err)
						}
						return resource.NonRetryableError(err)
					}
					return nil
				})
				if err != nil {
					return fmt.Errorf("error resetting DB Cluster Parameter Group: %s", err)
				}
			}
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.RdsUpdateTags(rdsconn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating RDS Cluster Parameter Group (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsRDSClusterParameterGroupRead(d, meta)
}

func resourceAwsRDSClusterParameterGroupDelete(d *schema.ResourceData, meta interface{}) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"pending"},
		Target:     []string{"destroyed"},
		Refresh:    resourceAwsRDSClusterParameterGroupDeleteRefreshFunc(d, meta),
		Timeout:    3 * time.Minute,
		MinTimeout: 1 * time.Second,
	}
	_, err := stateConf.WaitForState()
	return err
}

func resourceAwsRDSClusterParameterGroupDeleteRefreshFunc(
	d *schema.ResourceData,
	meta interface{}) resource.StateRefreshFunc {
	rdsconn := meta.(*AWSClient).rdsconn

	return func() (interface{}, string, error) {

		deleteOpts := rds.DeleteDBClusterParameterGroupInput{
			DBClusterParameterGroupName: aws.String(d.Id()),
		}

		if _, err := rdsconn.DeleteDBClusterParameterGroup(&deleteOpts); err != nil {
			rdserr, ok := err.(awserr.Error)
			if !ok {
				return d, "error", err
			}

			if rdserr.Code() != "DBParameterGroupNotFound" {
				return d, "error", err
			}
		}

		return d, "destroyed", nil
	}
}
