package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/glue/finder"
)

func resourceAwsGlueCatalogTable() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGlueCatalogTableCreate,
		Read:   resourceAwsGlueCatalogTableRead,
		Update: resourceAwsGlueCatalogTableUpdate,
		Delete: resourceAwsGlueCatalogTableDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"catalog_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"database_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringDoesNotMatch(regexp.MustCompile(`[A-Z]`), "uppercase characters cannot be used"),
				),
			},
			"owner": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"partition_keys": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"comment": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 131072),
						},
					},
				},
			},
			"retention": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"storage_descriptor": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket_columns": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 255),
							},
						},
						"columns": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"comment": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 255),
									},
									"name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
									"parameters": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"type": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 131072),
									},
								},
							},
						},
						"compressed": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"input_format": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"location": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"number_of_buckets": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"output_format": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"parameters": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ser_de_info": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
									"parameters": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"serialization_library": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
								},
							},
						},
						"schema_reference": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"schema_id": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"schema_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateArn,
													ExactlyOneOf: []string{"storage_descriptor.0.schema_reference.0.schema_id.0.schema_arn", "storage_descriptor.0.schema_reference.0.schema_id.0.schema_name"},
												},
												"schema_name": {
													Type:         schema.TypeString,
													Optional:     true,
													ExactlyOneOf: []string{"storage_descriptor.0.schema_reference.0.schema_id.0.schema_arn", "storage_descriptor.0.schema_reference.0.schema_id.0.schema_name"},
												},
												"registry_name": {
													Type:          schema.TypeString,
													Optional:      true,
													ConflictsWith: []string{"storage_descriptor.0.schema_reference.0.schema_id.0.schema_arn"},
												},
											},
										},
									},
									"schema_version_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ExactlyOneOf: []string{"storage_descriptor.0.schema_reference.0.schema_version_id", "storage_descriptor.0.schema_reference.0.schema_id"},
									},
									"schema_version_number": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(1, 100000),
									},
								},
							},
						},
						"skewed_info": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"skewed_column_names": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 255),
										},
									},
									"skewed_column_values": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"skewed_column_value_location_maps": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"sort_columns": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"column": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
									"sort_order": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntInSlice([]int{0, 1}),
									},
								},
							},
						},
						"stored_as_sub_directories": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"table_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"view_original_text": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 409600),
			},
			"view_expanded_text": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 409600),
			},
			"partition_index": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 3,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"index_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"keys": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"index_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func readAwsGlueTableID(id string) (catalogID string, dbName string, name string, error error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 3 {
		return "", "", "", fmt.Errorf("expected ID in format catalog-id:database-name:table-name, received: %s", id)
	}
	return idParts[0], idParts[1], idParts[2], nil
}

func resourceAwsGlueCatalogTableCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn
	catalogID := createAwsGlueCatalogID(d, meta.(*AWSClient).accountid)
	dbName := d.Get("database_name").(string)
	name := d.Get("name").(string)

	input := &glue.CreateTableInput{
		CatalogId:        aws.String(catalogID),
		DatabaseName:     aws.String(dbName),
		TableInput:       expandGlueTableInput(d),
		PartitionIndexes: expandGlueTablePartitionIndexes(d.Get("partition_index").([]interface{})),
	}

	log.Printf("[DEBUG] Glue catalog table input: %#v", input)
	_, err := conn.CreateTable(input)
	if err != nil {
		return fmt.Errorf("Error creating Glue Catalog Table: %w", err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", catalogID, dbName, name))

	return resourceAwsGlueCatalogTableRead(d, meta)
}

func resourceAwsGlueCatalogTableRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	catalogID, dbName, name, err := readAwsGlueTableID(d.Id())
	if err != nil {
		return err
	}

	out, err := finder.TableByName(conn, catalogID, dbName, name)
	if err != nil {

		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			log.Printf("[WARN] Glue Catalog Table (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error reading Glue Catalog Table: %w", err)
	}

	table := out.Table
	tableArn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "glue",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("table/%s/%s", dbName, aws.StringValue(table.Name)),
	}.String()
	d.Set("arn", tableArn)

	d.Set("name", table.Name)
	d.Set("catalog_id", catalogID)
	d.Set("database_name", dbName)
	d.Set("description", table.Description)
	d.Set("owner", table.Owner)
	d.Set("retention", table.Retention)

	if err := d.Set("storage_descriptor", flattenGlueStorageDescriptor(table.StorageDescriptor)); err != nil {
		return fmt.Errorf("error setting storage_descriptor: %w", err)
	}

	if err := d.Set("partition_keys", flattenGlueColumns(table.PartitionKeys)); err != nil {
		return fmt.Errorf("error setting partition_keys: %w", err)
	}

	d.Set("view_original_text", table.ViewOriginalText)
	d.Set("view_expanded_text", table.ViewExpandedText)
	d.Set("table_type", table.TableType)

	if err := d.Set("parameters", aws.StringValueMap(table.Parameters)); err != nil {
		return fmt.Errorf("error setting parameters: %w", err)
	}

	partIndexInput := &glue.GetPartitionIndexesInput{
		CatalogId:    out.Table.CatalogId,
		TableName:    out.Table.Name,
		DatabaseName: out.Table.DatabaseName,
	}
	partOut, err := conn.GetPartitionIndexes(partIndexInput)
	if err != nil {
		return fmt.Errorf("error getting Glue Partition Indexes: %w", err)
	}

	if partOut != nil && len(partOut.PartitionIndexDescriptorList) > 0 {
		if err := d.Set("partition_index", flattenGluePartitionIndexes(partOut.PartitionIndexDescriptorList)); err != nil {
			return fmt.Errorf("error setting partition_index: %w", err)
		}
	}

	return nil
}

func resourceAwsGlueCatalogTableUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	catalogID, dbName, _, err := readAwsGlueTableID(d.Id())
	if err != nil {
		return err
	}

	updateTableInput := &glue.UpdateTableInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		TableInput:   expandGlueTableInput(d),
	}

	if _, err := conn.UpdateTable(updateTableInput); err != nil {
		return fmt.Errorf("Error updating Glue Catalog Table: %w", err)
	}

	return resourceAwsGlueCatalogTableRead(d, meta)
}

func resourceAwsGlueCatalogTableDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	catalogID, dbName, name, tableIdErr := readAwsGlueTableID(d.Id())
	if tableIdErr != nil {
		return tableIdErr
	}

	log.Printf("[DEBUG] Glue Catalog Table: %s:%s:%s", catalogID, dbName, name)
	_, err := conn.DeleteTable(&glue.DeleteTableInput{
		CatalogId:    aws.String(catalogID),
		Name:         aws.String(name),
		DatabaseName: aws.String(dbName),
	})
	if err != nil {
		return fmt.Errorf("Error deleting Glue Catalog Table: %w", err)
	}
	return nil
}

func expandGlueTableInput(d *schema.ResourceData) *glue.TableInput {
	tableInput := &glue.TableInput{
		Name: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		tableInput.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("owner"); ok {
		tableInput.Owner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("retention"); ok {
		tableInput.Retention = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("storage_descriptor"); ok {
		tableInput.StorageDescriptor = expandGlueStorageDescriptor(v.([]interface{}))
	}

	if v, ok := d.GetOk("partition_keys"); ok {
		tableInput.PartitionKeys = expandGlueColumns(v.([]interface{}))
	} else {
		tableInput.PartitionKeys = []*glue.Column{}
	}

	if v, ok := d.GetOk("view_original_text"); ok {
		tableInput.ViewOriginalText = aws.String(v.(string))
	}

	if v, ok := d.GetOk("view_expanded_text"); ok {
		tableInput.ViewExpandedText = aws.String(v.(string))
	}

	if v, ok := d.GetOk("table_type"); ok {
		tableInput.TableType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parameters"); ok {
		tableInput.Parameters = stringMapToPointers(v.(map[string]interface{}))
	}

	return tableInput
}

func expandGlueTablePartitionIndexes(a []interface{}) []*glue.PartitionIndex {
	partitionIndexes := make([]*glue.PartitionIndex, 0, len(a))

	for _, m := range a {
		partitionIndexes = append(partitionIndexes, expandGlueTablePartitionIndex(m.(map[string]interface{})))
	}

	return partitionIndexes
}

func expandGlueTablePartitionIndex(m map[string]interface{}) *glue.PartitionIndex {
	partitionIndex := &glue.PartitionIndex{
		IndexName: aws.String(m["index_name"].(string)),
		Keys:      expandStringSet(m["keys"].(*schema.Set)),
	}

	return partitionIndex
}

func expandGlueStorageDescriptor(l []interface{}) *glue.StorageDescriptor {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	s := l[0].(map[string]interface{})
	storageDescriptor := &glue.StorageDescriptor{}

	if v, ok := s["columns"]; ok {
		storageDescriptor.Columns = expandGlueColumns(v.([]interface{}))
	}

	if v, ok := s["location"]; ok {
		storageDescriptor.Location = aws.String(v.(string))
	}

	if v, ok := s["input_format"]; ok {
		storageDescriptor.InputFormat = aws.String(v.(string))
	}

	if v, ok := s["output_format"]; ok {
		storageDescriptor.OutputFormat = aws.String(v.(string))
	}

	if v, ok := s["compressed"]; ok {
		storageDescriptor.Compressed = aws.Bool(v.(bool))
	}

	if v, ok := s["number_of_buckets"]; ok {
		storageDescriptor.NumberOfBuckets = aws.Int64(int64(v.(int)))
	}

	if v, ok := s["ser_de_info"]; ok {
		storageDescriptor.SerdeInfo = expandGlueSerDeInfo(v.([]interface{}))
	}

	if v, ok := s["bucket_columns"]; ok {
		storageDescriptor.BucketColumns = expandStringList(v.([]interface{}))
	}

	if v, ok := s["sort_columns"]; ok {
		storageDescriptor.SortColumns = expandGlueSortColumns(v.([]interface{}))
	}

	if v, ok := s["skewed_info"]; ok {
		storageDescriptor.SkewedInfo = expandGlueSkewedInfo(v.([]interface{}))
	}

	if v, ok := s["parameters"]; ok {
		storageDescriptor.Parameters = stringMapToPointers(v.(map[string]interface{}))
	}

	if v, ok := s["stored_as_sub_directories"]; ok {
		storageDescriptor.StoredAsSubDirectories = aws.Bool(v.(bool))
	}

	if v, ok := s["schema_reference"]; ok {
		storageDescriptor.SchemaReference = expandGlueTableSchemaReference(v.([]interface{}))
	}

	return storageDescriptor
}

func expandGlueColumns(columns []interface{}) []*glue.Column {
	columnSlice := []*glue.Column{}
	for _, element := range columns {
		elementMap := element.(map[string]interface{})

		column := &glue.Column{
			Name: aws.String(elementMap["name"].(string)),
		}

		if v, ok := elementMap["comment"]; ok {
			column.Comment = aws.String(v.(string))
		}

		if v, ok := elementMap["type"]; ok {
			column.Type = aws.String(v.(string))
		}

		if v, ok := elementMap["parameters"]; ok {
			column.Parameters = stringMapToPointers(v.(map[string]interface{}))
		}

		columnSlice = append(columnSlice, column)
	}

	return columnSlice
}

func expandGlueSerDeInfo(l []interface{}) *glue.SerDeInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	s := l[0].(map[string]interface{})
	serDeInfo := &glue.SerDeInfo{}

	if v := s["name"]; len(v.(string)) > 0 {
		serDeInfo.Name = aws.String(v.(string))
	}

	if v := s["parameters"]; len(v.(map[string]interface{})) > 0 {
		serDeInfo.Parameters = stringMapToPointers(v.(map[string]interface{}))
	}

	if v := s["serialization_library"]; len(v.(string)) > 0 {
		serDeInfo.SerializationLibrary = aws.String(v.(string))
	}

	return serDeInfo
}

func expandGlueSortColumns(columns []interface{}) []*glue.Order {
	orderSlice := make([]*glue.Order, len(columns))

	for i, element := range columns {
		elementMap := element.(map[string]interface{})

		order := &glue.Order{
			Column: aws.String(elementMap["column"].(string)),
		}

		if v, ok := elementMap["sort_order"]; ok {
			order.SortOrder = aws.Int64(int64(v.(int)))
		}

		orderSlice[i] = order
	}

	return orderSlice
}

func expandGlueSkewedInfo(l []interface{}) *glue.SkewedInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	s := l[0].(map[string]interface{})
	skewedInfo := &glue.SkewedInfo{}

	if v, ok := s["skewed_column_names"]; ok {
		skewedInfo.SkewedColumnNames = expandStringList(v.([]interface{}))
	}

	if v, ok := s["skewed_column_value_location_maps"]; ok {
		skewedInfo.SkewedColumnValueLocationMaps = stringMapToPointers(v.(map[string]interface{}))
	}

	if v, ok := s["skewed_column_values"]; ok {
		skewedInfo.SkewedColumnValues = expandStringList(v.([]interface{}))
	}

	return skewedInfo
}

func expandGlueTableSchemaReference(l []interface{}) *glue.SchemaReference {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	s := l[0].(map[string]interface{})
	schemaRef := &glue.SchemaReference{}

	if v, ok := s["schema_version_id"].(string); ok && v != "" {
		schemaRef.SchemaVersionId = aws.String(v)
	}

	if v, ok := s["schema_id"]; ok {
		schemaRef.SchemaId = expandGlueTableSchemaReferenceSchemaID(v.([]interface{}))
	}

	if v, ok := s["schema_version_number"].(int); ok {
		schemaRef.SchemaVersionNumber = aws.Int64(int64(v))
	}

	return schemaRef
}

func expandGlueTableSchemaReferenceSchemaID(l []interface{}) *glue.SchemaId {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	s := l[0].(map[string]interface{})
	schemaID := &glue.SchemaId{}

	if v, ok := s["registry_name"].(string); ok && v != "" {
		schemaID.RegistryName = aws.String(v)
	}

	if v, ok := s["schema_name"].(string); ok && v != "" {
		schemaID.SchemaName = aws.String(v)
	}

	if v, ok := s["schema_arn"].(string); ok && v != "" {
		schemaID.SchemaArn = aws.String(v)
	}

	return schemaID
}

func flattenGlueStorageDescriptor(s *glue.StorageDescriptor) []map[string]interface{} {
	if s == nil {
		storageDescriptors := make([]map[string]interface{}, 0)
		return storageDescriptors
	}

	storageDescriptors := make([]map[string]interface{}, 1)

	storageDescriptor := make(map[string]interface{})

	storageDescriptor["columns"] = flattenGlueColumns(s.Columns)
	storageDescriptor["location"] = aws.StringValue(s.Location)
	storageDescriptor["input_format"] = aws.StringValue(s.InputFormat)
	storageDescriptor["output_format"] = aws.StringValue(s.OutputFormat)
	storageDescriptor["compressed"] = aws.BoolValue(s.Compressed)
	storageDescriptor["number_of_buckets"] = aws.Int64Value(s.NumberOfBuckets)
	storageDescriptor["ser_de_info"] = flattenGlueSerDeInfo(s.SerdeInfo)
	storageDescriptor["bucket_columns"] = flattenStringList(s.BucketColumns)
	storageDescriptor["sort_columns"] = flattenGlueOrders(s.SortColumns)
	storageDescriptor["parameters"] = aws.StringValueMap(s.Parameters)
	storageDescriptor["skewed_info"] = flattenGlueSkewedInfo(s.SkewedInfo)
	storageDescriptor["stored_as_sub_directories"] = aws.BoolValue(s.StoredAsSubDirectories)

	if s.SchemaReference != nil {
		storageDescriptor["schema_reference"] = flattenGlueTableSchemaReference(s.SchemaReference)
	}

	storageDescriptors[0] = storageDescriptor

	return storageDescriptors
}

func flattenGlueColumns(cs []*glue.Column) []map[string]interface{} {
	columnsSlice := make([]map[string]interface{}, len(cs))
	if len(cs) > 0 {
		for i, v := range cs {
			columnsSlice[i] = flattenGlueColumn(v)
		}
	}

	return columnsSlice
}

func flattenGlueColumn(c *glue.Column) map[string]interface{} {
	column := make(map[string]interface{})

	if c == nil {
		return column
	}

	if v := aws.StringValue(c.Name); v != "" {
		column["name"] = v
	}

	if v := aws.StringValue(c.Type); v != "" {
		column["type"] = v
	}

	if v := aws.StringValue(c.Comment); v != "" {
		column["comment"] = v
	}

	if v := c.Parameters; v != nil {
		column["parameters"] = aws.StringValueMap(v)
	}

	return column
}

func flattenGluePartitionIndexes(cs []*glue.PartitionIndexDescriptor) []map[string]interface{} {
	partitionIndexSlice := make([]map[string]interface{}, len(cs))
	if len(cs) > 0 {
		for i, v := range cs {
			partitionIndexSlice[i] = flattenGluePartitionIndex(v)
		}
	}

	return partitionIndexSlice
}

func flattenGluePartitionIndex(c *glue.PartitionIndexDescriptor) map[string]interface{} {
	partitionIndex := make(map[string]interface{})

	if c == nil {
		return partitionIndex
	}

	if v := aws.StringValue(c.IndexName); v != "" {
		partitionIndex["index_name"] = v
	}

	if v := aws.StringValue(c.IndexStatus); v != "" {
		partitionIndex["index_status"] = v
	}

	if c.Keys != nil {
		names := make([]*string, 0, len(c.Keys))
		for _, key := range c.Keys {
			names = append(names, key.Name)
		}
		partitionIndex["keys"] = flattenStringSet(names)
	}

	return partitionIndex
}

func flattenGlueSerDeInfo(s *glue.SerDeInfo) []map[string]interface{} {
	if s == nil {
		serDeInfos := make([]map[string]interface{}, 0)
		return serDeInfos
	}

	serDeInfos := make([]map[string]interface{}, 1)
	serDeInfo := make(map[string]interface{})

	if v := aws.StringValue(s.Name); v != "" {
		serDeInfo["name"] = v
	}

	serDeInfo["parameters"] = aws.StringValueMap(s.Parameters)

	if v := aws.StringValue(s.SerializationLibrary); v != "" {
		serDeInfo["serialization_library"] = v
	}

	serDeInfos[0] = serDeInfo
	return serDeInfos
}

func flattenGlueOrders(os []*glue.Order) []map[string]interface{} {
	orders := make([]map[string]interface{}, len(os))
	for i, v := range os {
		order := make(map[string]interface{})
		order["column"] = aws.StringValue(v.Column)
		order["sort_order"] = int(aws.Int64Value(v.SortOrder))
		orders[i] = order
	}

	return orders
}

func flattenGlueSkewedInfo(s *glue.SkewedInfo) []map[string]interface{} {
	if s == nil {
		skewedInfoSlice := make([]map[string]interface{}, 0)
		return skewedInfoSlice
	}

	skewedInfoSlice := make([]map[string]interface{}, 1)

	skewedInfo := make(map[string]interface{})
	skewedInfo["skewed_column_names"] = flattenStringList(s.SkewedColumnNames)
	skewedInfo["skewed_column_value_location_maps"] = aws.StringValueMap(s.SkewedColumnValueLocationMaps)
	skewedInfo["skewed_column_values"] = flattenStringList(s.SkewedColumnValues)
	skewedInfoSlice[0] = skewedInfo

	return skewedInfoSlice
}

func flattenGlueTableSchemaReference(s *glue.SchemaReference) []map[string]interface{} {
	if s == nil {
		schemaReferenceInfoSlice := make([]map[string]interface{}, 0)
		return schemaReferenceInfoSlice
	}

	schemaReferenceInfoSlice := make([]map[string]interface{}, 1)

	schemaReferenceInfo := make(map[string]interface{})

	if s.SchemaVersionId != nil {
		schemaReferenceInfo["schema_version_id"] = aws.StringValue(s.SchemaVersionId)
	}

	if s.SchemaVersionNumber != nil {
		schemaReferenceInfo["schema_version_number"] = aws.Int64Value(s.SchemaVersionNumber)
	}

	if s.SchemaId != nil {
		schemaReferenceInfo["schema_id"] = flattenGlueTableSchemaReferenceSchemaID(s.SchemaId)
	}

	schemaReferenceInfoSlice[0] = schemaReferenceInfo

	return schemaReferenceInfoSlice
}

func flattenGlueTableSchemaReferenceSchemaID(s *glue.SchemaId) []map[string]interface{} {
	if s == nil {
		schemaIDInfoSlice := make([]map[string]interface{}, 0)
		return schemaIDInfoSlice
	}

	schemaIDInfoSlice := make([]map[string]interface{}, 1)

	schemaIDInfo := make(map[string]interface{})

	if s.RegistryName != nil {
		schemaIDInfo["registry_name"] = aws.StringValue(s.RegistryName)
	}

	if s.SchemaArn != nil {
		schemaIDInfo["schema_arn"] = aws.StringValue(s.SchemaArn)
	}

	if s.SchemaName != nil {
		schemaIDInfo["schema_name"] = aws.StringValue(s.SchemaName)
	}

	schemaIDInfoSlice[0] = schemaIDInfo

	return schemaIDInfoSlice
}
