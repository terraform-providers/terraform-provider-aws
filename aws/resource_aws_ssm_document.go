package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

const (
	SSM_DOCUMENT_PERMISSIONS_BATCH_LIMIT = 20
)

func resourceAwsSsmDocument() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSsmDocumentCreate,
		Read:   resourceAwsSsmDocumentRead,
		Update: resourceAwsSsmDocumentUpdate,
		Delete: resourceAwsSsmDocumentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateAwsSSMName,
			},
			"attachments_source": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								ssm.AttachmentsSourceKeyAttachmentReference,
								ssm.AttachmentsSourceKeySourceUrl,
								ssm.AttachmentsSourceKeyS3fileUrl,
							}, false),
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"values": {
							Type:     schema.TypeList,
							MinItems: 1,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"content": {
				Type:     schema.TypeString,
				Required: true,
			},
			"document_format": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  ssm.DocumentFormatJson,
				ValidateFunc: validation.StringInSlice([]string{
					ssm.DocumentFormatJson,
					ssm.DocumentFormatYaml,
				}, false),
			},
			"document_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					ssm.DocumentTypeCommand,
					ssm.DocumentTypePolicy,
					ssm.DocumentTypeAutomation,
					ssm.DocumentTypeSession,
					ssm.DocumentTypePackage,
				}, false),
			},
			"schema_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"document_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hash": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hash_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"latest_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"parameter": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"default_value": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"permissions": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"tags": tagsSchema(),
			"target_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAwsSsmDocumentCreate(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	// Validates permissions keys, if set, to be type and account_ids
	// since ValidateFunc validates only the value not the key.
	if v, ok := d.GetOk("permissions"); ok {
		if errors := validateSSMDocumentPermissions(v.(map[string]interface{})); len(errors) > 0 {
			return fmt.Errorf("Error validating Permissions: %v", errors)
		}
	}

	log.Printf("[INFO] Creating SSM Document: %s", d.Get("name").(string))

	docInput := &ssm.CreateDocumentInput{
		Name:           aws.String(d.Get("name").(string)),
		Content:        aws.String(d.Get("content").(string)),
		DocumentFormat: aws.String(d.Get("document_format").(string)),
		DocumentType:   aws.String(d.Get("document_type").(string)),
	}

	if v, ok := d.GetOk("tags"); ok {
		docInput.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().SsmTags()
	}

	if v, ok := d.GetOk("attachments_source"); ok {
		docInput.Attachments = expandSsmAttachmentsSources(v.([]interface{}))
	}

	if v, ok := d.GetOk("target_type"); ok {
		docInput.TargetType = aws.String(v.(string))
	}

	resp, err := ssmconn.CreateDocument(docInput)

	if err != nil {
		return fmt.Errorf("Error creating SSM document: %s", err)
	}

	d.SetId(*resp.DocumentDescription.Name)

	if v, ok := d.GetOk("permissions"); ok && v != nil {
		if err := setDocumentPermissions(d, meta); err != nil {
			return err
		}
	} else {
		log.Printf("[DEBUG] Not setting permissions for %q", d.Id())
	}

	return resourceAwsSsmDocumentRead(d, meta)
}

