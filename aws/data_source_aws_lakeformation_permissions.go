package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
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
			"policy_tag": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				ExactlyOneOf: []string{
					"catalog_resource",
					"data_location",
					"database",
					"table",
					"table_with_columns",
					"policy_tag",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
						"values": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							MaxItems: 15,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validatePolicyTagValues(),
							},
							Set: schema.HashString,
						},
						"catalog_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
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
						"wildcard": {
							Type:     schema.TypeBool,
							Optional: true,
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
		Resource: &lakeformation.Resource{},
	}

	if v, ok := d.GetOk("catalog_id"); ok {
		input.CatalogId = aws.String(v.(string))
	}

	if _, ok := d.GetOk("catalog_resource"); ok {
		input.Resource.Catalog = expandLakeFormationCatalogResource()
	}

	if v, ok := d.GetOk("data_location"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Resource.DataLocation = expandLakeFormationDataLocationResource(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("database"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Resource.Database = expandLakeFormationDatabaseResource(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("policy_tag"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Resource.LFTag = expandLakeFormationLFTagKeyResource(v.([]interface{})[0].(map[string]interface{}))
	}

	tableType := ""

	if v, ok := d.GetOk("table"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Resource.Table = expandLakeFormationTableResource(v.([]interface{})[0].(map[string]interface{}))
		tableType = TableTypeTable
	}

	if v, ok := d.GetOk("table_with_columns"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		// can't ListPermissions for TableWithColumns, so use Table instead
		input.Resource.Table = expandLakeFormationTableWithColumnsResourceAsTable(v.([]interface{})[0].(map[string]interface{}))
		tableType = TableTypeTableWithColumns
	}

	log.Printf("[DEBUG] Reading Lake Formation permissions: %v", input)
	var allPermissions []*lakeformation.PrincipalResourcePermissions

	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		err := conn.ListPermissionsPages(input, func(resp *lakeformation.ListPermissionsOutput, lastPage bool) bool {
			for _, permission := range resp.PrincipalResourcePermissions {
				if permission == nil {
					continue
				}

				allPermissions = append(allPermissions, permission)
			}
			return !lastPage
		})

		if err != nil {
			if tfawserr.ErrMessageContains(err, lakeformation.ErrCodeInvalidInputException, "Invalid principal") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(fmt.Errorf("error reading Lake Formation Permissions: %w", err))
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		err = conn.ListPermissionsPages(input, func(resp *lakeformation.ListPermissionsOutput, lastPage bool) bool {
			for _, permission := range resp.PrincipalResourcePermissions {
				if permission == nil {
					continue
				}

				allPermissions = append(allPermissions, permission)
			}
			return !lastPage
		})
	}

	d.SetId(fmt.Sprintf("%d", hashcode.String(input.String())))

	if err != nil {
		return fmt.Errorf("error reading Lake Formation permissions: %w", err)
	}

	var cleanPermissions []*lakeformation.PrincipalResourcePermissions

	if input.Resource.Catalog != nil {
		cleanPermissions = filterLakeFormationCatalogPermissions(allPermissions)
	}

	if input.Resource.DataLocation != nil {
		cleanPermissions = filterLakeFormationDataLocationPermissions(allPermissions)
	}

	if input.Resource.Database != nil {
		cleanPermissions = filterLakeFormationDatabasePermissions(allPermissions)
	}

	if input.Resource.LFTag != nil {
		cleanPermissions = filterLakeFormationLFTagPermissions(allPermissions)
	}

	if tableType == TableTypeTable {
		cleanPermissions = filterLakeFormationTablePermissions(
			aws.StringValue(input.Resource.Table.Name),
			input.Resource.Table.TableWildcard != nil,
			allPermissions,
		)
	}

	if tableType == TableTypeTableWithColumns {
		cleanPermissions = filterLakeFormationTableWithColumnsPermissions(
			d.Get("table_with_columns.0.database_name").(string),
			d.Get("table_with_columns.0.wildcard").(bool),
			expandStringList(d.Get("table_with_columns.0.column_names").([]interface{})),
			expandStringList(d.Get("table_with_columns.0.excluded_column_names").([]interface{})),
			allPermissions,
		)
	}

	if len(cleanPermissions) != len(allPermissions) {
		log.Printf("[INFO] Resource Lake Formation clean permissions (%d) and all permissions (%d) have different lengths (this is not necessarily a problem): %s", len(cleanPermissions), len(allPermissions), d.Id())
	}

	d.Set("principal", cleanPermissions[0].Principal.DataLakePrincipalIdentifier)
	d.Set("permissions", flattenLakeFormationPermissions(cleanPermissions))
	d.Set("permissions_with_grant_option", flattenLakeFormationGrantPermissions(cleanPermissions))

	if cleanPermissions[0].Resource.Catalog != nil {
		d.Set("catalog_resource", true)
	}

	if cleanPermissions[0].Resource.DataLocation != nil {
		if err := d.Set("data_location", []interface{}{flattenLakeFormationDataLocationResource(cleanPermissions[0].Resource.DataLocation)}); err != nil {
			return fmt.Errorf("error setting data_location: %w", err)
		}
	} else {
		d.Set("data_location", nil)
	}

	if cleanPermissions[0].Resource.Database != nil {
		if err := d.Set("database", []interface{}{flattenLakeFormationDatabaseResource(cleanPermissions[0].Resource.Database)}); err != nil {
			return fmt.Errorf("error setting database: %w", err)
		}
	} else {
		d.Set("database", nil)
	}

	if cleanPermissions[0].Resource.LFTag != nil {
		if err := d.Set("policy_tag", []interface{}{flattenLakeFormationLFTagKeyResource(cleanPermissions[0].Resource.LFTag)}); err != nil {
			return fmt.Errorf("error setting policy tag: %w", err)
		}
	} else {
		d.Set("policy_tag", nil)
	}

	tableSet := false

	if v, ok := d.GetOk("table"); ok && len(v.([]interface{})) > 0 {
		// since perm list could include TableWithColumns, get the right one
		for _, perm := range cleanPermissions {
			if perm.Resource.Table != nil {
				if err := d.Set("table", []interface{}{flattenLakeFormationTableResource(perm.Resource.Table)}); err != nil {
					return fmt.Errorf("error setting table: %w", err)
				}
				tableSet = true
				break
			}
		}
	}

	if !tableSet {
		d.Set("table", nil)
	}

	twcSet := false

	if v, ok := d.GetOk("table_with_columns"); ok && len(v.([]interface{})) > 0 {
		// since perm list could include Table, get the right one
		for _, perm := range cleanPermissions {
			if perm.Resource.TableWithColumns != nil {
				if err := d.Set("table_with_columns", []interface{}{flattenLakeFormationTableWithColumnsResource(perm.Resource.TableWithColumns)}); err != nil {
					return fmt.Errorf("error setting table_with_columns: %w", err)
				}
				twcSet = true
				break
			}
		}
	}

	if !twcSet {
		d.Set("table_with_columns", nil)
	}

	return nil
}
