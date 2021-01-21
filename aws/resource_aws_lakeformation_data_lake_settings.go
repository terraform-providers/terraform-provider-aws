package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
)

func resourceAwsLakeFormationDataLakeSettings() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLakeFormationDataLakeSettingsCreate,
		Update: resourceAwsLakeFormationDataLakeSettingsCreate,
		Read:   resourceAwsLakeFormationDataLakeSettingsRead,
		Delete: resourceAwsLakeFormationDataLakeSettingsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"admins": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateArn,
				},
			},
			"catalog_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"create_database_default_permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				MaxItems: 3,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"permissions": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(lakeformation.Permission_Values(), false),
							},
						},
						"principal": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validatePrincipal,
						},
					},
				},
			},
			"create_table_default_permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				MaxItems: 3,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"permissions": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(lakeformation.Permission_Values(), false),
							},
						},
						"principal": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validatePrincipal,
						},
					},
				},
			},
			"trusted_resource_owners": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateAwsAccountId,
				},
			},
		},
	}
}

func resourceAwsLakeFormationDataLakeSettingsCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn

	input := &lakeformation.PutDataLakeSettingsInput{}

	if v, ok := d.GetOk("catalog_id"); ok {
		input.CatalogId = aws.String(v.(string))
	}

	settings := &lakeformation.DataLakeSettings{}

	if v, ok := d.GetOk("create_database_default_permissions"); ok {
		settings.CreateDatabaseDefaultPermissions = expandDataLakeSettingsCreateDefaultPermissions(v.([]interface{}))
	}

	if v, ok := d.GetOk("create_table_default_permissions"); ok {
		settings.CreateTableDefaultPermissions = expandDataLakeSettingsCreateDefaultPermissions(v.([]interface{}))
	}

	if v, ok := d.GetOk("admins"); ok {
		settings.DataLakeAdmins = expandDataLakeSettingsAdmins(v.([]interface{}))
	}

	if v, ok := d.GetOk("trusted_resource_owners"); ok {
		settings.TrustedResourceOwners = expandStringList(v.([]interface{}))
	}

	input.DataLakeSettings = settings

	var output *lakeformation.PutDataLakeSettingsOutput
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		var err error
		output, err = conn.PutDataLakeSettings(input)
		if err != nil {
			if isAWSErr(err, lakeformation.ErrCodeInvalidInputException, "Invalid principal") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, lakeformation.ErrCodeConcurrentModificationException, "") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(fmt.Errorf("error creating Lake Formation data lake settings: %w", err))
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		output, err = conn.PutDataLakeSettings(input)
	}

	if err != nil {
		return fmt.Errorf("error creating Lake Formation data lake settings: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error creating Lake Formation data lake settings: empty response")
	}

	d.SetId(fmt.Sprintf("%d", hashcode.String(input.String())))

	return resourceAwsLakeFormationDataLakeSettingsRead(d, meta)
}

func resourceAwsLakeFormationDataLakeSettingsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn

	input := &lakeformation.GetDataLakeSettingsInput{}

	if v, ok := d.GetOk("catalog_id"); ok {
		input.CatalogId = aws.String(v.(string))
	}

	output, err := conn.GetDataLakeSettings(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeEntityNotFoundException) {
		log.Printf("[WARN] Lake Formation data lake settings (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Lake Formation data lake settings (%s): %w", d.Id(), err)
	}

	if output == nil || output.DataLakeSettings == nil {
		return fmt.Errorf("reading Lake Formation data lake settings (%s): empty response", d.Id())
	}

	settings := output.DataLakeSettings

	d.Set("create_database_default_permissions", flattenDataLakeSettingsCreateDefaultPermissions(settings.CreateDatabaseDefaultPermissions))
	d.Set("create_table_default_permissions", flattenDataLakeSettingsCreateDefaultPermissions(settings.CreateTableDefaultPermissions))
	d.Set("admins", flattenDataLakeSettingsAdmins(settings.DataLakeAdmins))
	d.Set("trusted_resource_owners", flattenStringList(settings.TrustedResourceOwners))

	return nil
}

