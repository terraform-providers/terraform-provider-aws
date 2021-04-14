package lister

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
)

// Custom Kinesisanalytics listing functions using similar formatting as other service generated code.

func ListApplicationsPages(conn *kinesisanalytics.KinesisAnalytics, input *kinesisanalytics.ListApplicationsInput, fn func(*kinesisanalytics.ListApplicationsOutput, bool) bool) error {
	for {
		output, err := conn.ListApplications(input)
		if err != nil {
			return err
		}

		lastPage := !aws.BoolValue(output.HasMoreApplications)
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.ExclusiveStartApplicationName = output.ApplicationSummaries[len(output.ApplicationSummaries)-1].ApplicationName
	}
	return nil
}
