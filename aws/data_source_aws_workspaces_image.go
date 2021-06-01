package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsWorkspacesImage() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsWorkspacesImageRead,

		Schema: map[string]*schema.Schema{
			"image_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"operating_system_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"required_tenancy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsWorkspacesImageRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn

	imageID := d.Get("image_id").(string)
	input := &workspaces.DescribeWorkspaceImagesInput{
		ImageIds: []*string{aws.String(imageID)},
	}

	resp, err := conn.DescribeWorkspaceImages(input)
	if err != nil {
		return fmt.Errorf("Failed describe workspaces images: %w", err)
	}
	if len(resp.Images) == 0 {
		return fmt.Errorf("Workspace image %s was not found", imageID)
	}

	image := resp.Images[0]
	d.SetId(imageID)
	d.Set("name", image.Name)
	d.Set("description", image.Description)
	d.Set("operating_system_type", image.OperatingSystem.Type)
	d.Set("required_tenancy", image.RequiredTenancy)
	d.Set("state", image.State)

	return nil
}
