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

func dataSourceAwsLakeFormationPermissions() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLakeFormationPermissionsRead,

		Schema: map[string]*schema.Schema{
			"catalog_id": {
				Type:         schema.TypeString,
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
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"permissions_with_grant_option": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"principal": {
				Type:         schema.TypeString,
				Required:     true,
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

func dataSourceAwsLakeFormationPermissionsRead(d *schema.ResourceData, meta interface{}) error {
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
	matchResource := expandLakeFormationResource(d, false)

	log.Printf("[DEBUG] Reading Lake Formation permissions: %v", input)
	var principalResourcePermissions []*lakeformation.PrincipalResourcePermissions

	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		err := conn.ListPermissionsPages(input, func(resp *lakeformation.ListPermissionsOutput, lastPage bool) bool {
			for _, permission := range resp.PrincipalResourcePermissions {
				if permission == nil {
					continue
				}

				if resourceAwsLakeFormationPermissionsCompareResource(*matchResource, *permission.Resource) {
					principalResourcePermissions = append(principalResourcePermissions, permission)
				}
			}
			return !lastPage
		})

		if err != nil {
			if isAWSErr(err, lakeformation.ErrCodeInvalidInputException, "Invalid principal") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(fmt.Errorf("error reading Lake Formation Permissions: %w", err))
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		err = conn.ListPermissionsPages(input, func(resp *lakeformation.ListPermissionsOutput, lastPage bool) bool {
			for _, permission := range resp.PrincipalResourcePermissions {
				if permission == nil {
					continue
				}

				if resourceAwsLakeFormationPermissionsCompareResource(*matchResource, *permission.Resource) {
					principalResourcePermissions = append(principalResourcePermissions, permission)
				}
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

	d.SetId(fmt.Sprintf("%d", hashcode.String(input.String())))
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
