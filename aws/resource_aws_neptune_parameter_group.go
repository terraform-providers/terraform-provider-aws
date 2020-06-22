package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

// We can only modify 20 parameters at a time, so walk them until
// we've got them all.
const maxParams = 20

func resourceAwsNeptuneParameterGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsNeptuneParameterGroupCreate,
		Read:   resourceAwsNeptuneParameterGroupRead,
		Update: resourceAwsNeptuneParameterGroupUpdate,
		Delete: resourceAwsNeptuneParameterGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				StateFunc: func(val interface{}) string {
					return strings.ToLower(val.(string))
				},
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

func resourceAwsNeptuneParameterGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).neptuneconn

	createOpts := neptune.CreateDBParameterGroupInput{
		DBParameterGroupName:   aws.String(d.Get("name").(string)),
		DBParameterGroupFamily: aws.String(d.Get("family").(string)),
		Description:            aws.String(d.Get("description").(string)),
		Tags:                   keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().NeptuneTags(),
	}

	log.Printf("[DEBUG] Create Neptune Parameter Group: %#v", createOpts)
	resp, err := conn.CreateDBParameterGroup(&createOpts)
	if err != nil {
		return fmt.Errorf("Error creating Neptune Parameter Group: %s", err)
	}

	d.SetId(aws.StringValue(resp.DBParameterGroup.DBParameterGroupName))
	d.Set("arn", resp.DBParameterGroup.DBParameterGroupArn)
	log.Printf("[INFO] Neptune Parameter Group ID: %s", d.Id())

	return resourceAwsNeptuneParameterGroupUpdate(d, meta)
}

func resourceAwsNeptuneParameterGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).neptuneconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	describeOpts := neptune.DescribeDBParameterGroupsInput{
		DBParameterGroupName: aws.String(d.Id()),
	}

	describeResp, err := conn.DescribeDBParameterGroups(&describeOpts)
	if err != nil {
		if isAWSErr(err, neptune.ErrCodeDBParameterGroupNotFoundFault, "") {
			log.Printf("[WARN] Neptune Parameter Group (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if describeResp == nil {
		return fmt.Errorf("Unable to get Describe Response for Neptune Parameter Group (%s)", d.Id())
	}

	if len(describeResp.DBParameterGroups) != 1 ||
		*describeResp.DBParameterGroups[0].DBParameterGroupName != d.Id() {
		return fmt.Errorf("Unable to find Parameter Group: %#v", describeResp.DBParameterGroups)
	}

	arn := aws.StringValue(describeResp.DBParameterGroups[0].DBParameterGroupArn)
	d.Set("arn", arn)
	d.Set("name", describeResp.DBParameterGroups[0].DBParameterGroupName)
	d.Set("family", describeResp.DBParameterGroups[0].DBParameterGroupFamily)
	d.Set("description", describeResp.DBParameterGroups[0].Description)

	// Only include user customized parameters as there's hundreds of system/default ones
	describeParametersOpts := neptune.DescribeDBParametersInput{
		DBParameterGroupName: aws.String(d.Id()),
		Source:               aws.String("user"),
	}

	var parameters []*neptune.Parameter
	err = conn.DescribeDBParametersPages(&describeParametersOpts,
		func(describeParametersResp *neptune.DescribeDBParametersOutput, lastPage bool) bool {
			parameters = append(parameters, describeParametersResp.Parameters...)
			return !lastPage
		})
	if err != nil {
		return err
	}

	if err := d.Set("parameter", flattenNeptuneParameters(parameters)); err != nil {
		return fmt.Errorf("error setting parameter: %s", err)
	}

	tags, err := keyvaluetags.NeptuneListTags(conn, d.Get("arn").(string))

	if err != nil {
		return fmt.Errorf("error listing tags for Neptune Parameter Group (%s): %s", d.Get("arn").(string), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsNeptuneParameterGroupUpdate(d *schema.ResourceData, meta interface{}) error {
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

		toRemove := expandNeptuneParameters(os.Difference(ns).List())

		log.Printf("[DEBUG] Parameters to remove: %#v", toRemove)

		toAdd := expandNeptuneParameters(ns.Difference(os).List())

		log.Printf("[DEBUG] Parameters to add: %#v", toAdd)

		for len(toRemove) > 0 {
			var paramsToModify []*neptune.Parameter
			if len(toRemove) <= maxParams {
				paramsToModify, toRemove = toRemove[:], nil
			} else {
				paramsToModify, toRemove = toRemove[:maxParams], toRemove[maxParams:]
			}
			resetOpts := neptune.ResetDBParameterGroupInput{
				DBParameterGroupName: aws.String(d.Get("name").(string)),
				Parameters:           paramsToModify,
			}

			log.Printf("[DEBUG] Reset Neptune Parameter Group: %s", resetOpts)
			err := resource.Retry(30*time.Second, func() *resource.RetryError {
				_, err := conn.ResetDBParameterGroup(&resetOpts)
				if err != nil {
					if isAWSErr(err, "InvalidDBParameterGroupState", " has pending changes") {
						return resource.RetryableError(err)
					}
					return resource.NonRetryableError(err)
				}
				return nil
			})
			if isResourceTimeoutError(err) {
				_, err = conn.ResetDBParameterGroup(&resetOpts)
			}
			if err != nil {
				return fmt.Errorf("Error resetting Neptune Parameter Group: %s", err)
			}
		}

		for len(toAdd) > 0 {
			var paramsToModify []*neptune.Parameter
			if len(toAdd) <= maxParams {
				paramsToModify, toAdd = toAdd[:], nil
			} else {
				paramsToModify, toAdd = toAdd[:maxParams], toAdd[maxParams:]
			}
			modifyOpts := neptune.ModifyDBParameterGroupInput{
				DBParameterGroupName: aws.String(d.Get("name").(string)),
				Parameters:           paramsToModify,
			}

			log.Printf("[DEBUG] Modify Neptune Parameter Group: %s", modifyOpts)
			_, err := conn.ModifyDBParameterGroup(&modifyOpts)
			if err != nil {
				return fmt.Errorf("Error modifying Neptune Parameter Group: %s", err)
			}
		}
	}

	if !d.IsNewResource() && d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.NeptuneUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Neptune Parameter Group (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return resourceAwsNeptuneParameterGroupRead(d, meta)
}

func resourceAwsNeptuneParameterGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).neptuneconn

	deleteOpts := neptune.DeleteDBParameterGroupInput{
		DBParameterGroupName: aws.String(d.Id()),
	}
	err := resource.Retry(3*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteDBParameterGroup(&deleteOpts)
		if err != nil {
			if isAWSErr(err, neptune.ErrCodeDBParameterGroupNotFoundFault, "") {
				return nil
			}
			if isAWSErr(err, neptune.ErrCodeInvalidDBParameterGroupStateFault, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteDBParameterGroup(&deleteOpts)
	}

	if isAWSErr(err, neptune.ErrCodeDBParameterGroupNotFoundFault, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error deleting Neptune Parameter Group: %s", err)
	}

	return nil
}
