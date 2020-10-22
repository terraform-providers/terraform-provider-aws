package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsEcrImage() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEcrImageRead,
		Schema: map[string]*schema.Schema{
			"registry_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"repository_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"image_digest": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"image_tag": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"image_pushed_at": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"image_size_in_bytes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"image_tags": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsEcrImageRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecrconn

	params := &ecr.DescribeImagesInput{
		RepositoryName: aws.String(d.Get("repository_name").(string)),
	}

	regId, ok := d.GetOk("registry_id")
	if ok {
		params.RegistryId = aws.String(regId.(string))
	}

	imgId := ecr.ImageIdentifier{}
	digest, ok := d.GetOk("image_digest")
	if ok {
		imgId.ImageDigest = aws.String(digest.(string))
	}
	tag, ok := d.GetOk("image_tag")
	if ok {
		imgId.ImageTag = aws.String(tag.(string))
	}

	if imgId.ImageDigest == nil && imgId.ImageTag == nil {
		return fmt.Errorf("At least one of either image_digest or image_tag must be defined")
	}

	params.ImageIds = []*ecr.ImageIdentifier{&imgId}

	var imageDetails []*ecr.ImageDetail
	log.Printf("[DEBUG] Reading ECR Images: %s", params)
	err := conn.DescribeImagesPages(params, func(page *ecr.DescribeImagesOutput, lastPage bool) bool {
		imageDetails = append(imageDetails, page.ImageDetails...)
		return true
	})
	if err != nil {
		return fmt.Errorf("Error describing ECR images: %q", err)
	}

	if len(imageDetails) == 0 {
		return fmt.Errorf("No matching image found")
	}
	if len(imageDetails) > 1 {
		return fmt.Errorf("More than one image found for tag/digest combination")
	}

	image := imageDetails[0]

	d.SetId(aws.StringValue(image.ImageDigest))
	if err = d.Set("registry_id", aws.StringValue(image.RegistryId)); err != nil {
		return fmt.Errorf("failed to set registry_id: %s", err)
	}
	if err = d.Set("image_digest", aws.StringValue(image.ImageDigest)); err != nil {
		return fmt.Errorf("failed to set image_digest: %s", err)
	}
	if err = d.Set("image_pushed_at", image.ImagePushedAt.Unix()); err != nil {
		return fmt.Errorf("failed to set image_pushed_at: %s", err)
	}
	if err = d.Set("image_size_in_bytes", aws.Int64Value(image.ImageSizeInBytes)); err != nil {
		return fmt.Errorf("failed to set image_size_in_bytes: %s", err)
	}
	if err := d.Set("image_tags", aws.StringValueSlice(image.ImageTags)); err != nil {
		return fmt.Errorf("failed to set image_tags: %s", err)
	}

	return nil
}
