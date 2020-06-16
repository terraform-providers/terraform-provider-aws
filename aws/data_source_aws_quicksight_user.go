package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsQuickSightUser() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsQuickSightUserRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identity_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_role": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"active": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"namespace": {
				Type:     schema.TypeString,
				Required: true,
			},
			"aws_account_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceAwsQuickSightUserRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).quicksightconn
	userName := d.Get("user_name").(string)
	req := &quicksight.DescribeUserInput{
		UserName:     aws.String(userName),
		AwsAccountId: aws.String(d.Get("aws_account_id").(string)),
		Namespace:    aws.String(d.Get("namespace").(string)),
	}

	log.Printf("[DEBUG] Reading quicksight User: %s", req)
	resp, err := conn.DescribeUser(req)
	if err != nil {
		return fmt.Errorf("error getting user: %s", err)
	}

	user := resp.User
	d.SetId(aws.StringValue(user.PrincipalId))
	d.Set("arn", user.Arn)
	d.Set("user_role", user.Role)
	d.Set("email", user.Email)
	d.Set("active", user.Active)
	d.Set("user_id", user.PrincipalId)

	if user.IdentityType != nil {
		d.Set("identity_type", user.IdentityType)
	}

	return nil
}
