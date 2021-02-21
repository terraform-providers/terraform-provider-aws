package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsEfsAccessPoints() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEfsAccessPointsRead,

		Schema: map[string]*schema.Schema{
			"arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"file_system_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsEfsAccessPointsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn

	fileSystemId := d.Get("file_system_id").(string)
	input := &efs.DescribeAccessPointsInput{
		FileSystemId: aws.String(fileSystemId),
	}

	var accessPoints []*efs.AccessPointDescription

	err := conn.DescribeAccessPointsPages(input, func(page *efs.DescribeAccessPointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		accessPoints = append(accessPoints, page.AccessPoints...)

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error reading EFS Access Points for File System (%s): %w", fileSystemId, err)
	}

	if len(accessPoints) == 0 {
		return fmt.Errorf("no matching EFS Access Points for File System (%s) found", fileSystemId)
	}

	d.SetId(fileSystemId)

	var arns, ids []string

	for _, accessPoint := range accessPoints {
		arns = append(arns, aws.StringValue(accessPoint.AccessPointArn))
		ids = append(ids, aws.StringValue(accessPoint.AccessPointId))
	}

	if err := d.Set("arns", arns); err != nil {
		return fmt.Errorf("error setting arns: %w", err)
	}

	if err := d.Set("ids", ids); err != nil {
		return fmt.Errorf("error setting ids: %w", err)
	}

	return nil
}
