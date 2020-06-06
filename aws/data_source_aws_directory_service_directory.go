package aws

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
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
							Set:      schema.HashString,
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
							Set:      schema.HashString,
						},
						"customer_username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"customer_dns_ips": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
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
				Set:      schema.HashString,
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
		return err
	}

	if len(out.DirectoryDescriptions) == 0 {
		log.Printf("[WARN] Directory %s not found", d.Id())
		d.SetId("")
		return nil
	}

	d.SetId(directoryID)

	dir := out.DirectoryDescriptions[0]
	log.Printf("[DEBUG] Received DS directory: %s", dir)

	d.Set("access_url", dir.AccessUrl)
	d.Set("alias", dir.Alias)
	d.Set("description", dir.Description)

	if *dir.Type == directoryservice.DirectoryTypeAdconnector {
		d.Set("dns_ip_addresses", schema.NewSet(schema.HashString, flattenStringList(dir.ConnectSettings.ConnectIps)))
	} else {
		d.Set("dns_ip_addresses", schema.NewSet(schema.HashString, flattenStringList(dir.DnsIpAddrs)))
	}
	d.Set("name", dir.Name)
	d.Set("short_name", dir.ShortName)
	d.Set("size", dir.Size)
	d.Set("edition", dir.Edition)
	d.Set("type", dir.Type)

	if err := d.Set("vpc_settings", flattenDSVpcSettings(dir.VpcSettings)); err != nil {
		return fmt.Errorf("error setting VPC settings: %s", err)
	}

	if err := d.Set("connect_settings", flattenDSConnectSettings(dir.DnsIpAddrs, dir.ConnectSettings)); err != nil {
		return fmt.Errorf("error setting connect settings: %s", err)
	}

	d.Set("enable_sso", dir.SsoEnabled)

	if aws.StringValue(dir.Type) == directoryservice.DirectoryTypeAdconnector {
		d.Set("security_group_id", aws.StringValue(dir.ConnectSettings.SecurityGroupId))
	} else {
		d.Set("security_group_id", aws.StringValue(dir.VpcSettings.SecurityGroupId))
	}

	tags, err := keyvaluetags.DirectoryserviceListTags(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error listing tags for Directory Service Directory (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}
