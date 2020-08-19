package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsQuickSightDataSource() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsQuickSightDataSourceCreate,
		Read:   resourceAwsQuickSightDataSourceRead,
		Update: resourceAwsQuickSightDataSourceUpdate,
		Delete: resourceAwsQuickSightDataSourceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"aws_account_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"credentials": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"credential_pair": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"password": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"username": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
								},
							},
						},
					},
				},
			},

			"data_source_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"parameters": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"amazon_elasticsearch": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"domain": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
								},
							},
						},
						"athena": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"work_group": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.NoZeroValues,
									},
								},
							},
						},
						"aurora": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"aurora_postgresql": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"aws_iot_analytics": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"data_set_name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
								},
							},
						},
						"jira": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"site_base_url": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
								},
							},
						},
						"maria_db": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"mysql": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"postgresql": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"presto": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"catalog": {
										Type:     schema.TypeString,
										Required: true,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						// The documentation is not clear on how to pass RDS parameters...
						"redshift": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cluster_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"port": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						"s3": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"manifest_file_location": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"bucket": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.NoZeroValues,
												},
												"key": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.NoZeroValues,
												},
											},
										},
									},
								},
							},
						},
						"service_now": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"site_base_url": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
								},
							},
						},
						"snowflake": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"warehouse": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"spark": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"host": {
										Type:     schema.TypeString,
										Required: true,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"sql_server": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"teradata": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"twitter": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_rows": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
									"query": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
								},
							},
						},
					},
				},
			},

			"permission": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"actions": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"principal": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.NoZeroValues,
						},
					},
				},
			},

			"ssl_properties": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disable_ssl": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},

			"tags": tagsSchema(),

			// This will be inferred from the passed `parameters` value
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"vpc_connection_properties": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vpc_connection_arn": {
							Type:         schema.TypeBool,
							Optional:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
		},
	}
}

func resourceAwsQuickSightDataSourceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).quicksightconn

	awsAccountId := meta.(*AWSClient).accountid
	id := d.Get("data_source_id").(string)

	if v, ok := d.GetOk("aws_account_id"); ok {
		awsAccountId = v.(string)
	}

	params := &quicksight.CreateDataSourceInput{
		AwsAccountId: aws.String(awsAccountId),
		DataSourceId: aws.String(id),
		Name:         aws.String(d.Get("name").(string)),
	}

	if credentials := resourceAwsQuickSightDataSourceCredentials(d); credentials != nil {
		params.Credentials = credentials
	}

	if dataSourceType, dataSourceParameters := resourceAwsQuickSightDataSourceParameters(d); dataSourceParameters != nil {
		params.Type = dataSourceType
		params.DataSourceParameters = dataSourceParameters
	}

	if v := d.Get("permission"); v != nil && len(v.([]interface{})) != 0 {
		params.Permissions = make([]*quicksight.ResourcePermission, 0)

		for _, v := range v.([]interface{}) {
			permissionResource := v.(map[string]interface{})
			permission := &quicksight.ResourcePermission{
				Actions:   expandStringSet(permissionResource["actions"].(*schema.Set)),
				Principal: aws.String(permissionResource["principal"].(string)),
			}

			params.Permissions = append(params.Permissions, permission)
		}
	}

	if sslProperties := resourceAwsQuickSightDataSourceSslProperties(d); sslProperties != nil {
		params.SslProperties = sslProperties
	}

	if v, ok := d.GetOk("tags"); ok {
		params.Tags = tagsFromMapQuickSight(v.(map[string]interface{}))
	}

	if vpcConnectionProperties := resourceAwsQuickSightDataSourceVpcConnectionProperties(d); vpcConnectionProperties != nil {
		params.VpcConnectionProperties = vpcConnectionProperties
	}

	_, err := conn.CreateDataSource(params)
	if err != nil {
		return fmt.Errorf("Error creating QuickSight Data Source: %s", err)
	}

	d.SetId(fmt.Sprintf("%s/%s", awsAccountId, id))

	return resourceAwsQuickSightDataSourceRead(d, meta)
}

func resourceAwsQuickSightDataSourceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).quicksightconn

	awsAccountId, dataSourceId, err := resourceAwsQuickSightDataSourceParseID(d.Id())
	if err != nil {
		return err
	}

	descOpts := &quicksight.DescribeDataSourceInput{
		AwsAccountId: aws.String(awsAccountId),
		DataSourceId: aws.String(dataSourceId),
	}

	var dataSourceResp *quicksight.DescribeDataSourceOutput
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		var err error
		dataSourceResp, err = conn.DescribeDataSource(descOpts)

		if dataSourceResp != nil && dataSourceResp.DataSource != nil {
			status := aws.StringValue(dataSourceResp.DataSource.Status)

			if status == quicksight.ResourceStatusCreationInProgress || status == quicksight.ResourceStatusUpdateInProgress {
				return resource.RetryableError(fmt.Errorf("Data Source operation still in progress (%s): %s", d.Id(), status))
			}
			if status == quicksight.ResourceStatusCreationFailed || status == quicksight.ResourceStatusUpdateFailed {
				return resource.NonRetryableError(fmt.Errorf("Data Source operation failed (%s): %s", d.Id(), status))
			}
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isAWSErr(err, quicksight.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] QuickSight Data Source %s is already gone", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error describing QuickSight Data Source (%s): %s", d.Id(), err)
	}

	permsResp, err := conn.DescribeDataSourcePermissions(&quicksight.DescribeDataSourcePermissionsInput{
		AwsAccountId: aws.String(awsAccountId),
		DataSourceId: aws.String(dataSourceId),
	})

	if err != nil {
		return fmt.Errorf("Error describing QuickSight Data Source permissions (%s): %s", d.Id(), err)
	}

	dataSource := dataSourceResp.DataSource

	d.Set("arn", dataSource.Arn)
	d.Set("name", dataSource.Name)
	d.Set("data_source_id", dataSource.DataSourceId)
	d.Set("aws_account_id", awsAccountId)

	if err := d.Set("permission", flattenQuickSightPermissions(permsResp.Permissions)); err != nil {
		return fmt.Errorf("Error setting permission error: %#v", err)
	}

	params := map[string]interface{}{}

	if dataSource.DataSourceParameters.AmazonElasticsearchParameters != nil {
		params = map[string]interface{}{
			"amazon_elasticsearch": []interface{}{
				map[string]interface{}{
					"domain": dataSource.DataSourceParameters.AmazonElasticsearchParameters.Domain,
				},
			},
		}
	}

	if dataSource.DataSourceParameters.AthenaParameters != nil {
		params = map[string]interface{}{
			"athena": []interface{}{
				map[string]interface{}{
					"work_group": dataSource.DataSourceParameters.AthenaParameters.WorkGroup,
				},
			},
		}
	}

	if dataSource.DataSourceParameters.AuroraParameters != nil {
		params = map[string]interface{}{
			"aurora": []interface{}{
				map[string]interface{}{
					"database": dataSource.DataSourceParameters.AuroraParameters.Database,
					"host":     dataSource.DataSourceParameters.AuroraParameters.Host,
					"port":     dataSource.DataSourceParameters.AuroraParameters.Port,
				},
			},
		}
	}

	if dataSource.DataSourceParameters.AuroraPostgreSqlParameters != nil {
		params = map[string]interface{}{
			"aurora_postgresql": []interface{}{
				map[string]interface{}{
					"database": dataSource.DataSourceParameters.AuroraPostgreSqlParameters.Database,
					"host":     dataSource.DataSourceParameters.AuroraPostgreSqlParameters.Host,
					"port":     dataSource.DataSourceParameters.AuroraPostgreSqlParameters.Port,
				},
			},
		}
	}

	if dataSource.DataSourceParameters.AwsIotAnalyticsParameters != nil {
		params = map[string]interface{}{
			"aws_iot_analytics": []interface{}{
				map[string]interface{}{
					"data_set_name": dataSource.DataSourceParameters.AwsIotAnalyticsParameters.DataSetName,
				},
			},
		}
	}

	if dataSource.DataSourceParameters.JiraParameters != nil {
		params = map[string]interface{}{
			"jira": []interface{}{
				map[string]interface{}{
					"site_base_url": dataSource.DataSourceParameters.JiraParameters.SiteBaseUrl,
				},
			},
		}
	}

	if dataSource.DataSourceParameters.MariaDbParameters != nil {
		params = map[string]interface{}{
			"maria_db": []interface{}{
				map[string]interface{}{
					"database": dataSource.DataSourceParameters.MariaDbParameters.Database,
					"host":     dataSource.DataSourceParameters.MariaDbParameters.Host,
					"port":     dataSource.DataSourceParameters.MariaDbParameters.Port,
				},
			},
		}
	}

	if dataSource.DataSourceParameters.MySqlParameters != nil {
		params = map[string]interface{}{
			"mysql": []interface{}{
				map[string]interface{}{
					"database": dataSource.DataSourceParameters.MySqlParameters.Database,
					"host":     dataSource.DataSourceParameters.MySqlParameters.Host,
					"port":     dataSource.DataSourceParameters.MySqlParameters.Port,
				},
			},
		}
	}

	if dataSource.DataSourceParameters.PostgreSqlParameters != nil {
		params = map[string]interface{}{
			"postgresql": []interface{}{
				map[string]interface{}{
					"database": dataSource.DataSourceParameters.PostgreSqlParameters.Database,
					"host":     dataSource.DataSourceParameters.PostgreSqlParameters.Host,
					"port":     dataSource.DataSourceParameters.PostgreSqlParameters.Port,
				},
			},
		}
	}

	if dataSource.DataSourceParameters.PrestoParameters != nil {
		params = map[string]interface{}{
			"presto": []interface{}{
				map[string]interface{}{
					"catalog": dataSource.DataSourceParameters.PrestoParameters.Catalog,
					"host":    dataSource.DataSourceParameters.PrestoParameters.Host,
					"port":    dataSource.DataSourceParameters.PrestoParameters.Port,
				},
			},
		}
	}

	if dataSource.DataSourceParameters.RedshiftParameters != nil {
		params = map[string]interface{}{
			"redshift": []interface{}{
				map[string]interface{}{
					"cluster_id": dataSource.DataSourceParameters.RedshiftParameters.ClusterId,
					"database":   dataSource.DataSourceParameters.RedshiftParameters.Database,
					"host":       dataSource.DataSourceParameters.RedshiftParameters.Host,
					"port":       dataSource.DataSourceParameters.RedshiftParameters.Port,
				},
			},
		}
	}

	if dataSource.DataSourceParameters.S3Parameters != nil {
		params = map[string]interface{}{
			"s3": []interface{}{
				map[string]interface{}{
					"manifest_file_location": []interface{}{
						map[string]interface{}{
							"bucket": dataSource.DataSourceParameters.S3Parameters.ManifestFileLocation.Bucket,
							"key":    dataSource.DataSourceParameters.S3Parameters.ManifestFileLocation.Key,
						},
					},
				},
			},
		}
	}

	if dataSource.DataSourceParameters.ServiceNowParameters != nil {
		params = map[string]interface{}{
			"service_now": []interface{}{
				map[string]interface{}{
					"site_base_url": dataSource.DataSourceParameters.ServiceNowParameters.SiteBaseUrl,
				},
			},
		}
	}

	if dataSource.DataSourceParameters.SnowflakeParameters != nil {
		params = map[string]interface{}{
			"snowflake": []interface{}{
				map[string]interface{}{
					"database":  dataSource.DataSourceParameters.SnowflakeParameters.Database,
					"host":      dataSource.DataSourceParameters.SnowflakeParameters.Host,
					"warehouse": dataSource.DataSourceParameters.SnowflakeParameters.Warehouse,
				},
			},
		}
	}

	if dataSource.DataSourceParameters.SparkParameters != nil {
		params = map[string]interface{}{
			"spark": []interface{}{
				map[string]interface{}{
					"host": dataSource.DataSourceParameters.SparkParameters.Host,
					"port": dataSource.DataSourceParameters.SparkParameters.Port,
				},
			},
		}
	}

	if dataSource.DataSourceParameters.SqlServerParameters != nil {
		params = map[string]interface{}{
			"sql_server": []interface{}{
				map[string]interface{}{
					"database": dataSource.DataSourceParameters.SqlServerParameters.Database,
					"host":     dataSource.DataSourceParameters.SqlServerParameters.Host,
					"port":     dataSource.DataSourceParameters.SqlServerParameters.Port,
				},
			},
		}
	}

	if dataSource.DataSourceParameters.TeradataParameters != nil {
		params = map[string]interface{}{
			"teradata": []interface{}{
				map[string]interface{}{
					"database": dataSource.DataSourceParameters.TeradataParameters.Database,
					"host":     dataSource.DataSourceParameters.TeradataParameters.Host,
					"port":     dataSource.DataSourceParameters.TeradataParameters.Port,
				},
			},
		}
	}

	if dataSource.DataSourceParameters.TwitterParameters != nil {
		params = map[string]interface{}{
			"twitter": []interface{}{
				map[string]interface{}{
					"max_rows": dataSource.DataSourceParameters.TwitterParameters.MaxRows,
					"query":    dataSource.DataSourceParameters.TwitterParameters.Query,
				},
			},
		}
	}

	d.Set("parameters", []interface{}{params})

	d.Set("type", inferQuickSightDataSourceTypeFromKey(params))

	return nil
}

func resourceAwsQuickSightDataSourceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).quicksightconn

	awsAccountId, dataSourceId, err := resourceAwsQuickSightDataSourceParseID(d.Id())
	if err != nil {
		return err
	}

	params := &quicksight.UpdateDataSourceInput{
		AwsAccountId: aws.String(awsAccountId),
		DataSourceId: aws.String(dataSourceId),
	}

	if credentials := resourceAwsQuickSightDataSourceCredentials(d); credentials != nil {
		params.Credentials = credentials
	}

	if dataSourceType, dataSourceParameters := resourceAwsQuickSightDataSourceParameters(d); dataSourceParameters != nil {
		params.DataSourceParameters = dataSourceParameters
		d.Set("type", dataSourceType)
	}

	if d.HasChange("permission") {
		oraw, nraw := d.GetChange("permission")
		o := oraw.([]interface{})
		n := nraw.([]interface{})
		toGrant, toRevoke := diffQuickSightPermissionsToGrantAndRevoke(o, n)

		if len(toGrant) > 0 || len(toRevoke) > 0 {
			params := &quicksight.UpdateDataSourcePermissionsInput{
				AwsAccountId:      aws.String(awsAccountId),
				DataSourceId:      aws.String(dataSourceId),
				GrantPermissions:  toGrant,
				RevokePermissions: toRevoke,
			}

			_, err := conn.UpdateDataSourcePermissions(params)
			if err != nil {
				return fmt.Errorf("Error updating QuickSight Data Source permissions: %s", err)
			}
		}
	}

	if sslProperties := resourceAwsQuickSightDataSourceSslProperties(d); sslProperties != nil {
		params.SslProperties = sslProperties
	}

	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})
		c, r := diffTagsQuickSight(tagsFromMapQuickSight(o), tagsFromMapQuickSight(n))

		if len(r) > 0 {
			_, err := conn.UntagResource(&quicksight.UntagResourceInput{
				ResourceArn: aws.String(quicksightDataSourceArn(meta.(*AWSClient).region, awsAccountId, dataSourceId)),
				TagKeys:     tagKeysQuickSight(r),
			})
			if err != nil {
				return fmt.Errorf("Error deleting QuickSight Data Source tags: %s", err)
			}
		}

		if len(c) > 0 {
			_, err := conn.TagResource(&quicksight.TagResourceInput{
				ResourceArn: aws.String(quicksightDataSourceArn(meta.(*AWSClient).region, awsAccountId, dataSourceId)),
				Tags:        c,
			})
			if err != nil {
				return fmt.Errorf("Error updating QuickSight Data Source tags: %s", err)
			}
		}
	}

	if vpcConnectionProperties := resourceAwsQuickSightDataSourceVpcConnectionProperties(d); vpcConnectionProperties != nil {
		params.VpcConnectionProperties = vpcConnectionProperties
	}

	_, err = conn.UpdateDataSource(params)
	if isAWSErr(err, quicksight.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] QuickSight Data Source %s is already gone", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error updating QuickSight Data Source %s: %s", d.Id(), err)
	}

	return resourceAwsQuickSightDataSourceRead(d, meta)
}

func resourceAwsQuickSightDataSourceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).quicksightconn

	awsAccountId, dataSourceId, err := resourceAwsQuickSightDataSourceParseID(d.Id())
	if err != nil {
		return err
	}

	deleteOpts := &quicksight.DeleteDataSourceInput{
		AwsAccountId: aws.String(awsAccountId),
		DataSourceId: aws.String(dataSourceId),
	}

	if _, err := conn.DeleteDataSource(deleteOpts); err != nil {
		if isAWSErr(err, quicksight.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting QuickSight Data Source %s: %s", d.Id(), err)
	}

	return nil
}

func resourceAwsQuickSightDataSourceCredentials(d *schema.ResourceData) *quicksight.DataSourceCredentials {
	if v := d.Get("credentials"); v != nil {
		for _, v := range v.([]interface{}) {
			credentials := v.(map[string]interface{})

			if v := credentials["credential_pair"]; v != nil && v.([]interface{}) != nil {
				for _, v := range v.([]interface{}) {
					credentialPairResource := v.(map[string]interface{})
					credentialPair := &quicksight.CredentialPair{}

					if v, ok := credentialPairResource["username"]; ok && v.(string) != "" {
						credentialPair.Username = aws.String(v.(string))
					}

					if v, ok := credentialPairResource["password"]; ok && v.(string) != "" {
						credentialPair.Password = aws.String(v.(string))
					}

					return &quicksight.DataSourceCredentials{
						CredentialPair: credentialPair,
					}
				}
			}
		}
	}

	return nil
}

