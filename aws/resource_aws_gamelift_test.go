package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

type testAccGameliftGame struct {
	Location   *gamelift.S3Location
	LaunchPath string
}

func (gg *testAccGameliftGame) Parameters(portNumber int) string {
	return fmt.Sprintf("+sv_port %d +gamelift_start_server", portNumber)
}

// Location found from CloudTrail event after finishing tutorial
// e.g. https://us-west-2.console.aws.amazon.com/gamelift/home?region=us-west-2#/r/fleets/sample
func testAccAWSGameliftSampleGame(region string) (*testAccGameliftGame, error) {
	version := "v1.2.0.0"
	accId, err := testAccGameliftAccountIdByRegion(region)
	if err != nil {
		return nil, err
	}
	bucket := fmt.Sprintf("gamelift-sample-builds-prod-%s", region)
	key := fmt.Sprintf("%s/server/sample_build_%s", version, version)
	roleArn := fmt.Sprintf("arn:%s:iam::%s:role/sample-build-upload-role-%s", testAccGetPartition(), accId, region)
	launchPath := `C:\game\Bin64.Release.Dedicated\MultiplayerProjectLauncher_Server.exe`

	gg := &testAccGameliftGame{
		Location: &gamelift.S3Location{
			Bucket:  aws.String(bucket),
			Key:     aws.String(key),
			RoleArn: aws.String(roleArn),
		},
		LaunchPath: launchPath,
	}

	return gg, nil
}

// Account ID found from CloudTrail event (role ARN) after finishing tutorial in given region
func testAccGameliftAccountIdByRegion(region string) (string, error) {
	m := map[string]string{
		"ap-northeast-1": "120069834884",
		"ap-northeast-2": "805673136642",
		"ap-south-1":     "134975661615",
		"ap-southeast-1": "077577004113",
		"ap-southeast-2": "112188327105",
		"ca-central-1":   "800535022691",
		"eu-central-1":   "797584052317",
		"eu-west-1":      "319803218673",
		"eu-west-2":      "937342764187",
		"sa-east-1":      "028872612690",
		"us-east-1":      "783764748367",
		"us-east-2":      "415729564621",
		"us-west-1":      "715879310420",
		"us-west-2":      "741061592171",
	}

	if accId, ok := m[region]; ok {
		return accId, nil
	}

	return "", &resource.NotFoundError{Message: fmt.Sprintf("GameLift Account ID not found for region %q", region)}
}
