package aws

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsDbParameterGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDbParameterGroupCreate,
		Read:   resourceAwsDbParameterGroupRead,
		Update: resourceAwsDbParameterGroupUpdate,
		Delete: resourceAwsDbParameterGroupDelete,
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

func resourceAwsDbParameterGroupCreate(d *schema.ResourceData, meta interface{}) error {
	rdsconn := meta.(*AWSClient).rdsconn
	tags := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().RdsTags()

	var groupName string
	if v, ok := d.GetOk("name"); ok {
		groupName = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		groupName = resource.PrefixedUniqueId(v.(string))
	} else {
		groupName = resource.UniqueId()
	}
	d.Set("name", groupName)

	createOpts := rds.CreateDBParameterGroupInput{
		DBParameterGroupName:   aws.String(groupName),
		DBParameterGroupFamily: aws.String(d.Get("family").(string)),
		Description:            aws.String(d.Get("description").(string)),
		Tags:                   tags,
	}

	log.Printf("[DEBUG] Create DB Parameter Group: %#v", createOpts)
	resp, err := rdsconn.CreateDBParameterGroup(&createOpts)
	if err != nil {
		return fmt.Errorf("Error creating DB Parameter Group: %s", err)
	}

	d.SetId(aws.StringValue(resp.DBParameterGroup.DBParameterGroupName))
	d.Set("arn", resp.DBParameterGroup.DBParameterGroupArn)
	log.Printf("[INFO] DB Parameter Group ID: %s", d.Id())

	return resourceAwsDbParameterGroupUpdate(d, meta)
}