var quickSightDataSourceParamToDataType = map[string]string{
	"amazon_elasticsearch": quicksight.DataSourceTypeAmazonElasticsearch,
	"athena":               quicksight.DataSourceTypeAthena,
	"aurora":               quicksight.DataSourceTypeAurora,
	"aurora_postgresql":    quicksight.DataSourceTypeAuroraPostgresql,
	"aws_iot_analytics":    quicksight.DataSourceTypeAwsIotAnalytics,
	"jira":                 quicksight.DataSourceTypeJira,
	"maria_db":             quicksight.DataSourceTypeMariadb,
	"mysql":                quicksight.DataSourceTypeMysql,
	"postgresql":           quicksight.DataSourceTypePostgresql,
	"presto":               quicksight.DataSourceTypePresto,
	"redshift":             quicksight.DataSourceTypeRedshift,
	"s3":                   quicksight.DataSourceTypeS3,
	"service_now":          quicksight.DataSourceTypeServicenow,
	"snowflake":            quicksight.DataSourceTypeSnowflake,
	"spark":                quicksight.DataSourceTypeSpark,
	"sql_server":           quicksight.DataSourceTypeSqlserver,
	"teradata":             quicksight.DataSourceTypeTeradata,
	"twitter":              quicksight.DataSourceTypeTwitter,
}

func inferQuickSightDataSourceTypeFromKey(params map[string]interface{}) string {
	if len(params) == 1 {
		for k := range params {
			if dataSourceType, found := quickSightDataSourceParamToDataType[k]; found {
				return dataSourceType
			}
		}
	}

	for k, v := range params {
		if dataSourceType, found := quickSightDataSourceParamToDataType[k]; found && v.([]interface{}) != nil && len(v.([]interface{})) > 0 {
			return dataSourceType
		}
	}

	return "UNKNOWN"
}

