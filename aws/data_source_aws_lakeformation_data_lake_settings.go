package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
)

func dataSourceAwsLakeFormationDataLakeSettings() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLakeFormationDataLakeSettingsRead,

		Schema: map[string]*schema.Schema{
			"admins": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"catalog_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"create_database_default_permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"permissions": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"principal": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"create_table_default_permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"permissions": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"principal": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"trusted_resource_owners": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsLakeFormationDataLakeSettingsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn

	input := &lakeformation.GetDataLakeSettingsInput{}

	if v, ok := d.GetOk("catalog_id"); ok {
		input.CatalogId = aws.String(v.(string))
	}
	d.SetId(fmt.Sprintf("%d", hashcode.String(input.String())))

	output, err := conn.GetDataLakeSettings(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeEntityNotFoundException) {
		log.Printf("[WARN] Lake Formation data lake settings (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Lake Formation data lake settings (%s): %w", d.Id(), err)
	}

	if output == nil || output.DataLakeSettings == nil {
		return fmt.Errorf("error reading Lake Formation data lake settings (%s): empty response", d.Id())
	}

	settings := output.DataLakeSettings

	d.Set("create_database_default_permissions", flattenDataLakeSettingsCreateDefaultPermissions(settings.CreateDatabaseDefaultPermissions))
	d.Set("create_table_default_permissions", flattenDataLakeSettingsCreateDefaultPermissions(settings.CreateTableDefaultPermissions))
	d.Set("admins", flattenDataLakeSettingsAdmins(settings.DataLakeAdmins))
	d.Set("trusted_resource_owners", flattenStringList(settings.TrustedResourceOwners))

	return nil
}
