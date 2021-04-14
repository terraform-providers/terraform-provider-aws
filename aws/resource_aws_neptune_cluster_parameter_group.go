package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

const neptuneClusterParameterGroupMaxParamsBulkEdit = 20

func resourceAwsNeptuneClusterParameterGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsNeptuneClusterParameterGroupCreate,
		Read:   resourceAwsNeptuneClusterParameterGroupRead,
		Update: resourceAwsNeptuneClusterParameterGroupUpdate,
		Delete: resourceAwsNeptuneClusterParameterGroupDelete,
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
				ValidateFunc:  validateNeptuneParamGroupName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validateNeptuneParamGroupNamePrefix,
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
							Default:  neptune.ApplyMethodPendingReboot,
							ValidateFunc: validation.StringInSlice([]string{
								neptune.ApplyMethodImmediate,
								neptune.ApplyMethodPendingReboot,
							}, false),
						},
					},
				},
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsNeptuneClusterParameterGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).neptuneconn
	tags := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().NeptuneTags()

	var groupName string
	if v, ok := d.GetOk("name"); ok {
		groupName = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		groupName = resource.PrefixedUniqueId(v.(string))
	} else {
		groupName = resource.UniqueId()
	}

	createOpts := neptune.CreateDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(groupName),
		DBParameterGroupFamily:      aws.String(d.Get("family").(string)),
		Description:                 aws.String(d.Get("description").(string)),
		Tags:                        tags,
	}

	_, err := conn.CreateDBClusterParameterGroup(&createOpts)
	if err != nil {
		return fmt.Errorf("error creating Neptune Cluster Parameter Group (%s): %w", groupName, err)
	}

	d.SetId(aws.StringValue(createOpts.DBClusterParameterGroupName))

	if v, ok := d.GetOk("parameter"); ok && v.(*schema.Set).Len() > 0 {
		err := modifyNeptuneClusterParameterGroupParameters(conn, d.Id(), expandNeptuneParameters(v.(*schema.Set).List()))
		if err != nil {
			return fmt.Errorf("error modifying Neptune Cluster Parameter Group (%s): %w", d.Id(), err)
		}
	}

	return resourceAwsNeptuneClusterParameterGroupRead(d, meta)
}

func resourceAwsNeptuneClusterParameterGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).neptuneconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	describeOpts := neptune.DescribeDBClusterParameterGroupsInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
	}

	describeResp, err := conn.DescribeDBClusterParameterGroups(&describeOpts)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBParameterGroupNotFoundFault) {
		log.Printf("[WARN] Neptune Cluster Parameter Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Neptune Cluster Parameter Group (%s): %w", d.Id(), err)
	}

	if describeResp == nil || len(describeResp.DBClusterParameterGroups) == 0 {
		if d.IsNewResource() {
			return fmt.Errorf("error reading Neptune Cluster Parameter Group (%s): not found", d.Id())
		}

		log.Printf("[WARN] Neptune Cluster Parameter Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("name", describeResp.DBClusterParameterGroups[0].DBClusterParameterGroupName)
	d.Set("family", describeResp.DBClusterParameterGroups[0].DBParameterGroupFamily)
	d.Set("description", describeResp.DBClusterParameterGroups[0].Description)
	arn := aws.StringValue(describeResp.DBClusterParameterGroups[0].DBClusterParameterGroupArn)
	d.Set("arn", arn)

	// Only include user customized parameters as there's hundreds of system/default ones
	describeParametersOpts := neptune.DescribeDBClusterParametersInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
		Source:                      aws.String("user"),
	}

	describeParametersResp, err := conn.DescribeDBClusterParameters(&describeParametersOpts)
	if err != nil {
		return err
	}

	if err := d.Set("parameter", flattenNeptuneParameters(describeParametersResp.Parameters)); err != nil {
		return fmt.Errorf("error setting neptune parameter: %w", err)
	}

	tags, err := keyvaluetags.NeptuneListTags(conn, d.Get("arn").(string))

	if err != nil {
		return fmt.Errorf("error listing tags for Neptune Cluster Parameter Group (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsNeptuneClusterParameterGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).neptuneconn

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

		parameters := expandNeptuneParameters(ns.Difference(os).List())

		if len(parameters) > 0 {
			err := modifyNeptuneClusterParameterGroupParameters(conn, d.Id(), parameters)
			if err != nil {
				return fmt.Errorf("error updating Neptune Cluster Parameter Group (%s) parameter: %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.NeptuneUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Neptune Cluster Parameter Group (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsNeptuneClusterParameterGroupRead(d, meta)
}

func resourceAwsNeptuneClusterParameterGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).neptuneconn

	input := neptune.DeleteDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
	}

	_, err := conn.DeleteDBClusterParameterGroup(&input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBParameterGroupNotFoundFault) {
			return nil
		}
		return fmt.Errorf("error deleting Neptune Cluster Parameter Group (%s): %w", d.Id(), err)
	}

	return nil
}

func modifyNeptuneClusterParameterGroupParameters(conn *neptune.Neptune, name string, parameters []*neptune.Parameter) error {
	// We can only modify 20 parameters at a time, so walk them until
	// we've got them all.
	for parameters != nil {
		var paramsToModify []*neptune.Parameter
		if len(parameters) <= neptuneClusterParameterGroupMaxParamsBulkEdit {
			paramsToModify, parameters = parameters[:], nil
		} else {
			paramsToModify, parameters = parameters[:neptuneClusterParameterGroupMaxParamsBulkEdit], parameters[neptuneClusterParameterGroupMaxParamsBulkEdit:]
		}

		modifyOpts := neptune.ModifyDBClusterParameterGroupInput{
			DBClusterParameterGroupName: aws.String(name),
			Parameters:                  paramsToModify,
		}

		_, err := conn.ModifyDBClusterParameterGroup(&modifyOpts)
		if err != nil {
			return err
		}
	}

	return nil
}
