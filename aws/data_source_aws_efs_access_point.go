package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsEfsAccessPoint() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEfsAccessPointRead,

		Schema: map[string]*schema.Schema{
			"access_point_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"file_system_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_system_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"posix_user": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gid": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"uid": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"secondary_gids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeInt},
							Set:      schema.HashInt,
							Computed: true,
						},
					},
				},
			},
			"root_directory": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"path": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"creation_info": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"owner_gid": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"owner_uid": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"permissions": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func dataSourceAwsEfsAccessPointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeAccessPoints(&efs.DescribeAccessPointsInput{
		AccessPointId: aws.String(d.Get("access_point_id").(string)),
	})
	if err != nil {
		return fmt.Errorf("Error reading EFS access point %s: %s", d.Id(), err)
	}
	if len(resp.AccessPoints) != 1 {
		return fmt.Errorf("Search returned %d results, please revise so only one is returned", len(resp.AccessPoints))
	}

	ap := resp.AccessPoints[0]

	log.Printf("[DEBUG] Found EFS access point: %#v", ap)

	d.SetId(*ap.AccessPointId)

	fsARN := arn.ARN{
		AccountID: meta.(*AWSClient).accountid,
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("file-system/%s", aws.StringValue(ap.FileSystemId)),
		Service:   "elasticfilesystem",
	}.String()

	d.Set("file_system_arn", fsARN)
	d.Set("file_system_id", ap.FileSystemId)
	d.Set("arn", ap.AccessPointArn)
	d.Set("owner_id", ap.OwnerId)

	if err := d.Set("posix_user", flattenEfsAccessPointPosixUser(ap.PosixUser)); err != nil {
		return fmt.Errorf("error setting posix user: %s", err)
	}

	if err := d.Set("root_directory", flattenEfsAccessPointRootDirectory(ap.RootDirectory)); err != nil {
		return fmt.Errorf("error setting root directory: %s", err)
	}

	if err := d.Set("tags", keyvaluetags.EfsKeyValueTags(ap.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}