func resourceAwsQuickSightDataSourceParameters(d *schema.ResourceData) (*string, *quicksight.DataSourceParameters) {
	if v := d.Get("parameters"); v != nil {
		dataSourceParamsResource := &quicksight.DataSourceParameters{}
		var dataSourceType string
		for _, v := range v.([]interface{}) {
			dataSourceParams := v.(map[string]interface{})
			dataSourceType = inferQuickSightDataSourceTypeFromKey(dataSourceParams)

			if v := dataSourceParams["amazon_elasticsearch"]; v != nil && v.([]interface{}) != nil {
				for _, v := range v.([]interface{}) {
					psResource := v.(map[string]interface{})
					dataSourceParamsResource.AmazonElasticsearchParameters = &quicksight.AmazonElasticsearchParameters{
						Domain: aws.String(psResource["domain"].(string)),
					}
				}
			}

			if v := dataSourceParams["athena"]; v != nil && v.([]interface{}) != nil {
				for _, v := range v.([]interface{}) {
					psResource := v.(map[string]interface{})
					ps := &quicksight.AthenaParameters{}

					if v, ok := psResource["work_group"]; ok && v.(string) != "" {
						ps.WorkGroup = aws.String(v.(string))
					}

					dataSourceParamsResource.AthenaParameters = ps
				}
			}

			if v := dataSourceParams["aurora"]; v != nil && v.([]interface{}) != nil {
				for _, v := range v.([]interface{}) {
					psResource := v.(map[string]interface{})
					dataSourceParamsResource.AuroraParameters = &quicksight.AuroraParameters{
						Database: aws.String(psResource["database"].(string)),
						Host:     aws.String(psResource["host"].(string)),
						Port:     aws.Int64(psResource["port"].(int64)),
					}
				}
			}

			if v := dataSourceParams["aurora_postgresql"]; v != nil && v.([]interface{}) != nil {
				for _, v := range v.([]interface{}) {
					psResource := v.(map[string]interface{})
					dataSourceParamsResource.AuroraPostgreSqlParameters = &quicksight.AuroraPostgreSqlParameters{
						Database: aws.String(psResource["database"].(string)),
						Host:     aws.String(psResource["host"].(string)),
						Port:     aws.Int64(psResource["port"].(int64)),
					}
				}
			}

			if v := dataSourceParams["aws_iot_analytics"]; v != nil && v.([]interface{}) != nil {
				for _, v := range v.([]interface{}) {
					psResource := v.(map[string]interface{})
					dataSourceParamsResource.AwsIotAnalyticsParameters = &quicksight.AwsIotAnalyticsParameters{
						DataSetName: aws.String(psResource["data_set_name"].(string)),
					}
				}
			}

			if v := dataSourceParams["jira"]; v != nil && v.([]interface{}) != nil {
				for _, v := range v.([]interface{}) {
					psResource := v.(map[string]interface{})
					dataSourceParamsResource.JiraParameters = &quicksight.JiraParameters{
						SiteBaseUrl: aws.String(psResource["site_base_url"].(string)),
					}
				}
			}

			if v := dataSourceParams["maria_db"]; v != nil && v.([]interface{}) != nil {
				for _, v := range v.([]interface{}) {
					psResource := v.(map[string]interface{})
					dataSourceParamsResource.MariaDbParameters = &quicksight.MariaDbParameters{
						Database: aws.String(psResource["database"].(string)),
						Host:     aws.String(psResource["host"].(string)),
						Port:     aws.Int64(psResource["port"].(int64)),
					}
				}
			}

			if v := dataSourceParams["mysql"]; v != nil && v.([]interface{}) != nil {
				for _, v := range v.([]interface{}) {
					psResource := v.(map[string]interface{})
					dataSourceParamsResource.MySqlParameters = &quicksight.MySqlParameters{
						Database: aws.String(psResource["database"].(string)),
						Host:     aws.String(psResource["host"].(string)),
						Port:     aws.Int64(psResource["port"].(int64)),
					}
				}
			}

			if v := dataSourceParams["postgresql"]; v != nil && v.([]interface{}) != nil {
				for _, v := range v.([]interface{}) {
					psResource := v.(map[string]interface{})
					dataSourceParamsResource.PostgreSqlParameters = &quicksight.PostgreSqlParameters{
						Database: aws.String(psResource["database"].(string)),
						Host:     aws.String(psResource["host"].(string)),
						Port:     aws.Int64(psResource["port"].(int64)),
					}
				}
			}

			if v := dataSourceParams["presto"]; v != nil && v.([]interface{}) != nil {
				for _, v := range v.([]interface{}) {
					psResource := v.(map[string]interface{})
					dataSourceParamsResource.PrestoParameters = &quicksight.PrestoParameters{
						Catalog: aws.String(psResource["catalog"].(string)),
						Host:    aws.String(psResource["host"].(string)),
						Port:    aws.Int64(psResource["port"].(int64)),
					}
				}
			}

			if v := dataSourceParams["redshift"]; v != nil && v.([]interface{}) != nil {
				for _, v := range v.([]interface{}) {
					psResource := v.(map[string]interface{})
					ps := &quicksight.RedshiftParameters{
						Database: aws.String(psResource["database"].(string)),
					}

					if v, ok := psResource["cluster_id"]; ok && v.(string) != "" {
						ps.ClusterId = aws.String(v.(string))
					}

					if v, ok := psResource["host"]; ok && v.(string) != "" {
						ps.Host = aws.String(v.(string))
					}

					if v, ok := psResource["port"]; ok && v.(int64) != 0 {
						ps.Port = aws.Int64(v.(int64))
					}

					dataSourceParamsResource.RedshiftParameters = ps
				}
			}

			if v := dataSourceParams["s3"]; v != nil && v.([]interface{}) != nil {
				for _, v := range v.([]interface{}) {
					s3 := v.(map[string]interface{})
					if v := s3["manifest_file_location"]; v != nil && v.([]interface{}) != nil {
						for _, v := range v.([]interface{}) {
							psResource := v.(map[string]interface{})
							dataSourceParamsResource.S3Parameters = &quicksight.S3Parameters{
								ManifestFileLocation: &quicksight.ManifestFileLocation{
									Bucket: aws.String(psResource["bucket"].(string)),
									Key:    aws.String(psResource["key"].(string)),
								},
							}
						}
					}
				}
			}

			if v := dataSourceParams["service_now"]; v != nil && v.([]interface{}) != nil {
				for _, v := range v.([]interface{}) {
					psResource := v.(map[string]interface{})
					dataSourceParamsResource.ServiceNowParameters = &quicksight.ServiceNowParameters{
						SiteBaseUrl: aws.String(psResource["site_base_url"].(string)),
					}
				}
			}

			if v := dataSourceParams["snowflake"]; v != nil && v.([]interface{}) != nil {
				for _, v := range v.([]interface{}) {
					psResource := v.(map[string]interface{})
					dataSourceParamsResource.SnowflakeParameters = &quicksight.SnowflakeParameters{
						Database:  aws.String(psResource["database"].(string)),
						Host:      aws.String(psResource["host"].(string)),
						Warehouse: aws.String(psResource["warehouse"].(string)),
					}
				}
			}

			if v := dataSourceParams["spark"]; v != nil && v.([]interface{}) != nil {
				for _, v := range v.([]interface{}) {
					psResource := v.(map[string]interface{})
					dataSourceParamsResource.SparkParameters = &quicksight.SparkParameters{
						Host: aws.String(psResource["host"].(string)),
						Port: aws.Int64(psResource["port"].(int64)),
					}
				}
			}

			if v := dataSourceParams["sql_server"]; v != nil && v.([]interface{}) != nil {
				for _, v := range v.([]interface{}) {
					psResource := v.(map[string]interface{})
					dataSourceParamsResource.SqlServerParameters = &quicksight.SqlServerParameters{
						Database: aws.String(psResource["database"].(string)),
						Host:     aws.String(psResource["host"].(string)),
						Port:     aws.Int64(psResource["port"].(int64)),
					}
				}
			}

			if v := dataSourceParams["teradata"]; v != nil && v.([]interface{}) != nil {
				for _, v := range v.([]interface{}) {
					psResource := v.(map[string]interface{})
					dataSourceParamsResource.TeradataParameters = &quicksight.TeradataParameters{
						Database: aws.String(psResource["database"].(string)),
						Host:     aws.String(psResource["host"].(string)),
						Port:     aws.Int64(psResource["port"].(int64)),
					}
				}
			}

			if v := dataSourceParams["twitter"]; v != nil && v.([]interface{}) != nil {
				for _, v := range v.([]interface{}) {
					psResource := v.(map[string]interface{})
					dataSourceParamsResource.TwitterParameters = &quicksight.TwitterParameters{
						MaxRows: aws.Int64(psResource["max_rows"].(int64)),
						Query:   aws.String(psResource["query"].(string)),
					}
				}
			}

		}
		return aws.String(dataSourceType), dataSourceParamsResource
	}

	return aws.String(""), nil
}

