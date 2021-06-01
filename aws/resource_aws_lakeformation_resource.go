package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsLakeFormationResource() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLakeFormationResourceCreate,
		Read:   resourceAwsLakeFormationResourceRead,
		Delete: resourceAwsLakeFormationResourceDelete,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
		},
	}
}

func resourceAwsLakeFormationResourceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn
	resourceArn := d.Get("arn").(string)

	input := &lakeformation.RegisterResourceInput{
		ResourceArn: aws.String(resourceArn),
	}

	if v, ok := d.GetOk("role_arn"); ok {
		input.RoleArn = aws.String(v.(string))
	} else {
		input.UseServiceLinkedRole = aws.Bool(true)
	}

	_, err := conn.RegisterResource(input)

	if tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeAlreadyExistsException) {
		log.Printf("[WARN] Lake Formation Resource (%s) already exists", resourceArn)
	} else if err != nil {
		return fmt.Errorf("error registering Lake Formation Resource (%s): %s", resourceArn, err)
	}

	d.SetId(resourceArn)
	return resourceAwsLakeFormationResourceRead(d, meta)
}

func resourceAwsLakeFormationResourceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn
	resourceArn := d.Get("arn").(string)

	input := &lakeformation.DescribeResourceInput{
		ResourceArn: aws.String(resourceArn),
	}

	output, err := conn.DescribeResource(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeEntityNotFoundException) {
		log.Printf("[WARN] Resource Lake Formation Resource (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading resource Lake Formation Resource (%s): %w", d.Id(), err)
	}

	if output == nil || output.ResourceInfo == nil {
		return fmt.Errorf("error reading resource Lake Formation Resource (%s): empty response", d.Id())
	}

	// d.Set("arn", output.ResourceInfo.ResourceArn) // output not including resource arn currently
	d.Set("role_arn", output.ResourceInfo.RoleArn)
	if output.ResourceInfo.LastModified != nil { // output not including last modified currently
		d.Set("last_modified", output.ResourceInfo.LastModified.Format(time.RFC3339))
	}

	return nil
}

func resourceAwsLakeFormationResourceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn
	resourceArn := d.Get("arn").(string)

	input := &lakeformation.DeregisterResourceInput{
		ResourceArn: aws.String(resourceArn),
	}

	_, err := conn.DeregisterResource(input)
	if err != nil {
		return fmt.Errorf("error deregistering Lake Formation Resource (%s): %w", d.Id(), err)
	}

	return nil
}