func resourceAwsSsmDocumentRead(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading SSM Document: %s", d.Id())

	describeDocumentInput := &ssm.DescribeDocumentInput{
		Name: aws.String(d.Id()),
	}

	describeDocumentOutput, err := ssmconn.DescribeDocument(describeDocumentInput)

	if isAWSErr(err, ssm.ErrCodeInvalidDocument, "") {
		log.Printf("[WARN] SSM Document not found so removing from state")
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing SSM Document (%s): %s", d.Id(), err)
	}

	if describeDocumentOutput == nil || describeDocumentOutput.Document == nil {
		return fmt.Errorf("error describing SSM Document (%s): empty result", d.Id())
	}

	getDocumentInput := &ssm.GetDocumentInput{
		DocumentFormat:  describeDocumentOutput.Document.DocumentFormat,
		DocumentVersion: aws.String("$LATEST"),
		Name:            describeDocumentOutput.Document.Name,
	}

	getDocumentOutput, err := ssmconn.GetDocument(getDocumentInput)

	if err != nil {
		return fmt.Errorf("error getting SSM Document (%s): %s", d.Id(), err)
	}

	if getDocumentOutput == nil {
		return fmt.Errorf("error getting SSM Document (%s): empty result", d.Id())
	}

	doc := describeDocumentOutput.Document

	d.Set("content", getDocumentOutput.Content)
	d.Set("created_date", aws.TimeValue(doc.CreatedDate).Format(time.RFC3339))
	d.Set("default_version", doc.DefaultVersion)
	d.Set("description", doc.Description)
	d.Set("schema_version", doc.SchemaVersion)
	d.Set("document_format", doc.DocumentFormat)
	d.Set("document_type", doc.DocumentType)
	d.Set("document_version", doc.DocumentVersion)
	d.Set("hash", doc.Hash)
	d.Set("hash_type", doc.HashType)
	d.Set("latest_version", doc.LatestVersion)
	d.Set("name", doc.Name)
	d.Set("owner", doc.Owner)
	d.Set("platform_types", flattenStringList(doc.PlatformTypes))
	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "ssm",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("document/%s", *doc.Name),
	}.String()
	if err := d.Set("arn", arn); err != nil {
		return fmt.Errorf("Error setting arn error: %#v", err)
	}

	d.Set("status", doc.Status)

	gp, err := getDocumentPermissions(d, meta)

	if err != nil {
		return fmt.Errorf("Error reading SSM document permissions: %s", err)
	}

	d.Set("permissions", gp)

	params := make([]map[string]interface{}, 0)
	for i := 0; i < len(doc.Parameters); i++ {

		dp := doc.Parameters[i]
		param := make(map[string]interface{})

		if dp.DefaultValue != nil {
			param["default_value"] = *dp.DefaultValue
		}
		if dp.Description != nil {
			param["description"] = *dp.Description
		}
		if dp.Name != nil {
			param["name"] = *dp.Name
		}
		if dp.Type != nil {
			param["type"] = *dp.Type
		}
		params = append(params, param)
	}

	if err := d.Set("parameter", params); err != nil {
		return err
	}

	if err := d.Set("tags", keyvaluetags.SsmKeyValueTags(doc.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("target_type", doc.TargetType); err != nil {
		return fmt.Errorf("error setting target type: %s", err)
	}

	return nil
}

func resourceAwsSsmDocumentUpdate(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	// Validates permissions keys, if set, to be type and account_ids
	// since ValidateFunc validates only the value not the key.
	if v, ok := d.GetOk("permissions"); ok {
		if errors := validateSSMDocumentPermissions(v.(map[string]interface{})); len(errors) > 0 {
			return fmt.Errorf("Error validating Permissions: %v", errors)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.SsmUpdateTags(ssmconn, d.Id(), ssm.ResourceTypeForTaggingDocument, o, n); err != nil {
			return fmt.Errorf("error updating SSM Document (%s) tags: %s", d.Id(), err)
		}
	}

	if d.HasChange("permissions") {
		if err := setDocumentPermissions(d, meta); err != nil {
			return err
		}
	} else {
		log.Printf("[DEBUG] Not setting document permissions on %q", d.Id())
	}

	// update for schema version 1.x is not allowed
	isSchemaVersion1, _ := regexp.MatchString("^1[.][0-9]$", d.Get("schema_version").(string))

	if !d.HasChange("content") && isSchemaVersion1 {
		return nil
	}

	if err := updateAwsSSMDocument(d, meta); err != nil {
		return err
	}

	return resourceAwsSsmDocumentRead(d, meta)
}

func resourceAwsSsmDocumentDelete(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	if err := deleteDocumentPermissions(d, meta); err != nil {
		return err
	}

	log.Printf("[INFO] Deleting SSM Document: %s", d.Id())

	params := &ssm.DeleteDocumentInput{
		Name: aws.String(d.Get("name").(string)),
	}

	_, err := ssmconn.DeleteDocument(params)
	if err != nil {
		return err
	}

	input := &ssm.DescribeDocumentInput{
		Name: aws.String(d.Get("name").(string)),
	}
	log.Printf("[DEBUG] Waiting for SSM Document %q to be deleted", d.Get("name").(string))
	err = resource.Retry(10*time.Minute, func() *resource.RetryError {
		_, err := ssmconn.DescribeDocument(input)

		if isAWSErr(err, ssm.ErrCodeInvalidDocument, "") {
			return nil
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return resource.RetryableError(fmt.Errorf("SSM Document (%s) still exists", d.Id()))
	})

	if isResourceTimeoutError(err) {
		_, err = ssmconn.DescribeDocument(input)
	}
	if isAWSErr(err, ssm.ErrCodeInvalidDocument, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error waiting for SSM Document (%s) deletion: %s", d.Id(), err)
	}
	return nil
}

func expandSsmAttachmentsSources(a []interface{}) []*ssm.AttachmentsSource {
	if len(a) == 0 {
		return nil
	}

	results := make([]*ssm.AttachmentsSource, 0)
	for _, raw := range a {
		at := raw.(map[string]interface{})
		s := &ssm.AttachmentsSource{}
		if val, ok := at["key"]; ok {
			s.Key = aws.String(val.(string))
		}
		if val, ok := at["name"]; ok && val != "" {
			s.Name = aws.String(val.(string))
		}
		if val, ok := at["values"]; ok {
			s.Values = expandStringList(val.([]interface{}))
		}

		results = append(results, s)
	}
	return results

}

func setDocumentPermissions(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[INFO] Setting permissions for document: %s", d.Id())

	if d.HasChange("permissions") {
		o, n := d.GetChange("permissions")
		oldPermissions := o.(map[string]interface{})
		newPermissions := n.(map[string]interface{})
		oldPermissionsAccountIds := make([]interface{}, 0)
		if v, ok := oldPermissions["account_ids"]; ok && v.(string) != "" {
			parts := strings.Split(v.(string), ",")
			oldPermissionsAccountIds = make([]interface{}, len(parts))
			for i, v := range parts {
				oldPermissionsAccountIds[i] = v
			}
		}
		newPermissionsAccountIds := make([]interface{}, 0)
		if v, ok := newPermissions["account_ids"]; ok && v.(string) != "" {
			parts := strings.Split(v.(string), ",")
			newPermissionsAccountIds = make([]interface{}, len(parts))
			for i, v := range parts {
				newPermissionsAccountIds[i] = v
			}
		}

		// Since AccountIdsToRemove has higher priority than AccountIdsToAdd,
		// we filter out accounts from both lists
		accountIdsToRemove := make([]interface{}, 0)
		for _, oldPermissionsAccountId := range oldPermissionsAccountIds {
			if _, contains := sliceContainsString(newPermissionsAccountIds, oldPermissionsAccountId.(string)); !contains {
				accountIdsToRemove = append(accountIdsToRemove, oldPermissionsAccountId.(string))
			}
		}
		accountIdsToAdd := make([]interface{}, 0)
		for _, newPermissionsAccountId := range newPermissionsAccountIds {
			if _, contains := sliceContainsString(oldPermissionsAccountIds, newPermissionsAccountId.(string)); !contains {
				accountIdsToAdd = append(accountIdsToAdd, newPermissionsAccountId.(string))
			}
		}

		if err := modifyDocumentPermissions(ssmconn, d.Get("name").(string), accountIdsToAdd, accountIdsToRemove); err != nil {
			return fmt.Errorf("error modifying SSM document permissions: %s", err)
		}

	}

	return nil
}

func getDocumentPermissions(d *schema.ResourceData, meta interface{}) (map[string]interface{}, error) {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[INFO] Getting permissions for document: %s", d.Id())

	//How to get from nested scheme resource?
	permissionType := "Share"

	permInput := &ssm.DescribeDocumentPermissionInput{
		Name:           aws.String(d.Get("name").(string)),
		PermissionType: aws.String(permissionType),
	}

	resp, err := ssmconn.DescribeDocumentPermission(permInput)

	if err != nil {
		return nil, fmt.Errorf("Error setting permissions for SSM document: %s", err)
	}

	var account_ids = make([]string, len(resp.AccountIds))
	for i := 0; i < len(resp.AccountIds); i++ {
		account_ids[i] = *resp.AccountIds[i]
	}

	ids := ""
	if len(account_ids) == 1 {
		ids = account_ids[0]
	} else if len(account_ids) > 1 {
		ids = strings.Join(account_ids, ",")
	}

	if ids == "" {
		return nil, nil
	}

	perms := make(map[string]interface{})
	perms["type"] = permissionType
	perms["account_ids"] = ids

	return perms, nil
}

func deleteDocumentPermissions(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[INFO] Removing permissions from document: %s", d.Id())

	permission := d.Get("permissions").(map[string]interface{})

	accountIdsToRemove := make([]interface{}, 0)

	if permission["account_ids"] != nil {

		if v, ok := permission["account_ids"]; ok && v.(string) != "" {
			parts := strings.Split(v.(string), ",")
			accountIdsToRemove = make([]interface{}, len(parts))
			for i, v := range parts {
				accountIdsToRemove[i] = v
			}
		}

		if err := modifyDocumentPermissions(ssmconn, d.Get("name").(string), nil, accountIdsToRemove); err != nil {
			return fmt.Errorf("error removing SSM document permissions: %s", err)
		}

	}

	return nil
}

func modifyDocumentPermissions(conn *ssm.SSM, name string, accountIdsToAdd []interface{}, accountIdstoRemove []interface{}) error {

	if accountIdsToAdd != nil {

		accountIdsToAddBatch := make([]string, 0, SSM_DOCUMENT_PERMISSIONS_BATCH_LIMIT)
		accountIdsToAddBatches := make([][]string, 0, len(accountIdsToAdd)/SSM_DOCUMENT_PERMISSIONS_BATCH_LIMIT+1)
		for _, accountId := range accountIdsToAdd {
			if len(accountIdsToAddBatch) == SSM_DOCUMENT_PERMISSIONS_BATCH_LIMIT {
				accountIdsToAddBatches = append(accountIdsToAddBatches, accountIdsToAddBatch)
				accountIdsToAddBatch = make([]string, 0, SSM_DOCUMENT_PERMISSIONS_BATCH_LIMIT)
			}
			accountIdsToAddBatch = append(accountIdsToAddBatch, accountId.(string))
		}
		accountIdsToAddBatches = append(accountIdsToAddBatches, accountIdsToAddBatch)

		for _, accountIdsToAdd := range accountIdsToAddBatches {
			_, err := conn.ModifyDocumentPermission(&ssm.ModifyDocumentPermissionInput{
				Name:            aws.String(name),
				PermissionType:  aws.String("Share"),
				AccountIdsToAdd: aws.StringSlice(accountIdsToAdd),
			})
			if err != nil {
				return err
			}
		}
	}

	if accountIdstoRemove != nil {

		accountIdsToRemoveBatch := make([]string, 0, SSM_DOCUMENT_PERMISSIONS_BATCH_LIMIT)
		accountIdsToRemoveBatches := make([][]string, 0, len(accountIdstoRemove)/SSM_DOCUMENT_PERMISSIONS_BATCH_LIMIT+1)
		for _, accountId := range accountIdstoRemove {
			if len(accountIdsToRemoveBatch) == SSM_DOCUMENT_PERMISSIONS_BATCH_LIMIT {
				accountIdsToRemoveBatches = append(accountIdsToRemoveBatches, accountIdsToRemoveBatch)
				accountIdsToRemoveBatch = make([]string, 0, SSM_DOCUMENT_PERMISSIONS_BATCH_LIMIT)
			}
			accountIdsToRemoveBatch = append(accountIdsToRemoveBatch, accountId.(string))
		}
		accountIdsToRemoveBatches = append(accountIdsToRemoveBatches, accountIdsToRemoveBatch)

		for _, accountIdsToRemove := range accountIdsToRemoveBatches {
			_, err := conn.ModifyDocumentPermission(&ssm.ModifyDocumentPermissionInput{
				Name:               aws.String(name),
				PermissionType:     aws.String("Share"),
				AccountIdsToRemove: aws.StringSlice(accountIdsToRemove),
			})
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func updateAwsSSMDocument(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[INFO] Updating SSM Document: %s", d.Id())

	name := d.Get("name").(string)

	updateDocInput := &ssm.UpdateDocumentInput{
		Name:            aws.String(name),
		Content:         aws.String(d.Get("content").(string)),
		DocumentFormat:  aws.String(d.Get("document_format").(string)),
		DocumentVersion: aws.String(d.Get("default_version").(string)),
	}

	if v, ok := d.GetOk("target_type"); ok {
		updateDocInput.TargetType = aws.String(v.(string))
	}

	if d.HasChange("attachments_source") {
		updateDocInput.Attachments = expandSsmAttachmentsSources(d.Get("attachments_source").([]interface{}))
	}

	newDefaultVersion := d.Get("default_version").(string)

	ssmconn := meta.(*AWSClient).ssmconn
	updated, err := ssmconn.UpdateDocument(updateDocInput)

	if isAWSErr(err, ssm.ErrCodeDuplicateDocumentContent, "") {
		log.Printf("[DEBUG] Content is a duplicate of the latest version so update is not necessary: %s", d.Id())
		log.Printf("[INFO] Updating the default version to the latest version %s: %s", newDefaultVersion, d.Id())

		newDefaultVersion = d.Get("latest_version").(string)
	} else if err != nil {
		return fmt.Errorf("Error updating SSM document: %s", err)
	} else {
		log.Printf("[INFO] Updating the default version to the new version %s: %s", newDefaultVersion, d.Id())
		newDefaultVersion = *updated.DocumentDescription.DocumentVersion
	}

	updateDefaultInput := &ssm.UpdateDocumentDefaultVersionInput{
		Name:            aws.String(name),
		DocumentVersion: aws.String(newDefaultVersion),
	}

	_, err = ssmconn.UpdateDocumentDefaultVersion(updateDefaultInput)

	if err != nil {
		return fmt.Errorf("Error updating the default document version to that of the updated document: %s", err)
	}
	return nil
}

//Validates that type and account_ids are defined
func validateSSMDocumentPermissions(v map[string]interface{}) (errors []error) {
	k := "permissions"
	t, hasType := v["type"].(string)
	_, hasAccountIds := v["account_ids"].(string)

	if hasType {
		if t != ssm.DocumentPermissionTypeShare {
			errors = append(errors, fmt.Errorf("%q: only %s \"type\" supported", k, ssm.DocumentPermissionTypeShare))
		}
	} else {
		errors = append(errors, fmt.Errorf("%q: \"type\" must be defined", k))
	}
	if !hasAccountIds {
		errors = append(errors, fmt.Errorf("%q: \"account_ids\" must be defined", k))
	}

	return
}