func resourceAwsLakeFormationDataLakeSettingsDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn

	input := &lakeformation.PutDataLakeSettingsInput{
		DataLakeSettings: &lakeformation.DataLakeSettings{
			CreateDatabaseDefaultPermissions: make([]*lakeformation.PrincipalPermissions, 0),
			CreateTableDefaultPermissions:    make([]*lakeformation.PrincipalPermissions, 0),
			DataLakeAdmins:                   make([]*lakeformation.DataLakePrincipal, 0),
			TrustedResourceOwners:            make([]*string, 0),
		},
	}

	if v, ok := d.GetOk("catalog_id"); ok {
		input.CatalogId = aws.String(v.(string))
	}

	_, err := conn.PutDataLakeSettings(input)

	if tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeEntityNotFoundException) {
		log.Printf("[WARN] Lake Formation data lake settings (%s) not found, removing from state", d.Id())
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Lake Formation data lake settings (%s): %w", d.Id(), err)
	}

	return nil
}

func expandDataLakeSettingsCreateDefaultPermissions(tfMaps []interface{}) []*lakeformation.PrincipalPermissions {
	apiObjects := make([]*lakeformation.PrincipalPermissions, 0, len(tfMaps))

	for _, tfMap := range tfMaps {
		apiObjects = append(apiObjects, expandDataLakeSettingsCreateDefaultPermission(tfMap.(map[string]interface{})))
	}

	return apiObjects
}

func expandDataLakeSettingsCreateDefaultPermission(tfMap map[string]interface{}) *lakeformation.PrincipalPermissions {
	apiObject := &lakeformation.PrincipalPermissions{
		Permissions: expandStringSet(tfMap["permissions"].(*schema.Set)),
		Principal: &lakeformation.DataLakePrincipal{
			DataLakePrincipalIdentifier: aws.String(tfMap["principal"].(string)),
		},
	}

	return apiObject
}

func flattenDataLakeSettingsCreateDefaultPermissions(apiObjects []*lakeformation.PrincipalPermissions) []map[string]interface{} {
	if apiObjects == nil {
		return nil
	}

	tfMaps := make([]map[string]interface{}, len(apiObjects))
	for i, v := range apiObjects {
		tfMaps[i] = flattenDataLakeSettingsCreateDefaultPermission(v)
	}

	return tfMaps
}

func flattenDataLakeSettingsCreateDefaultPermission(apiObject *lakeformation.PrincipalPermissions) map[string]interface{} {
	tfMap := make(map[string]interface{})

	if apiObject == nil {
		return tfMap
	}

	if apiObject.Permissions != nil {
		tfMap["permissions"] = flattenStringSet(apiObject.Permissions)
	}

	if v := aws.StringValue(apiObject.Principal.DataLakePrincipalIdentifier); v != "" {
		tfMap["principal"] = v
	}

	return tfMap
}

func expandDataLakeSettingsAdmins(tfSlice []interface{}) []*lakeformation.DataLakePrincipal {
	apiObjects := make([]*lakeformation.DataLakePrincipal, 0, len(tfSlice))

	for _, tfItem := range tfSlice {
		val, ok := tfItem.(string)
		if ok && val != "" {
			apiObjects = append(apiObjects, &lakeformation.DataLakePrincipal{
				DataLakePrincipalIdentifier: aws.String(tfItem.(string)),
			})
		}
	}

	return apiObjects
}

func flattenDataLakeSettingsAdmins(apiObjects []*lakeformation.DataLakePrincipal) []interface{} {
	if apiObjects == nil {
		return nil
	}

	tfSlice := make([]interface{}, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfSlice = append(tfSlice, *apiObject.DataLakePrincipalIdentifier)
	}

	return tfSlice
}
