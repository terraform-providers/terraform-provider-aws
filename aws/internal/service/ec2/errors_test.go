package ec2_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
)

func TestUnsuccessfulItemError(t *testing.T) {
	unsuccessfulItemError := &ec2.UnsuccessfulItemError{
		Code:    aws.String("test code"),
		Message: aws.String("test message"),
	}

	err := tfec2.UnsuccessfulItemError(unsuccessfulItemError)

	if !tfawserr.ErrCodeEquals(err, "test code") {
		t.Errorf("tfawserr.ErrCodeEquals failed: %s", err)
	}

	if !tfawserr.ErrMessageContains(err, "test code", "est mess") {
		t.Errorf("tfawserr.ErrMessageContains failed: %s", err)
	}
}

func TestUnsuccessfulItemsError(t *testing.T) {
	testCases := []struct {
		Name     string
		Items    []*ec2.UnsuccessfulItem
		Expected bool
	}{
		{
			Name: "no items",
		},
		{
			Name: "one item no error",
			Items: []*ec2.UnsuccessfulItem{
				{
					ResourceId: aws.String("test resource"),
				},
			},
		},
		{
			Name: "one item",
			Items: []*ec2.UnsuccessfulItem{
				{
					Error: &ec2.UnsuccessfulItemError{
						Code:    aws.String("test code"),
						Message: aws.String("test message"),
					},
					ResourceId: aws.String("test resource"),
				},
			},
			Expected: true,
		},
		{
			Name: "two items, first no error",
			Items: []*ec2.UnsuccessfulItem{
				{
					ResourceId: aws.String("test resource 1"),
				},
				{
					Error: &ec2.UnsuccessfulItemError{
						Code:    aws.String("test code"),
						Message: aws.String("test message"),
					},
					ResourceId: aws.String("test resource 2"),
				},
			},
			Expected: true,
		},
		{
			Name: "two items, first not as expected",
			Items: []*ec2.UnsuccessfulItem{
				{
					Error: &ec2.UnsuccessfulItemError{
						Code:    aws.String("not what is required"),
						Message: aws.String("not what is wanted"),
					},
					ResourceId: aws.String("test resource 1"),
				},
				{
					Error: &ec2.UnsuccessfulItemError{
						Code:    aws.String("test code"),
						Message: aws.String("test message"),
					},
					ResourceId: aws.String("test resource 2"),
				},
			},
		},
		{
			Name: "two items, first as expected",
			Items: []*ec2.UnsuccessfulItem{
				{
					Error: &ec2.UnsuccessfulItemError{
						Code:    aws.String("test code"),
						Message: aws.String("test message"),
					},
					ResourceId: aws.String("test resource 1"),
				},
				{
					Error: &ec2.UnsuccessfulItemError{
						Code:    aws.String("not what is required"),
						Message: aws.String("not what is wanted"),
					},
					ResourceId: aws.String("test resource 2"),
				},
			},
			Expected: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			err := tfec2.UnsuccessfulItemsError(testCase.Items)

			got := tfawserr.ErrCodeEquals(err, "test code")

			if got != testCase.Expected {
				t.Errorf("ErrCodeEquals got %t, expected %t", got, testCase.Expected)
			}

			got = tfawserr.ErrMessageContains(err, "test code", "est mess")

			if got != testCase.Expected {
				t.Errorf("ErrMessageContains got %t, expected %t", got, testCase.Expected)
			}
		})
	}
}