func resourceAwsQuickSightDataSourceSslProperties(d *schema.ResourceData) *quicksight.SslProperties {
	if v := d.Get("ssl_properties"); v != nil {
		for _, v := range v.([]interface{}) {
			sslProperties := v.(map[string]interface{})

			if v, present := sslProperties["disable_ssl"]; present {
				return &quicksight.SslProperties{
					DisableSsl: aws.Bool(v.(bool)),
				}
			}
		}
	}

	return nil
}

func resourceAwsQuickSightDataSourceVpcConnectionProperties(d *schema.ResourceData) *quicksight.VpcConnectionProperties {
	if v := d.Get("vpc_connection_properties"); v != nil {
		for _, v := range v.([]interface{}) {
			vpcConnectionProperties := v.(map[string]interface{})

			if v := vpcConnectionProperties["vpc_connection_arn"]; v != nil && v.(string) != "" {
				return &quicksight.VpcConnectionProperties{
					VpcConnectionArn: aws.String(v.(string)),
				}
			}
		}
	}

	return nil
}

func resourceAwsQuickSightDataSourceParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID/DATA_SOURCE_ID", id)
	}
	return parts[0], parts[1], nil
}

func quicksightDataSourceArn(awsRegion string, awsAccountId string, dataSourceId string) string {
	return fmt.Sprintf("arn:aws:quicksight:%s:%s:datasource/%s", awsRegion, awsAccountId, dataSourceId)
}
