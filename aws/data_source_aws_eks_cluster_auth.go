package aws

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/eks/token"
)

func dataSourceAwsEksClusterAuth() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEksClusterAuthRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceAwsEksClusterAuthRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).stsconn
	name := d.Get("name").(string)
	generator, err := token.NewGenerator(false, false)
	if err != nil {
		return fmt.Errorf("error getting token generator: %w", err)
	}
	token, err := generator.GetWithSTS(name, conn)
	if err != nil {
		return fmt.Errorf("error getting token: %w", err)
	}

	d.SetId(name)
	d.Set("token", token.Token)

	return nil
}
