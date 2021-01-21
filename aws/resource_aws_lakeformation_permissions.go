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

func resourceAwsLakeFormationPermissions() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLakeFormationPermissionsCreate,
		Read:   resourceAwsLakeFormationPermissionsRead,
		Update: resourceAwsLakeFormationPermissionsCreate,
		Delete: resourceAwsLakeFormationPermissionsDelete,

		Schema: map[string]*schema.Schema{
			"catalog_id": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				ValidateFunc: validateAwsAccountId,
			},
			"catalog_resource": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"data_location": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
						"catalog_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validateAwsAccountId,
						},
					},
				},
			},
			"database": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"catalog_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validateAwsAccountId,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"permissions": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(lakeformation.Permission_Values(), false),
				},
			},
			"permissions_with_grant_option": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(lakeformation.Permission_Values(), false),
				},
			},
			"principal": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validatePrincipal,
			},
			"table": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"catalog_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validateAwsAccountId,
						},
						"database_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"wildcard": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"table_with_columns": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"catalog_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validateAwsAccountId,
						},
						"column_names": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.NoZeroValues,
							},
						},
						"database_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"excluded_column_names": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.NoZeroValues,
							},
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsLakeFormationPermissionsCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn

	input := &lakeformation.GrantPermissionsInput{
		Permissions: expandStringList(d.Get("permissions").([]interface{})),
		Principal: &lakeformation.DataLakePrincipal{
			DataLakePrincipalIdentifier: aws.String(d.Get("principal").(string)),
		},
	}

	if v, ok := d.GetOk("catalog_id"); ok {
		input.CatalogId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("permissions_with_grant_option"); ok {
		input.PermissionsWithGrantOption = expandStringList(v.([]interface{}))
	}

	input.Resource = expandLakeFormationResource(d, false)

	var output *lakeformation.GrantPermissionsOutput
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		var err error
		output, err = conn.GrantPermissions(input)
		if err != nil {
			if isAWSErr(err, lakeformation.ErrCodeInvalidInputException, "Invalid principal") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, lakeformation.ErrCodeInvalidInputException, "Grantee has no permissions") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, lakeformation.ErrCodeInvalidInputException, "register the S3 path") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, lakeformation.ErrCodeConcurrentModificationException, "") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, "AccessDeniedException", "is not authorized to access requested permissions") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(fmt.Errorf("error creating Lake Formation Permissions: %w", err))
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		output, err = conn.GrantPermissions(input)
	}

	if err != nil {
		return fmt.Errorf("error creating Lake Formation Permissions (input: %v): %w", input, err)
	}

	if output == nil {
		return fmt.Errorf("error creating Lake Formation Permissions: empty response")
	}

	d.SetId(fmt.Sprintf("%d", hashcode.String(input.String())))

	return resourceAwsLakeFormationPermissionsRead(d, meta)
}

func resourceAwsLakeFormationPermissionsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn

	input := &lakeformation.ListPermissionsInput{
		Principal: &lakeformation.DataLakePrincipal{
			DataLakePrincipalIdentifier: aws.String(d.Get("principal").(string)),
		},
	}

	if v, ok := d.GetOk("catalog_id"); ok {
		input.CatalogId = aws.String(v.(string))
	}

	input.Resource = expandLakeFormationResource(d, true)

	log.Printf("[DEBUG] Reading Lake Formation permissions: %v", input)
	var principalResourcePermissions []*lakeformation.PrincipalResourcePermissions

	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		err := conn.ListPermissionsPages(input, func(resp *lakeformation.ListPermissionsOutput, lastPage bool) bool {
			for _, permission := range resp.PrincipalResourcePermissions {
				if permission == nil {
					continue
				}

				principalResourcePermissions = append(principalResourcePermissions, permission)
			}
			return !lastPage
		})

		if err != nil {
			if isAWSErr(err, lakeformation.ErrCodeInvalidInputException, "Invalid principal") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(fmt.Errorf("error creating Lake Formation Permissions: %w", err))
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		err = conn.ListPermissionsPages(input, func(resp *lakeformation.ListPermissionsOutput, lastPage bool) bool {
			for _, permission := range resp.PrincipalResourcePermissions {
				if permission == nil {
					continue
				}

				principalResourcePermissions = append(principalResourcePermissions, permission)
			}
			return !lastPage
		})
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeEntityNotFoundException) {
		log.Printf("[WARN] Resource Lake Formation permissions (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Lake Formation permissions: %w", err)
	}

	if len(principalResourcePermissions) > 1 {
		return fmt.Errorf("error reading Lake Formation permissions: %s", "multiple permissions found")
	}

	for _, permissions := range principalResourcePermissions {
		d.Set("principal", permissions.Principal.DataLakePrincipalIdentifier)
		d.Set("permissions", permissions.Permissions)
		d.Set("permissions_with_grant_option", permissions.PermissionsWithGrantOption)

		if permissions.Resource.Catalog != nil {
			d.Set("catalog_resource", true)
		}

		if permissions.Resource.DataLocation != nil {
			d.Set("data_location", []interface{}{flattenLakeFormationDataLocationResource(permissions.Resource.DataLocation)})
		} else {
			d.Set("data_location", nil)
		}

		if permissions.Resource.Database != nil {
			d.Set("database", []interface{}{flattenLakeFormationDatabaseResource(permissions.Resource.Database)})
		} else {
			d.Set("database", nil)
		}

		// table with columns permissions will include the table and table with columns
		if permissions.Resource.TableWithColumns != nil {
			d.Set("table_with_columns", []interface{}{flattenLakeFormationTableWithColumnsResource(permissions.Resource.TableWithColumns)})
		} else if permissions.Resource.Table != nil {
			d.Set("table_with_columns", nil)
			d.Set("table", []interface{}{flattenLakeFormationTableResource(permissions.Resource.Table)})
		} else {
			d.Set("table", nil)
		}
	}

	return nil
}

func resourceAwsLakeFormationPermissionsDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn

	input := &lakeformation.RevokePermissionsInput{
		Permissions: expandStringList(d.Get("permissions").([]interface{})),
		Principal: &lakeformation.DataLakePrincipal{
			DataLakePrincipalIdentifier: aws.String(d.Get("principal").(string)),
		},
	}

	if v, ok := d.GetOk("catalog_id"); ok {
		input.CatalogId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("permissions_with_grant_option"); ok {
		input.PermissionsWithGrantOption = expandStringList(v.([]interface{}))
	}

	input.Resource = expandLakeFormationResource(d, false)

	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		var err error
		_, err = conn.RevokePermissions(input)
		if err != nil {
			if isAWSErr(err, lakeformation.ErrCodeInvalidInputException, "register the S3 path") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, lakeformation.ErrCodeConcurrentModificationException, "") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(fmt.Errorf("unable to revoke Lake Formation Permissions: %w", err))
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.RevokePermissions(input)
	}

	if err != nil {
		return fmt.Errorf("unable to revoke LakeFormation Permissions (input: %v): %w", input, err)
	}

	return nil
}

