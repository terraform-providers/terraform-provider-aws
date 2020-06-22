package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsWorkspaceBundle() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsWorkspaceBundleRead,

		Schema: map[string]*schema.Schema{
			"bundle_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"owner", "name"},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"bundle_id"},
			},
			"owner": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"bundle_id"},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"compute_type": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"user_storage": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"capacity": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"root_storage": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"capacity": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAwsWorkspaceBundleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn

	var bundle *workspaces.WorkspaceBundle

	if bundleID, ok := d.GetOk("bundle_id"); ok {
		resp, err := conn.DescribeWorkspaceBundles(&workspaces.DescribeWorkspaceBundlesInput{
			BundleIds: []*string{aws.String(bundleID.(string))},
		})
		if err != nil {
			return err
		}

		if len(resp.Bundles) != 1 {
			return fmt.Errorf("expected 1 result for Workspace bundle %q, found %d", bundleID, len(resp.Bundles))
		}

		bundle = resp.Bundles[0]

		if bundle == nil {
			return fmt.Errorf("no Workspace bundle with ID %q found", bundleID)
		}
	}

	if name, ok := d.GetOk("name"); ok {
		name := name.(string)
		input := &workspaces.DescribeWorkspaceBundlesInput{
			Owner: aws.String(d.Get("owner").(string)),
		}
		err := conn.DescribeWorkspaceBundlesPages(input, func(out *workspaces.DescribeWorkspaceBundlesOutput, lastPage bool) bool {
			for _, b := range out.Bundles {
				if aws.StringValue(b.Name) == name {
					bundle = b
					return true
				}
			}

			return !lastPage
		})
		if err != nil {
			return err
		}

		if bundle == nil {
			return fmt.Errorf("no Workspace bundle with name %q found", name)
		}
	}

	d.SetId(aws.StringValue(bundle.BundleId))
	d.Set("bundle_id", aws.StringValue(bundle.BundleId))
	d.Set("description", aws.StringValue(bundle.Description))
	d.Set("name", aws.StringValue(bundle.Name))
	d.Set("owner", aws.StringValue(bundle.Owner))

	computeType := make([]map[string]interface{}, 1)
	if bundle.ComputeType != nil {
		computeType[0] = map[string]interface{}{
			"name": aws.StringValue(bundle.ComputeType.Name),
		}
	}
	if err := d.Set("compute_type", computeType); err != nil {
		return fmt.Errorf("error setting compute_type: %s", err)
	}

	rootStorage := make([]map[string]interface{}, 1)
	if bundle.RootStorage != nil {
		rootStorage[0] = map[string]interface{}{
			"capacity": aws.StringValue(bundle.RootStorage.Capacity),
		}
	}
	if err := d.Set("root_storage", rootStorage); err != nil {
		return fmt.Errorf("error setting root_storage: %s", err)
	}

	userStorage := make([]map[string]interface{}, 1)
	if bundle.UserStorage != nil {
		userStorage[0] = map[string]interface{}{
			"capacity": aws.StringValue(bundle.UserStorage.Capacity),
		}
	}
	if err := d.Set("user_storage", userStorage); err != nil {
		return fmt.Errorf("error setting user_storage: %s", err)
	}

	return nil
}
