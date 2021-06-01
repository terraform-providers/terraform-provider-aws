// +build !generate

package keyvaluetags

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
)

// Custom S3control tagging functions using similar formatting as other service generated code.

// S3controlBucketListTags lists S3control bucket tags.
// The identifier is the bucket ARN.
func S3controlBucketListTags(conn *s3control.S3Control, identifier string) (KeyValueTags, error) {
	parsedArn, err := arn.Parse(identifier)

	if err != nil {
		return New(nil), fmt.Errorf("error parsing S3 Control Bucket ARN (%s): %w", identifier, err)
	}

	input := &s3control.GetBucketTaggingInput{
		AccountId: aws.String(parsedArn.AccountID),
		Bucket:    aws.String(identifier),
	}

	output, err := conn.GetBucketTagging(input)

	if tfawserr.ErrCodeEquals(err, "NoSuchTagSet") {
		return New(nil), nil
	}

	if err != nil {
		return New(nil), err
	}

	return S3controlKeyValueTags(output.TagSet), nil
}

// S3controlBucketUpdateTags updates S3control bucket tags.
// The identifier is the bucket ARN.
func S3controlBucketUpdateTags(conn *s3control.S3Control, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	parsedArn, err := arn.Parse(identifier)

	if err != nil {
		return fmt.Errorf("error parsing S3 Control Bucket ARN (%s): %w", identifier, err)
	}

	oldTags := New(oldTagsMap)
	newTags := New(newTagsMap)

	// We need to also consider any existing ignored tags.
	allTags, err := S3controlBucketListTags(conn, identifier)

	if err != nil {
		return fmt.Errorf("error listing resource tags (%s): %w", identifier, err)
	}

	ignoredTags := allTags.Ignore(oldTags).Ignore(newTags)

	if len(newTags)+len(ignoredTags) > 0 {
		input := &s3control.PutBucketTaggingInput{
			AccountId: aws.String(parsedArn.AccountID),
			Bucket:    aws.String(identifier),
			Tagging: &s3control.Tagging{
				TagSet: newTags.Merge(ignoredTags).S3controlTags(),
			},
		}

		_, err := conn.PutBucketTagging(input)

		if err != nil {
			return fmt.Errorf("error setting resource tags (%s): %w", identifier, err)
		}
	} else if len(oldTags) > 0 && len(ignoredTags) == 0 {
		input := &s3control.DeleteBucketTaggingInput{
			AccountId: aws.String(parsedArn.AccountID),
			Bucket:    aws.String(identifier),
		}

		_, err := conn.DeleteBucketTagging(input)

		if err != nil {
			return fmt.Errorf("error deleting resource tags (%s): %w", identifier, err)
		}
	}

	return nil
}
