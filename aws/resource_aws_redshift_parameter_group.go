package aws

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsRedshiftParameterGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRedshiftParameterGroupCreate,
		Read:   resourceAwsRedshiftParameterGroupRead,
		Update: resourceAwsRedshiftParameterGroupUpdate,
		Delete: resourceAwsRedshiftParameterGroupDelete,
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
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-z-]+$`), "must contain only lowercase alphanumeric characters and hyphens"),
					validation.StringMatch(regexp.MustCompile(`(?i)^[a-z]`), "first character must be a letter"),
					validation.StringDoesNotMatch(regexp.MustCompile(`--`), "cannot contain two consecutive hyphens"),
					validation.StringDoesNotMatch(regexp.MustCompile(`-$`), "cannot end with a hyphen"),
				),
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
					},
				},
				Set: resourceAwsRedshiftParameterHash,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsRedshiftParameterGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).redshiftconn

	createOpts := redshift.CreateClusterParameterGroupInput{
		ParameterGroupName:   aws.String(d.Get("name").(string)),
		ParameterGroupFamily: aws.String(d.Get("family").(string)),
		Description:          aws.String(d.Get("description").(string)),
		Tags:                 keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().RedshiftTags(),
	}

	log.Printf("[DEBUG] Create Redshift Parameter Group: %#v", createOpts)
	_, err := conn.CreateClusterParameterGroup(&createOpts)
	if err != nil {
		return fmt.Errorf("Error creating Redshift Parameter Group: %s", err)
	}

	d.SetId(*createOpts.ParameterGroupName)

	if v := d.Get("parameter").(*schema.Set); v.Len() > 0 {
		parameters := expandRedshiftParameters(v.List())

		modifyOpts := redshift.ModifyClusterParameterGroupInput{
			ParameterGroupName: aws.String(d.Id()),
			Parameters:         parameters,
		}

		if _, err := conn.ModifyClusterParameterGroup(&modifyOpts); err != nil {
			return fmt.Errorf("error adding Redshift Parameter Group (%s) parameters: %s", d.Id(), err)
		}
	}

	return resourceAwsRedshiftParameterGroupRead(d, meta)
}

func resourceAwsRedshiftParameterGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).redshiftconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	describeOpts := redshift.DescribeClusterParameterGroupsInput{
		ParameterGroupName: aws.String(d.Id()),
	}

	describeResp, err := conn.DescribeClusterParameterGroups(&describeOpts)
	if err != nil {
		return err
	}

	if len(describeResp.ParameterGroups) != 1 ||
		aws.StringValue(describeResp.ParameterGroups[0].ParameterGroupName) != d.Id() {
		d.SetId("")
		return fmt.Errorf("Unable to find Parameter Group: %#v", describeResp.ParameterGroups)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "redshift",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("parametergroup:%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	d.Set("name", describeResp.ParameterGroups[0].ParameterGroupName)
	d.Set("family", describeResp.ParameterGroups[0].ParameterGroupFamily)
	d.Set("description", describeResp.ParameterGroups[0].Description)
	if err := d.Set("tags", keyvaluetags.RedshiftKeyValueTags(describeResp.ParameterGroups[0].Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	describeParametersOpts := redshift.DescribeClusterParametersInput{
		ParameterGroupName: aws.String(d.Id()),
		Source:             aws.String("user"),
	}

	describeParametersResp, err := conn.DescribeClusterParameters(&describeParametersOpts)
	if err != nil {
		return err
	}

	d.Set("parameter", flattenRedshiftParameters(describeParametersResp.Parameters))
	return nil
}

func resourceAwsRedshiftParameterGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).redshiftconn

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

		// Expand the "parameter" set to aws-sdk-go compat []redshift.Parameter
		parameters := expandRedshiftParameters(ns.Difference(os).List())

		if len(parameters) > 0 {
			modifyOpts := redshift.ModifyClusterParameterGroupInput{
				ParameterGroupName: aws.String(d.Get("name").(string)),
				Parameters:         parameters,
			}

			log.Printf("[DEBUG] Modify Redshift Parameter Group: %s", modifyOpts)
			_, err := conn.ModifyClusterParameterGroup(&modifyOpts)
			if err != nil {
				return fmt.Errorf("Error modifying Redshift Parameter Group: %s", err)
			}
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.RedshiftUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Redshift Parameter Group (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return resourceAwsRedshiftParameterGroupRead(d, meta)
}

func resourceAwsRedshiftParameterGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).redshiftconn

	_, err := conn.DeleteClusterParameterGroup(&redshift.DeleteClusterParameterGroupInput{
		ParameterGroupName: aws.String(d.Id()),
	})
	if err != nil && isAWSErr(err, "RedshiftParameterGroupNotFoundFault", "") {
		return nil
	}
	return err
}

func resourceAwsRedshiftParameterHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	// Store the value as a lower case string, to match how we store them in flattenParameters
	buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(m["value"].(string))))

	return hashcode.String(buf.String())
}
