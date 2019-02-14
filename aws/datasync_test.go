package aws

import (
	"testing"
)

func TestDataSyncParseLocationURI(t *testing.T) {
	testCases := []struct {
		LocationURI  string
		Subdirectory string
	}{
		{
			LocationURI:  "efs://us-east-2.fs-abcd1234/",
			Subdirectory: "/",
		},
		{
			LocationURI:  "efs://us-east-2.fs-abcd1234/path",
			Subdirectory: "/path",
		},
		{
			LocationURI:  "nfs://example.com/",
			Subdirectory: "/",
		},
		{
			LocationURI:  "nfs://example.com/path",
			Subdirectory: "/path",
		},
		{
			LocationURI:  "s3://myBucket/",
			Subdirectory: "/",
		},
		{
			LocationURI:  "s3://myBucket/path",
			Subdirectory: "/path",
		},
	}

	for i, tc := range testCases {
		subdirectory, err := dataSyncParseLocationURI(tc.LocationURI)
		if err != nil {
			t.Fatalf("%d: received error parsing (%s): %s", i, tc.LocationURI, err)
		}
		if subdirectory != tc.Subdirectory {
			t.Fatalf("%d: expected subdirectory (%s), received: %s", i, tc.Subdirectory, subdirectory)
		}
	}
}