func resourceAwsDbParameterGroupRead(d *schema.ResourceData, meta interface{}) error {
	rdsconn := meta.(*AWSClient).rdsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	describeOpts := rds.DescribeDBParameterGroupsInput{
		DBParameterGroupName: aws.String(d.Id()),
	}

	describeResp, err := rdsconn.DescribeDBParameterGroups(&describeOpts)
	if err != nil {
		if isAWSErr(err, rds.ErrCodeDBParameterGroupNotFoundFault, "") {
			log.Printf("[WARN] DB Parameter Group (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if len(describeResp.DBParameterGroups) != 1 ||
		aws.StringValue(describeResp.DBParameterGroups[0].DBParameterGroupName) != d.Id() {
		return fmt.Errorf("Unable to find Parameter Group: %#v", describeResp.DBParameterGroups)
	}

	d.Set("name", describeResp.DBParameterGroups[0].DBParameterGroupName)
	d.Set("family", describeResp.DBParameterGroups[0].DBParameterGroupFamily)
	d.Set("description", describeResp.DBParameterGroups[0].Description)

	configParams := d.Get("parameter").(*schema.Set)
	describeParametersOpts := rds.DescribeDBParametersInput{
		DBParameterGroupName: aws.String(d.Id()),
	}
	if configParams.Len() < 1 {
		// if we don't have any params in the ResourceData already, two possibilities
		// first, we don't have a config available to us. Second, we do, but it has
		// no parameters. We're going to assume the first, to be safe. In this case,
		// we're only going to ask for the user-modified values, because any defaults
		// the user may have _also_ set are indistinguishable from the hundreds of
		// defaults AWS sets. If the user hasn't set any parameters, this will return
		// an empty list anyways, so we just make some unnecessary requests. But in
		// the more common case (I assume) of an import, this will make fewer requests
		// and "do the right thing".
		describeParametersOpts.Source = aws.String("user")
	}

	var parameters []*rds.Parameter
	err = rdsconn.DescribeDBParametersPages(&describeParametersOpts,
		func(describeParametersResp *rds.DescribeDBParametersOutput, lastPage bool) bool {
			parameters = append(parameters, describeParametersResp.Parameters...)
			return !lastPage
		})
	if err != nil {
		return err
	}

	var userParams []*rds.Parameter
	if configParams.Len() < 1 {
		// if we have no config/no parameters in config, we've already asked for only
		// user-modified values, so we can just use the entire response.
		userParams = parameters
	} else {
		// if we have a config available to us, we have two possible classes of value
		// in the config. On the one hand, the user could have specified a parameter
		// that _actually_ changed things, in which case its Source would be set to
		// user. On the other, they may have specified a parameter that coincides with
		// the default value. In that case, the Source will be set to "system" or
		// "engine-default". We need to set the union of all "user" Source parameters
		// _and_ the "system"/"engine-default" Source parameters _that appear in the
		// config_ in the state, or the user gets a perpetual diff. See
		// terraform-providers/terraform-provider-aws#593 for more context and details.
		confParams := expandParameters(configParams.List())
		for _, param := range parameters {
			if param.Source == nil || param.ParameterName == nil {
				continue
			}
			if aws.StringValue(param.Source) == "user" {
				userParams = append(userParams, param)
				continue
			}
			var paramFound bool
			for _, cp := range confParams {
				if cp.ParameterName == nil {
					continue
				}
				if aws.StringValue(cp.ParameterName) == aws.StringValue(param.ParameterName) {
					userParams = append(userParams, param)
					break
				}
			}
			if !paramFound {
				log.Printf("[DEBUG] Not persisting %s to state, as its source is %q and it isn't in the config", aws.StringValue(param.ParameterName), aws.StringValue(param.Source))
			}
		}
	}

	err = d.Set("parameter", flattenParameters(userParams))
	if err != nil {
		return fmt.Errorf("error setting 'parameter' in state: %#v", err)
	}

	arn := aws.StringValue(describeResp.DBParameterGroups[0].DBParameterGroupArn)
	d.Set("arn", arn)

	tags, err := keyvaluetags.RdsListTags(rdsconn, d.Get("arn").(string))

	if err != nil {
		return fmt.Errorf("error listing tags for RDS DB Parameter Group (%s): %s", d.Get("arn").(string), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsDbParameterGroupUpdate(d *schema.ResourceData, meta interface{}) error {
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
			maxParams := 20
			for parameters != nil {
				var paramsToModify []*rds.Parameter
				if len(parameters) <= maxParams {
					paramsToModify, parameters = parameters[:], nil
				} else {
					paramsToModify, parameters = parameters[:maxParams], parameters[maxParams:]
				}
				modifyOpts := rds.ModifyDBParameterGroupInput{
					DBParameterGroupName: aws.String(d.Get("name").(string)),
					Parameters:           paramsToModify,
				}

				log.Printf("[DEBUG] Modify DB Parameter Group: %s", modifyOpts)
				_, err := rdsconn.ModifyDBParameterGroup(&modifyOpts)
				if err != nil {
					return fmt.Errorf("Error modifying DB Parameter Group: %s", err)
				}
			}
		}

		// Reset parameters that have been removed
		resetParameters := expandParameters(os.Difference(ns).List())
		if len(resetParameters) > 0 {
			maxParams := 20
			for resetParameters != nil {
				var paramsToReset []*rds.Parameter
				if len(resetParameters) <= maxParams {
					paramsToReset, resetParameters = resetParameters[:], nil
				} else {
					paramsToReset, resetParameters = resetParameters[:maxParams], resetParameters[maxParams:]
				}

				parameterGroupName := d.Get("name").(string)
				resetOpts := rds.ResetDBParameterGroupInput{
					DBParameterGroupName: aws.String(parameterGroupName),
					Parameters:           paramsToReset,
					ResetAllParameters:   aws.Bool(false),
				}

				log.Printf("[DEBUG] Reset DB Parameter Group: %s", resetOpts)
				_, err := rdsconn.ResetDBParameterGroup(&resetOpts)
				if err != nil {
					return fmt.Errorf("Error resetting DB Parameter Group: %s", err)
				}
			}
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.RdsUpdateTags(rdsconn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating RDS DB Parameter Group (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return resourceAwsDbParameterGroupRead(d, meta)
}

func resourceAwsDbParameterGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn
	deleteOpts := rds.DeleteDBParameterGroupInput{
		DBParameterGroupName: aws.String(d.Id()),
	}
	err := resource.Retry(3*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteDBParameterGroup(&deleteOpts)
		if err != nil {
			if isAWSErr(err, "DBParameterGroupNotFoundFault", "") || isAWSErr(err, "InvalidDBParameterGroupState", "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DeleteDBParameterGroup(&deleteOpts)
	}
	if err != nil {
		return fmt.Errorf("Error deleting DB parameter group: %s", err)
	}
	return nil
}

func resourceAwsDbParameterHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	// Store the value as a lower case string, to match how we store them in flattenParameters
	buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(m["value"].(string))))

	return hashcode.String(buf.String())
}
