package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsLightsailInstance() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLightsailInstanceRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 255),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z]`), "must begin with an alphabetic character"),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_\-.]+[^._\-]$`), "must contain only alphanumeric characters, underscores, hyphens, and dots"),
				),
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"blueprint_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bundle_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			// Optional attributes
			"key_pair_name": {
				// Not compatible with aws_key_pair (yet)
				// We'll need a new aws_lightsail_key_pair resource
				Type:     schema.TypeString,
				Computed: true,
			},

			"ip_address_type": {
				Type:     schema.TypeString,
				Optional: true,
			},

			// cannot be retrieved from the API
			"user_data": {
				Type:     schema.TypeString,
				Computed: true,
			},

			// additional info returned from the API
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cpu_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ram_size": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"ipv6_address": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "use `ipv6_addresses` attribute instead",
			},
			"ipv6_addresses": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"is_static_ip": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"private_ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func dataSourceAwsLightsailInstanceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.GetInstance(&lightsail.GetInstanceInput{
		InstanceName: aws.String(d.Get("name").(string)),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NotFoundException" {
				log.Printf("[WARN] Lightsail Instance (%s) not found, removing from state", d.Id())
				return fmt.Errorf("no matching Lightsail Instance found")
			}
			return err
		}
		return err
	}

	if resp == nil {
		log.Printf("[WARN] Lightsail Instance (%s) not found, nil response from server, removing from state", d.Id())
		return fmt.Errorf("no matching Lightsail Instance found, nil response from server")
	}

	i := resp.Instance

	d.SetId(d.Get("name").(string))
	d.Set("availability_zone", i.Location.AvailabilityZone)
	d.Set("blueprint_id", i.BlueprintId)
	d.Set("bundle_id", i.BundleId)
	d.Set("key_pair_name", i.SshKeyName)
	d.Set("name", i.Name)

	// additional attributes
	d.Set("arn", i.Arn)
	d.Set("username", i.Username)
	d.Set("created_at", i.CreatedAt.Format(time.RFC3339))
	d.Set("cpu_count", i.Hardware.CpuCount)
	d.Set("ram_size", i.Hardware.RamSizeInGb)

	// Deprecated: AWS Go SDK v1.36.25 removed Ipv6Address field
	if len(i.Ipv6Addresses) > 0 {
		d.Set("ipv6_address", i.Ipv6Addresses[0])
	} else {
		// Setting empty value if no address returned
		d.Set("ipv6_address", "")
	}

	d.Set("ipv6_addresses", aws.StringValueSlice(i.Ipv6Addresses))
	d.Set("is_static_ip", i.IsStaticIp)
	d.Set("private_ip_address", i.PrivateIpAddress)
	d.Set("public_ip_address", i.PublicIpAddress)
	d.Set("ip_address_type", i.IpAddressType)

	if err := d.Set("tags", keyvaluetags.LightsailKeyValueTags(i.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}
