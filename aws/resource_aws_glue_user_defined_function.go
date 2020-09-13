package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsGlueUserDefinedFunction() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGlueUserDefinedFunctionCreate,
		Read:   resourceAwsGlueUserDefinedFunctionRead,
		Update: resourceAwsGlueUserDefinedFunctionUpdate,
		Delete: resourceAwsGlueUserDefinedFunctionDelete,
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
			},
			"database_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"class_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"owner_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"owner_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(glue.PrincipalType_Values(), false),
			},
			"resource_uris": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1000,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(glue.ResourceType_Values(), false),
						},
						"uri": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
					},
				},
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsGlueUserDefinedFunctionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn
	catalogID := createAwsGlueCatalogID(d, meta.(*AWSClient).accountid)
	dbName := d.Get("database_name").(string)
	funcName := d.Get("name").(string)

	input := &glue.CreateUserDefinedFunctionInput{
		CatalogId:     aws.String(catalogID),
		DatabaseName:  aws.String(dbName),
		FunctionInput: expandAwsGlueUserDefinedFunctionInput(d),
	}

	_, err := conn.CreateUserDefinedFunction(input)
	if err != nil {
		return fmt.Errorf("error creating Glue User Defined Function: %w", err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", catalogID, dbName, funcName))

	return resourceAwsGlueUserDefinedFunctionRead(d, meta)
}

func resourceAwsGlueUserDefinedFunctionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	catalogID, dbName, funcName, err := readAwsGlueUDFID(d.Id())
	if err != nil {
		return err
	}

	input := &glue.UpdateUserDefinedFunctionInput{
		CatalogId:     aws.String(catalogID),
		DatabaseName:  aws.String(dbName),
		FunctionName:  aws.String(funcName),
		FunctionInput: expandAwsGlueUserDefinedFunctionInput(d),
	}

	if _, err := conn.UpdateUserDefinedFunction(input); err != nil {
		return fmt.Errorf("error updating Glue User Defined Function (%s): %w", d.Id(), err)
	}

	return resourceAwsGlueUserDefinedFunctionRead(d, meta)
}

func resourceAwsGlueUserDefinedFunctionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	catalogID, dbName, funcName, err := readAwsGlueUDFID(d.Id())
	if err != nil {
		return err
	}

	input := &glue.GetUserDefinedFunctionInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		FunctionName: aws.String(funcName),
	}

	out, err := conn.GetUserDefinedFunction(input)
	if err != nil {

		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			log.Printf("[WARN] Glue User Defined Function (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error reading Glue User Defined Function: %w", err)
	}

	udf := out.UserDefinedFunction

	udfArn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "glue",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("userDefinedFunction/%s/%s", dbName, aws.StringValue(udf.FunctionName)),
	}.String()

	d.Set("arn", udfArn)
	d.Set("name", udf.FunctionName)
	d.Set("catalog_id", catalogID)
	d.Set("database_name", dbName)
	d.Set("owner_type", udf.OwnerType)
	d.Set("owner_name", udf.OwnerName)
	d.Set("class_name", udf.ClassName)
	if udf.CreateTime != nil {
		d.Set("create_time", udf.CreateTime.Format(time.RFC3339))
	}
	if err := d.Set("resource_uris", flattenAwsGlueUserDefinedFunctionResourceUri(udf.ResourceUris)); err != nil {
		return err
	}

	return nil
}

func resourceAwsGlueUserDefinedFunctionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn
	catalogID, dbName, funcName, err := readAwsGlueUDFID(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Glue User Defined Function: %s", d.Id())
	_, err = conn.DeleteUserDefinedFunction(&glue.DeleteUserDefinedFunctionInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		FunctionName: aws.String(funcName),
	})
	if err != nil {
		return fmt.Errorf("error deleting Glue User Defined Function: %w", err)
	}
	return nil
}

func readAwsGlueUDFID(id string) (catalogID string, dbName string, funcName string, err error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 3 {
		return "", "", "", fmt.Errorf("unexpected format of ID (%q), expected CATALOG-ID:DATABASE-NAME:FUNCTION-NAME", id)
	}
	return idParts[0], idParts[1], idParts[2], nil
}

func expandAwsGlueUserDefinedFunctionInput(d *schema.ResourceData) *glue.UserDefinedFunctionInput {

	udf := &glue.UserDefinedFunctionInput{
		ClassName:    aws.String(d.Get("class_name").(string)),
		FunctionName: aws.String(d.Get("name").(string)),
		OwnerName:    aws.String(d.Get("owner_name").(string)),
		OwnerType:    aws.String(d.Get("owner_type").(string)),
	}

	if v, ok := d.GetOk("resource_uris"); ok && v.(*schema.Set).Len() > 0 {
		udf.ResourceUris = expandAwsGlueUserDefinedFunctionResourceUri(d.Get("resource_uris").(*schema.Set))
	}

	return udf
}

func expandAwsGlueUserDefinedFunctionResourceUri(conf *schema.Set) []*glue.ResourceUri {
	result := make([]*glue.ResourceUri, 0, conf.Len())

	for _, r := range conf.List() {
		uriRaw, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		uri := &glue.ResourceUri{
			ResourceType: aws.String(uriRaw["resource_type"].(string)),
			Uri:          aws.String(uriRaw["uri"].(string)),
		}

		result = append(result, uri)
	}

	return result
}

func flattenAwsGlueUserDefinedFunctionResourceUri(uris []*glue.ResourceUri) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(uris))

	for _, i := range uris {
		l := map[string]interface{}{
			"resource_type": aws.StringValue(i.ResourceType),
			"uri":           aws.StringValue(i.Uri),
		}

		result = append(result, l)
	}
	return result
}
