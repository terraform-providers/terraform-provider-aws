package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsDirectoryServiceDirectory() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsDirectoryServiceDirectoryRead,

		Schema: map[string]*schema.Schema{
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"size": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"short_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
			"vpc_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"availability_zones": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"connect_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connect_ips": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"customer_username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"customer_dns_ips": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"availability_zones": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"enable_sso": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"access_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_ip_addresses": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"edition": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsDirectoryServiceDirectoryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	directoryID := d.Get("directory_id").(string)
	out, err := conn.DescribeDirectories(&directoryservice.DescribeDirectoriesInput{
		DirectoryIds: []*string{aws.String(directoryID)},
	})
	if err != nil {
		if isAWSErr(err, directoryservice.ErrCodeEntityDoesNotExistException, "") {
			return fmt.Errorf("DirectoryService Directory (%s) not found", directoryID)
		}
		return fmt.Errorf("error reading DirectoryService Directory: %w", err)
	}

	if out == nil || len(out.DirectoryDescriptions) == 0 {
		return fmt.Errorf("error reading DirectoryService Directory (%s): empty output", directoryID)
	}

	d.SetId(directoryID)

	dir := out.DirectoryDescriptions[0]
	log.Printf("[DEBUG] Received DS directory: %s", dir)

	d.Set("access_url", dir.AccessUrl)
	d.Set("alias", dir.Alias)
	d.Set("description", dir.Description)

	var addresses []interface{}
	if aws.StringValue(dir.Type) == directoryservice.DirectoryTypeAdconnector {
		addresses = flattenStringList(dir.ConnectSettings.ConnectIps)
	} else {
		addresses = flattenStringList(dir.DnsIpAddrs)
	}
	if err := d.Set("dns_ip_addresses", addresses); err != nil {
		return fmt.Errorf("error setting dns_ip_addresses: %w", err)
	}

	d.Set("name", dir.Name)
	d.Set("short_name", dir.ShortName)
	d.Set("size", dir.Size)
	d.Set("edition", dir.Edition)
	d.Set("type", dir.Type)

	if err := d.Set("vpc_settings", flattenDSVpcSettings(dir.VpcSettings)); err != nil {
		return fmt.Errorf("error setting VPC settings: %w", err)
	}

	if err := d.Set("connect_settings", flattenDSConnectSettings(dir.DnsIpAddrs, dir.ConnectSettings)); err != nil {
		return fmt.Errorf("error setting connect settings: %w", err)
	}

	d.Set("enable_sso", dir.SsoEnabled)

	var securityGroupId *string
	if aws.StringValue(dir.Type) == directoryservice.DirectoryTypeAdconnector {
		securityGroupId = dir.ConnectSettings.SecurityGroupId
	} else {
		securityGroupId = dir.VpcSettings.SecurityGroupId
	}
	d.Set("security_group_id", aws.StringValue(securityGroupId))

	tags, err := keyvaluetags.DirectoryserviceListTags(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error listing tags for Directory Service Directory (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}
