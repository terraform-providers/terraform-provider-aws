package aws

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsSsoAdminRole() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsSsoAdminRoleRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"assume_role_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"permissions_boundary": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"unique_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"permission_set_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"create_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"max_session_duration": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsSsoAdminRoleRead(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	permissionSetName := d.Get("permission_set_name").(string)
	pathPrefix := aws.String("/aws-reserved/sso.amazonaws.com/")
	roleNamePrefix := fmt.Sprintf("AWSReservedSSO_%s_", permissionSetName)

	// Get all AWS SSO roles
	input := &iam.ListRolesInput{
		PathPrefix: pathPrefix,
	}

	roles := []*iam.Role{}

	err := iamconn.ListRolesPages(
		input,
		func(page *iam.ListRolesOutput, lastPage bool) bool {
			for _, role := range page.Roles {
				if strings.HasPrefix(aws.StringValue(role.RoleName), roleNamePrefix) {
					roles = append(roles, role)
				}
			}
			return !lastPage
		},
	)
	if err != nil {
		return fmt.Errorf("error reading roles with path prefix (%s): %w", aws.StringValue(pathPrefix), err)
	}

	if len(roles) > 1 {
		return fmt.Errorf("found too many SSO roles (%d) matching the permission set name", len(roles))
	}
	if len(roles) == 0 {
		return fmt.Errorf("couldn't find any SSO roles matching the permission set name")
	}

	role := roles[0]

	d.Set("arn", role.Arn)
	if err := d.Set("create_date", role.CreateDate.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("error setting create_date: %w", err)
	}
	d.Set("description", role.Description)
	d.Set("max_session_duration", role.MaxSessionDuration)
	d.Set("name", role.RoleName)
	d.Set("path", role.Path)
	d.Set("permissions_boundary", "")
	if role.PermissionsBoundary != nil {
		d.Set("permissions_boundary", role.PermissionsBoundary.PermissionsBoundaryArn)
	}
	d.Set("unique_id", role.RoleId)
	d.Set("tags", keyvaluetags.IamKeyValueTags(role.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map())

	assumRolePolicy, err := url.QueryUnescape(aws.StringValue(role.AssumeRolePolicyDocument))
	if err != nil {
		return fmt.Errorf("error parsing assume role policy document: %w", err)
	}
	if err := d.Set("assume_role_policy", assumRolePolicy); err != nil {
		return fmt.Errorf("error setting assume_role_policy: %w", err)
	}

	d.SetId(aws.StringValue(role.RoleName))

	return nil
}
