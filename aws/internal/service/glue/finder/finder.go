package finder

import (
	"github.com/aws/aws-sdk-go/service/glue"
	tfglue "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/glue"
)

// RegistryByID returns the Registry corresponding to the specified ID.
func RegistryByID(conn *glue.Glue, id string) (*glue.GetRegistryOutput, error) {
	input := &glue.GetRegistryInput{
		RegistryId: tfglue.CreateAwsGlueRegistryID(id),
	}

	output, err := conn.GetRegistry(input)
	if err != nil {
		return nil, err
	}

	return output, nil
}
