package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

const docdbClusterParameterGroupMaxParamsBulkEdit = 20

func resourceAwsDocDBClusterParameterGroup() *schema.Resource {

	return &schema.Resource{
		Create: resourceAwsDocDBClusterParameterGroupCreate,
		Read:   resourceAwsDocDBClusterParameterGroupRead,
		Update: resourceAwsDocDBClusterParameterGroupUpdate,
		Delete: resourceAwsDocDBClusterParameterGroupDelete,
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
				ValidateFunc:  validateDocDBParamGroupName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validateDocDBParamGroupNamePrefix,
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
							Default:  docdb.ApplyMethodPendingReboot,
							ValidateFunc: validation.StringInSlice([]string{
								docdb.ApplyMethodImmediate,
								docdb.ApplyMethodPendingReboot,
							}, false),
						},
					},
				},
			},

			"tags": tagsSchema(),
		},
	}

}

func resourceAwsDocDBClusterParameterGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).docdbconn
	tags := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().DocdbTags()

	var groupName string
	if v, ok := d.GetOk("name"); ok {
		groupName = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		groupName = resource.PrefixedUniqueId(v.(string))
	} else {
		groupName = resource.UniqueId()
	}

	createOpts := docdb.CreateDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(groupName),
		DBParameterGroupFamily:      aws.String(d.Get("family").(string)),
		Description:                 aws.String(d.Get("description").(string)),
		Tags:                        tags,
	}

	log.Printf("[DEBUG] Create DocDB Cluster Parameter Group: %#v", createOpts)

	resp, err := conn.CreateDBClusterParameterGroup(&createOpts)
	if err != nil {
		return fmt.Errorf("Error creating DocDB Cluster Parameter Group: %s", err)
	}

	d.SetId(aws.StringValue(createOpts.DBClusterParameterGroupName))

	d.Set("arn", resp.DBClusterParameterGroup.DBClusterParameterGroupArn)

	return resourceAwsDocDBClusterParameterGroupUpdate(d, meta)
}

func resourceAwsDocDBClusterParameterGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).docdbconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	describeOpts := &docdb.DescribeDBClusterParameterGroupsInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
	}

	describeResp, err := conn.DescribeDBClusterParameterGroups(describeOpts)
	if err != nil {
		if isAWSErr(err, docdb.ErrCodeDBParameterGroupNotFoundFault, "") {
			log.Printf("[WARN] DocDB Cluster Parameter Group (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading DocDB Cluster Parameter Group (%s): %s", d.Id(), err)
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

	describeParametersOpts := &docdb.DescribeDBClusterParametersInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
		Source:                      aws.String("user"),
	}

	describeParametersResp, err := conn.DescribeDBClusterParameters(describeParametersOpts)
	if err != nil {
		return fmt.Errorf("error reading DocDB Cluster Parameter Group (%s) parameters: %s", d.Id(), err)
	}

	if err := d.Set("parameter", flattenDocDBParameters(describeParametersResp.Parameters)); err != nil {
		return fmt.Errorf("error setting docdb cluster parameter: %s", err)
	}

	tags, err := keyvaluetags.DocdbListTags(conn, d.Get("arn").(string))

	if err != nil {
		return fmt.Errorf("error listing tags for DocumentDB Cluster Parameter Group (%s): %s", d.Get("arn").(string), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsDocDBClusterParameterGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).docdbconn

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

		parameters := expandDocDBParameters(ns.Difference(os).List())
		if len(parameters) > 0 {
			// We can only modify 20 parameters at a time, so walk them until
			// we've got them all.
			for parameters != nil {
				var paramsToModify []*docdb.Parameter
				if len(parameters) <= docdbClusterParameterGroupMaxParamsBulkEdit {
					paramsToModify, parameters = parameters[:], nil
				} else {
					paramsToModify, parameters = parameters[:docdbClusterParameterGroupMaxParamsBulkEdit], parameters[docdbClusterParameterGroupMaxParamsBulkEdit:]
				}
				parameterGroupName := d.Id()
				modifyOpts := docdb.ModifyDBClusterParameterGroupInput{
					DBClusterParameterGroupName: aws.String(parameterGroupName),
					Parameters:                  paramsToModify,
				}

				log.Printf("[DEBUG] Modify DocDB Cluster Parameter Group: %#v", modifyOpts)
				_, err := conn.ModifyDBClusterParameterGroup(&modifyOpts)
				if err != nil {
					if isAWSErr(err, docdb.ErrCodeDBParameterGroupNotFoundFault, "") {
						log.Printf("[WARN] DocDB Cluster Parameter Group (%s) not found, removing from state", d.Id())
						d.SetId("")
						return nil
					}
					return fmt.Errorf("Error modifying DocDB Cluster Parameter Group: %s", err)
				}
			}
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.DocdbUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating DocumentDB Cluster Parameter Group (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return resourceAwsDocDBClusterParameterGroupRead(d, meta)
}

func resourceAwsDocDBClusterParameterGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).docdbconn

	deleteOpts := &docdb.DeleteDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
	}

	_, err := conn.DeleteDBClusterParameterGroup(deleteOpts)
	if err != nil {
		if isAWSErr(err, docdb.ErrCodeDBParameterGroupNotFoundFault, "") {
			return nil
		}
		return err
	}

	return waitForDocDBClusterParameterGroupDeletion(conn, d.Id())
}

func waitForDocDBClusterParameterGroupDeletion(conn *docdb.DocDB, name string) error {
	params := &docdb.DescribeDBClusterParameterGroupsInput{
		DBClusterParameterGroupName: aws.String(name),
	}

	err := resource.Retry(10*time.Minute, func() *resource.RetryError {
		_, err := conn.DescribeDBClusterParameterGroups(params)

		if isAWSErr(err, docdb.ErrCodeDBParameterGroupNotFoundFault, "") {
			return nil
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return resource.RetryableError(fmt.Errorf("DocDB Parameter Group (%s) still exists", name))
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DescribeDBClusterParameterGroups(params)
		if isAWSErr(err, docdb.ErrCodeDBParameterGroupNotFoundFault, "") {
			return nil
		}
	}
	if err != nil {
		return fmt.Errorf("Error deleting DocDB cluster parameter group: %s", err)
	}
	return nil
}
