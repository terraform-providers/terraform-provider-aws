package waiter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// ImageStatus fetches the Image and its Status
func ImageStatus(conn *imagebuilder.Imagebuilder, imageBuildVersionArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &imagebuilder.GetImageInput{
			ImageBuildVersionArn: aws.String(imageBuildVersionArn),
		}

		output, err := conn.GetImage(input)

		if err != nil {
			return nil, imagebuilder.ImageStatusPending, err
		}

		if output == nil || output.Image == nil || output.Image.State == nil {
			return nil, imagebuilder.ImageStatusPending, nil
		}

		status := aws.StringValue(output.Image.State.Status)

		if status == imagebuilder.ImageStatusFailed {
			return output.Image, status, fmt.Errorf("%s", aws.StringValue(output.Image.State.Reason))
		}

		return output.Image, status, nil
	}
}