func expandLakeFormationResource(d *schema.ResourceData, squashTableWithColumns bool) *lakeformation.Resource {
	res := &lakeformation.Resource{}

	if v, ok := d.GetOk("catalog_resource"); ok {
		if v.(bool) {
			res.Catalog = &lakeformation.CatalogResource{}
		}
	}

	if v, ok := d.GetOk("data_location"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		res.DataLocation = expandLakeFormationDataLocationResource(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("database"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		res.Database = expandLakeFormationDatabaseResource(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("table"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		res.Table = expandLakeFormationTableResource(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("table_with_columns"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		if squashTableWithColumns {
			// ListPermissions does not support getting privileges by tables with columns. Instead,
			// use the table which will return both table and table with columns.
			res.Table = expandLakeFormationTableResource(v.([]interface{})[0].(map[string]interface{}))
		} else {
			res.TableWithColumns = expandLakeFormationTableWithColumnsResource(v.([]interface{})[0].(map[string]interface{}))
		}
	}

	return res
}

func expandLakeFormationDataLocationResource(tfMap map[string]interface{}) *lakeformation.DataLocationResource {
	if tfMap == nil {
		return nil
	}

	apiObject := &lakeformation.DataLocationResource{}

	if v, ok := tfMap["catalog_id"].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}

	if v, ok := tfMap["arn"].(string); ok && v != "" {
		apiObject.ResourceArn = aws.String(v)
	}

	return apiObject
}

func flattenLakeFormationDataLocationResource(apiObject *lakeformation.DataLocationResource) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CatalogId; v != nil {
		tfMap["catalog_id"] = aws.StringValue(v)
	}

	if v := apiObject.ResourceArn; v != nil {
		tfMap["arn"] = aws.StringValue(v)
	}

	return tfMap
}

func expandLakeFormationDatabaseResource(tfMap map[string]interface{}) *lakeformation.DatabaseResource {
	if tfMap == nil {
		return nil
	}

	apiObject := &lakeformation.DatabaseResource{}

	if v, ok := tfMap["catalog_id"].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func flattenLakeFormationDatabaseResource(apiObject *lakeformation.DatabaseResource) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CatalogId; v != nil {
		tfMap["catalog_id"] = aws.StringValue(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	return tfMap
}

func expandLakeFormationTableResource(tfMap map[string]interface{}) *lakeformation.TableResource {
	if tfMap == nil {
		return nil
	}

	apiObject := &lakeformation.TableResource{}

	if v, ok := tfMap["catalog_id"].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}

	if v, ok := tfMap["database_name"].(string); ok && v != "" {
		apiObject.DatabaseName = aws.String(v)
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["wildcard"].(bool); ok && v {
		apiObject.TableWildcard = &lakeformation.TableWildcard{}
	}

	return apiObject
}

func flattenLakeFormationTableResource(apiObject *lakeformation.TableResource) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CatalogId; v != nil {
		tfMap["catalog_id"] = aws.StringValue(v)
	}

	if v := apiObject.DatabaseName; v != nil {
		tfMap["database_name"] = aws.StringValue(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	if v := apiObject.TableWildcard; v != nil {
		tfMap["wildcard"] = true
	}

	return tfMap
}

func expandLakeFormationTableWithColumnsResource(tfMap map[string]interface{}) *lakeformation.TableWithColumnsResource {
	if tfMap == nil {
		return nil
	}

	apiObject := &lakeformation.TableWithColumnsResource{}

	if v, ok := tfMap["catalog_id"].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}

	if v, ok := tfMap["column_names"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.ColumnNames = expandStringList(v.([]interface{}))
	}

	if v, ok := tfMap["database_name"].(string); ok && v != "" {
		apiObject.DatabaseName = aws.String(v)
	}

	if v, ok := tfMap["excluded_column_names"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.ColumnWildcard = &lakeformation.ColumnWildcard{
			ExcludedColumnNames: expandStringList(v.([]interface{})),
		}
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func flattenLakeFormationTableWithColumnsResource(apiObject *lakeformation.TableWithColumnsResource) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CatalogId; v != nil {
		tfMap["catalog_id"] = aws.StringValue(v)
	}

	tfMap["column_names"] = flattenStringList(apiObject.ColumnNames)

	if v := apiObject.DatabaseName; v != nil {
		tfMap["database_name"] = aws.StringValue(v)
	}

	if v := apiObject.ColumnWildcard; v != nil {
		tfMap["excluded_column_names"] = flattenStringList(v.ExcludedColumnNames)
	}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	return tfMap
}
