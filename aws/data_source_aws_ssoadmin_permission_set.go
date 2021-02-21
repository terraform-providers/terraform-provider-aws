package aws

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsSsoAdminPermissionSet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsSsoAdminPermissionSetRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateArn,
				ExactlyOneOf: []string{"arn", "name"},
			},

			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"instance_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexp.MustCompile(`[\w+=,.@-]+`), "must match [\\w+=,.@-]"),
				),
				ExactlyOneOf: []string{"name", "arn"},
			},

			"relay_state": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"session_duration": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsSsoAdminPermissionSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssoadminconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	instanceArn := d.Get("instance_arn").(string)

	var permissionSet *ssoadmin.PermissionSet

	if v, ok := d.GetOk("arn"); ok {
		arn := v.(string)

		input := &ssoadmin.DescribePermissionSetInput{
			InstanceArn:      aws.String(instanceArn),
			PermissionSetArn: aws.String(arn),
		}

		output, err := conn.DescribePermissionSet(input)
		if err != nil {
			return fmt.Errorf("error reading SSO Admin Permission Set (%s): %w", arn, err)
		}

		if output == nil {
			return fmt.Errorf("error reading SSO Admin Permission Set (%s): empty output", arn)
		}

		permissionSet = output.PermissionSet
	} else if v, ok := d.GetOk("name"); ok {
		name := v.(string)
		var describeErr error

		input := &ssoadmin.ListPermissionSetsInput{
			InstanceArn: aws.String(instanceArn),
		}

		err := conn.ListPermissionSetsPages(input, func(page *ssoadmin.ListPermissionSetsOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, permissionSetArn := range page.PermissionSets {
				if permissionSetArn == nil {
					continue
				}

				output, describeErr := conn.DescribePermissionSet(&ssoadmin.DescribePermissionSetInput{
					InstanceArn:      aws.String(instanceArn),
					PermissionSetArn: permissionSetArn,
				})

				if describeErr != nil {
					return false
				}

				if output == nil || output.PermissionSet == nil {
					continue
				}

				if aws.StringValue(output.PermissionSet.Name) == name {
					permissionSet = output.PermissionSet
					return false
				}
			}

			return !lastPage
		})

		if err != nil {
			return fmt.Errorf("error listing SSO Permission Sets: %w", err)
		}

		if describeErr != nil {
			return fmt.Errorf("error reading SSO Permission Set (%s): %w", name, describeErr)
		}
	}

	if permissionSet == nil {
		return errors.New("error reading SSO Permission Set: not found")
	}

	arn := aws.StringValue(permissionSet.PermissionSetArn)

	d.SetId(arn)
	d.Set("arn", arn)
	d.Set("created_date", permissionSet.CreatedDate.Format(time.RFC3339))
	d.Set("description", permissionSet.Description)
	d.Set("instance_arn", instanceArn)
	d.Set("name", permissionSet.Name)
	d.Set("session_duration", permissionSet.SessionDuration)
	d.Set("relay_state", permissionSet.RelayState)

	tags, err := keyvaluetags.SsoadminListTags(conn, arn, instanceArn)
	if err != nil {
		return fmt.Errorf("error listing tags for SSO Permission Set (%s): %w", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}
