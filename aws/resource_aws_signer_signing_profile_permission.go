package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/signer"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
)

func resourceAwsSignerSigningProfilePermission() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSignerSigningProfilePermissionCreate,
		Read:   resourceAwsSignerSigningProfilePermissionRead,
		Delete: resourceAwsSignerSigningProfilePermissionDelete,

		Importer: &schema.ResourceImporter{
			State: resourceAwsSignerSigningProfilePermissionImport,
		},

		Schema: map[string]*schema.Schema{
			"profile_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(2, 64),
			},
			"action": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"signer:StartSigningJob",
					"signer:GetSigningProfile",
					"signer:RevokeSignature"},
					false),
			},
			"principal": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"profile_version": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(10, 10),
			},
			"statement_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"statement_id_prefix"},
				ValidateFunc:  validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-_]{0,64}$`), "must be alphanumeric with max length of 64 characters"),
			},
			"statement_id_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"statement_id"},
				ValidateFunc:  validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-_]{0,38}$`), "must be alphanumeric with max length of 38 characters"),
			},
		},
	}
}

func resourceAwsSignerSigningProfilePermissionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).signerconn

	profileName := d.Get("profile_name").(string)

	awsMutexKV.Lock(profileName)
	defer awsMutexKV.Unlock(profileName)

	listProfilePermissionsInput := &signer.ListProfilePermissionsInput{
		ProfileName: aws.String(profileName),
	}

	var revisionId string
	getProfilePermissionsOutput, err := conn.ListProfilePermissions(listProfilePermissionsInput)
	if err != nil {
		if isAWSErr(err, signer.ErrCodeResourceNotFoundException, "") {
			revisionId = ""
		} else {
			return err
		}
	} else {
		revisionId = *getProfilePermissionsOutput.RevisionId
	}

	statementId := naming.Generate(d.Get("statement_id").(string), d.Get("statement_id_prefix").(string))

	addProfilePermissionInput := &signer.AddProfilePermissionInput{
		Action:      aws.String(d.Get("action").(string)),
		Principal:   aws.String(d.Get("principal").(string)),
		ProfileName: aws.String(profileName),
		RevisionId:  aws.String(revisionId),
		StatementId: aws.String(statementId),
	}

	if v, ok := d.GetOk("profile_version"); ok {
		addProfilePermissionInput.ProfileVersion = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Adding new Signer signing profile permission: %s", addProfilePermissionInput)
	//var addProfilePermissionOutput *signer.AddProfilePermissionOutput
	// Retry for IAM eventual consistency
	err = resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		var err error
		_, err = conn.AddProfilePermission(addProfilePermissionInput)

		if isAWSErr(err, signer.ErrCodeConflictException, "") || isAWSErr(err, signer.ErrCodeResourceNotFoundException, "") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.AddProfilePermission(addProfilePermissionInput)
	}

	if err != nil {
		return fmt.Errorf("error adding new Signer signing profile permission for %q: %s", profileName, err)
	}

	err = resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		// IAM is eventually consistent :/
		err := resourceAwsSignerSigningProfilePermissionRead(d, meta)
		if err != nil {
			if isAWSErr(err, signer.ErrCodeResourceNotFoundException, "") {
				return resource.RetryableError(
					fmt.Errorf("error reading newly created Signer signing profile permission for %s, retrying: %s",
						*addProfilePermissionInput.ProfileName, err))
			}

			log.Printf("[ERROR] An actual error occurred when expecting Signer signing profile permission to be there: %s", err)
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		err = resourceAwsSignerSigningProfilePermissionRead(d, meta)
	}
	if err != nil {
		return fmt.Errorf("error reading new Signer permissions: %s", err)
	}

	d.Set("statement_id", statementId)
	d.SetId(fmt.Sprintf("%s/%s", profileName, statementId))

	return nil
}

func resourceAwsSignerSigningProfilePermissionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).signerconn

	listProfilePermissionsInput := &signer.ListProfilePermissionsInput{
		ProfileName: aws.String(d.Get("profile_name").(string)),
	}

	log.Printf("[DEBUG] Getting Signer signing profile permissions: %s", listProfilePermissionsInput)
	var listProfilePermissionsOutput *signer.ListProfilePermissionsOutput
	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		// IAM is eventually consistent :/
		var err error
		listProfilePermissionsOutput, err = conn.ListProfilePermissions(listProfilePermissionsInput)
		if err != nil {
			if isAWSErr(err, signer.ErrCodeResourceNotFoundException, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if isResourceTimeoutError(err) {
		listProfilePermissionsOutput, err = conn.ListProfilePermissions(listProfilePermissionsInput)
	}

	if !d.IsNewResource() && isAWSErr(err, signer.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] No Signer Signing Profile Permissions found (%s), removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Signer Signing Profile Permissions (%s): %w", d.Id(), err)
	}

	statementId := d.Get("statement_id").(string)
	permission := getProfilePermission(listProfilePermissionsOutput.Permissions, statementId)
	if permission == nil {
		log.Printf("[WARN] No Signer signing profile permission found matching statement id: %s", statementId)
		d.SetId("")
		return nil
	}

	if err := d.Set("action", permission.Action); err != nil {
		return fmt.Errorf("error setting signer signing profile permission action: %s", err)
	}
	if err := d.Set("principal", permission.Principal); err != nil {
		return fmt.Errorf("error setting signer signing profile permission principal: %s", err)
	}
	if err := d.Set("profile_version", permission.ProfileVersion); err != nil {
		return fmt.Errorf("error setting signer signing profile permission profile version: %s", err)
	}
	if err := d.Set("statement_id", permission.StatementId); err != nil {
		return fmt.Errorf("error setting signer signing profile permission statement id: %s", err)
	}

	return nil
}

func getProfilePermission(permissions []*signer.Permission, statementId string) *signer.Permission {
	for _, permission := range permissions {
		if permission == nil {
			continue
		}

		if aws.StringValue(permission.StatementId) == statementId {
			return permission
		}
	}

	return nil
}

func resourceAwsSignerSigningProfilePermissionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).signerconn

	profileName := d.Get("profile_name").(string)

	awsMutexKV.Lock(profileName)
	defer awsMutexKV.Unlock(profileName)

	listProfilePermissionsInput := &signer.ListProfilePermissionsInput{
		ProfileName: aws.String(profileName),
	}

	listProfilePermissionsOutput, err := conn.ListProfilePermissions(listProfilePermissionsInput)
	if err != nil {
		if isAWSErr(err, signer.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] No Signer signing profile permission found for: %v", listProfilePermissionsInput)
			return nil
		}
		return err
	}

	revisionId := aws.StringValue(listProfilePermissionsOutput.RevisionId)

	statementId := d.Get("statement_id").(string)
	permission := getProfilePermission(listProfilePermissionsOutput.Permissions, statementId)
	if permission == nil {
		log.Printf("[WARN] No Signer signing profile permission found matching statement id: %s", statementId)
		return nil
	}

	removeProfilePermissionInput := &signer.RemoveProfilePermissionInput{
		ProfileName: aws.String(profileName),
		RevisionId:  aws.String(revisionId),
		StatementId: permission.StatementId,
	}

	log.Printf("[DEBUG] Removing Signer singing profile permission: %s", removeProfilePermissionInput)
	_, err = conn.RemoveProfilePermission(removeProfilePermissionInput)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, signer.ErrCodeResourceNotFoundException) {
			log.Printf("[WARN] No Signer Signing Profile Permission found: %v", removeProfilePermissionInput)
			return nil
		}
		return fmt.Errorf("error removing Signer Signing Profile Permission (%s): %w", d.Id(), err)
	}

	params := &signer.ListProfilePermissionsInput{
		ProfileName: aws.String(profileName),
	}

	resp, err := conn.ListProfilePermissions(params)

	if isAWSErr(err, signer.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Signer signing profile permissions: %s", err)
	}

	if len(resp.Permissions) > 0 {
		permission := getProfilePermission(resp.Permissions, statementId)
		if permission != nil {
			return fmt.Errorf("failed to delete Signer singing profile permission with ID %q", statementId)
		}
	}

	log.Printf("[DEBUG] Signer signing profile permission with ID %q removed", statementId)

	return nil
}

func resourceAwsSignerSigningProfilePermissionImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected PROFILE_NAME/STATEMENT_ID", d.Id())
	}

	profileName := idParts[0]
	statementId := idParts[1]

	d.Set("profile_name", profileName)
	d.Set("statement_id", statementId)
	d.SetId(statementId)
	return []*schema.ResourceData{d}, nil
}
