package equivalency

import (
	"testing"
)

func TestEquivalentBatchContainerPropertiesJSON(t *testing.T) {
	testCases := []struct {
		Name              string
		ApiJson           string
		ConfigurationJson string
		ExpectEquivalent  bool
		ExpectError       bool
	}{
		{
			Name:              "empty",
			ApiJson:           ``,
			ConfigurationJson: ``,
			ExpectEquivalent:  true,
		},
		{
			Name: "empty ResourceRequirements",
			ApiJson: `
{
	"command": ["ls", "-la"],
	"environment": [
		{
			"name": "VARNAME",
			"value": "VARVAL"
		}
	],
	"image": "busybox",
	"memory":512,
	"mountPoints": [
		{
			"containerPath": "/tmp",
			"readOnly": false,
			"sourceVolume": "tmp"
		}
	],
	"resourceRequirements": [],
	"ulimits": [
		{
			"hardLimit": 1024,
			"name": "nofile",
			"softLimit": 1024
		}
	],
	"vcpus": 1,
	"volumes": [
		{
			"host": {
				"sourcePath": "/tmp"
			},
			"name": "tmp"
		}
	]
}
`,
			ConfigurationJson: `
{
	"command": ["ls", "-la"],
	"environment": [
		{
			"name": "VARNAME",
			"value": "VARVAL"
		}
	],
	"image": "busybox",
	"memory":512,
	"mountPoints": [
		{
			"containerPath": "/tmp",
			"readOnly": false,
			"sourceVolume": "tmp"
		}
	],
	"ulimits": [
		{
			"hardLimit": 1024,
			"name": "nofile",
			"softLimit": 1024
		}
	],
	"vcpus": 1,
	"volumes": [
		{
			"host": {
				"sourcePath": "/tmp"
			},
			"name": "tmp"
		}
	]
}
`,
			ExpectEquivalent: true,
		},
		{
			Name: "reordered Environment",
			ApiJson: `
{
	"command": ["ls", "-la"],
	"environment": [
		{
			"name": "VARNAME1",
			"value": "VARVAL1"
		},
		{
			"name": "VARNAME2",
			"value": "VARVAL2"
		}
	],
	"image": "busybox",
	"memory":512,
	"mountPoints": [
		{
			"containerPath": "/tmp",
			"readOnly": false,
			"sourceVolume": "tmp"
		}
	],
	"resourceRequirements": [],
	"ulimits": [
		{
			"hardLimit": 1024,
			"name": "nofile",
			"softLimit": 1024
		}
	],
	"vcpus": 1,
	"volumes": [
		{
			"host": {
				"sourcePath": "/tmp"
			},
			"name": "tmp"
		}
	]
}
`,
			ConfigurationJson: `
{
	"command": ["ls", "-la"],
	"environment": [
		{
			"name": "VARNAME2",
			"value": "VARVAL2"
		},
		{
			"name": "VARNAME1",
			"value": "VARVAL1"
		}
	],
	"image": "busybox",
	"memory":512,
	"mountPoints": [
		{
			"containerPath": "/tmp",
			"readOnly": false,
			"sourceVolume": "tmp"
		}
	],
	"resourceRequirements": [],
	"ulimits": [
		{
			"hardLimit": 1024,
			"name": "nofile",
			"softLimit": 1024
		}
	],
	"vcpus": 1,
	"volumes": [
		{
			"host": {
				"sourcePath": "/tmp"
			},
			"name": "tmp"
		}
	]
}
`,
			ExpectEquivalent: true,
		},
		{
			Name: "empty environment, mountPoints, ulimits, and volumes",
			ApiJson: `
{
	"image": "example:image",
	"vcpus": 8,
	"memory": 2048,
	"command": ["start.py", "Ref::S3bucket", "Ref::S3key"],
	"jobRoleArn": "arn:aws:iam::123456789012:role/example",
	"volumes": [],
	"environment": [],
	"mountPoints": [],
	"ulimits": [],
	"resourceRequirements": []
}
`,
			ConfigurationJson: `
{
	"command": ["start.py", "Ref::S3bucket", "Ref::S3key"],
	"image": "example:image",
	"memory": 2048,
	"vcpus": 8,
	"jobRoleArn": "arn:aws:iam::123456789012:role/example"
}
`,
			ExpectEquivalent: true,
		},
		{
			Name: "empty command, mountPoints, resourceRequirements, secrets, ulimits, volumes",
			ApiJson: `
{
	"image": "123.dkr.ecr.us-east-1.amazonaws.com/my-app",
	"vcpus": 1,
	"memory": 4096,
	"command": [],
	"jobRoleArn": "arn:aws:iam::123:role/role-test",
	"volumes": [],
	"environment": [{"name":"ENVIRONMENT","value":"test"}],
	"mountPoints": [],
	"ulimits": [],
	"resourceRequirements": [],
	"secrets": []
}
`,
			ConfigurationJson: `
{
    "image": "123.dkr.ecr.us-east-1.amazonaws.com/my-app",
    "memory": 4096,
    "vcpus": 1,
    "jobRoleArn": "arn:aws:iam::123:role/role-test",
    "environment": [
      {
        "name": "ENVIRONMENT",
        "value": "test"
      }
   ]
}
`,
			ExpectEquivalent: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			got, err := EquivalentBatchContainerPropertiesJSON(testCase.ConfigurationJson, testCase.ApiJson)

			if err != nil && !testCase.ExpectError {
				t.Errorf("got unexpected error: %s", err)
			}

			if err == nil && testCase.ExpectError {
				t.Errorf("expected error, but received none")
			}

			if got != testCase.ExpectEquivalent {
				t.Errorf("got %t, expected %t", got, testCase.ExpectEquivalent)
			}
		})
	}
}
