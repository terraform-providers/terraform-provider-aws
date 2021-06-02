package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudfront/finder"
)

func dataSourceAwsCloudFrontFunction() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCloudFrontFunctionRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"code": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"comment": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"runtime": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"stage": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(cloudfront.FunctionStage_Values(), false),
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsCloudFrontFunctionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	name := d.Get("name").(string)
	stage := d.Get("stage").(string)

	describeFunctionOutput, err := finder.FunctionByNameAndStage(conn, name, stage)

	if err != nil {
		return fmt.Errorf("error describing CloudFront Function (%s/%s): %w", name, stage, err)
	}

	d.Set("arn", describeFunctionOutput.FunctionSummary.FunctionMetadata.FunctionARN)
	d.Set("comment", describeFunctionOutput.FunctionSummary.FunctionConfig.Comment)
	d.Set("etag", describeFunctionOutput.ETag)
	d.Set("last_modified_time", describeFunctionOutput.FunctionSummary.FunctionMetadata.LastModifiedTime.Format(time.RFC3339))
	d.Set("name", describeFunctionOutput.FunctionSummary.Name)
	d.Set("runtime", describeFunctionOutput.FunctionSummary.FunctionConfig.Runtime)
	d.Set("status", describeFunctionOutput.FunctionSummary.Status)

	getFunctionOutput, err := conn.GetFunction(&cloudfront.GetFunctionInput{
		Name:  aws.String(name),
		Stage: aws.String(stage),
	})

	if err != nil {
		return fmt.Errorf("error getting CloudFront Function (%s): %w", d.Id(), err)
	}

	d.Set("code", string(getFunctionOutput.FunctionCode))

	d.SetId(aws.StringValue(describeFunctionOutput.FunctionSummary.Name))

	return nil
}
