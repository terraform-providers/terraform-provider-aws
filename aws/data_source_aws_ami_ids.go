package aws

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAwsAmiIds() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsAmiIdsRead,

		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"executable_users": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
			},
			"owners": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.NoZeroValues,
				},
			},
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"sort_ascending": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
		},
	}
}

func dataSourceAwsAmiIdsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	params := &ec2.DescribeImagesInput{
		Owners: expandStringList(d.Get("owners").([]interface{})),
	}

	if v, ok := d.GetOk("executable_users"); ok {
		params.ExecutableUsers = expandStringList(v.([]interface{}))
	}
	if v, ok := d.GetOk("filter"); ok {
		params.Filters = buildAwsDataSourceFilters(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Reading AMI IDs: %s", params)
	resp, err := conn.DescribeImages(params)
	if err != nil {
		return err
	}

	var filteredImages []*ec2.Image
	imageIds := make([]string, 0)

	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		for _, image := range resp.Images {
			// Check for a very rare case where the response would include no
			// image name. No name means nothing to attempt a match against,
			// therefore we are skipping such image.
			if image.Name == nil || *image.Name == "" {
				log.Printf("[WARN] Unable to find AMI name to match against "+
					"for image ID %q owned by %q, nothing to do.",
					*image.ImageId, *image.OwnerId)
				continue
			}
			if r.MatchString(*image.Name) {
				filteredImages = append(filteredImages, image)
			}
		}
	} else {
		filteredImages = resp.Images[:]
	}

	sort.Slice(filteredImages, func(i, j int) bool {
		itime, _ := time.Parse(time.RFC3339, aws.StringValue(filteredImages[i].CreationDate))
		jtime, _ := time.Parse(time.RFC3339, aws.StringValue(filteredImages[j].CreationDate))
		if d.Get("sort_ascending").(bool) {
			return itime.Unix() < jtime.Unix()
		}
		return itime.Unix() > jtime.Unix()
	})
	for _, image := range filteredImages {
		imageIds = append(imageIds, *image.ImageId)
	}

	d.SetId(fmt.Sprintf("%d", hashcode.String(params.String())))
	d.Set("ids", imageIds)

	return nil
}
