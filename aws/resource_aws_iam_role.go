package aws

import (
	"fmt"
	"log"
	"net/url"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
)

func resourceAwsIamRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIamRoleCreate,
		Read:   resourceAwsIamRoleRead,
		Update: resourceAwsIamRoleUpdate,
		Delete: resourceAwsIamRoleDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsIamRoleImport,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"unique_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[\w+=,.@-]*$`), "must match [\\w+=,.@-]"),
				),
			},

			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexp.MustCompile(`^[\w+=,.@-]*$`), "must match [\\w+=,.@-]"),
				),
			},

			"path": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/",
				ForceNew: true,
			},

			"permissions_boundary": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 1000),
					validation.StringDoesNotMatch(regexp.MustCompile("[“‘]"), "cannot contain specially formatted single or double quotes: [“‘]"),
					validation.StringMatch(regexp.MustCompile(`[\p{L}\p{M}\p{Z}\p{S}\p{N}\p{P}]*`), `must satisfy regular expression pattern: [\p{L}\p{M}\p{Z}\p{S}\p{N}\p{P}]*)`),
				),
			},

			"assume_role_policy": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
				ValidateFunc:     validation.StringIsJSON,
			},

			"force_detach_policies": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"create_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"max_session_duration": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      3600,
				ValidateFunc: validation.IntBetween(3600, 43200),
			},

			"tags": tagsSchema(),

			"inline_policy": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.All(
								validation.StringIsNotEmpty,
								validateIamRolePolicyName,
							),
						},
						"policy": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateFunc:     validateIAMPolicyJson,
							DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
						},
					},
				},
			},

			"managed_policy_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func resourceAwsIamRoleImport(
	d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("force_detach_policies", false)
	return []*schema.ResourceData{d}, nil
}

func resourceAwsIamRoleCreate(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		name = resource.PrefixedUniqueId(v.(string))
	} else {
		name = resource.UniqueId()
	}

	request := &iam.CreateRoleInput{
		Path:                     aws.String(d.Get("path").(string)),
		RoleName:                 aws.String(name),
		AssumeRolePolicyDocument: aws.String(d.Get("assume_role_policy").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		request.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_session_duration"); ok {
		request.MaxSessionDuration = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("permissions_boundary"); ok {
		request.PermissionsBoundary = aws.String(v.(string))
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		request.Tags = keyvaluetags.New(v).IgnoreAws().IamTags()
	}

	var createResp *iam.CreateRoleOutput
	err := resource.Retry(30*time.Second, func() *resource.RetryError {
		var err error
		createResp, err = iamconn.CreateRole(request)
		// IAM users (referenced in Principal field of assume policy)
		// can take ~30 seconds to propagate in AWS
		if isAWSErr(err, "MalformedPolicyDocument", "Invalid principal in policy") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		createResp, err = iamconn.CreateRole(request)
	}
	if err != nil {
		return fmt.Errorf("Error creating IAM Role %s: %s", name, err)
	}

	roleName := aws.StringValue(createResp.Role.RoleName)

	if v, ok := d.GetOk("inline_policy"); ok && v.(*schema.Set).Len() > 0 {
		policies := expandIamInlinePolicies(roleName, v.(*schema.Set).List())
		if err := addIamInlinePolicies(policies, meta); err != nil {
			return err
		}
	}

	if v, ok := d.GetOk("managed_policy_arns"); ok && v.(*schema.Set).Len() > 0 {
		managedPolicies := expandStringSet(v.(*schema.Set))
		if err := addIamManagedPolicies(roleName, managedPolicies, meta); err != nil {
			return err
		}
	}

	d.SetId(roleName)
	return resourceAwsIamRoleRead(d, meta)
}

func resourceAwsIamRoleRead(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	request := &iam.GetRoleInput{
		RoleName: aws.String(d.Id()),
	}

	getResp, err := iamconn.GetRole(request)
	if err != nil {
		if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			log.Printf("[WARN] IAM Role %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading IAM Role %s: %s", d.Id(), err)
	}

	if getResp == nil || getResp.Role == nil {
		log.Printf("[WARN] IAM Role %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	role := getResp.Role

	d.Set("arn", role.Arn)
	if err := d.Set("create_date", role.CreateDate.Format(time.RFC3339)); err != nil {
		return err
	}
	d.Set("description", role.Description)
	d.Set("max_session_duration", role.MaxSessionDuration)
	d.Set("name", role.RoleName)
	d.Set("path", role.Path)
	if role.PermissionsBoundary != nil {
		d.Set("permissions_boundary", role.PermissionsBoundary.PermissionsBoundaryArn)
	}
	d.Set("unique_id", role.RoleId)

	if err := d.Set("tags", keyvaluetags.IamKeyValueTags(role.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	assumeRolePolicy, err := url.QueryUnescape(*role.AssumeRolePolicyDocument)
	if err != nil {
		return err
	}
	if err := d.Set("assume_role_policy", assumeRolePolicy); err != nil {
		return err
	}

	inlinePolicies, err := readIamInlinePolicies(aws.StringValue(role.RoleName), meta)
	if err != nil {
		return fmt.Errorf("reading inline policies for IAM role %s, error: %s", d.Id(), err)
	}
	if err := d.Set("inline_policy", flattenIamInlinePolicies(inlinePolicies)); err != nil {
		return fmt.Errorf("error setting inline_policy: %w", err)
	}

	managedPolicies, err := readIamRolePolicyAttachments(iamconn, aws.StringValue(role.RoleName))
	if err != nil {
		return fmt.Errorf("reading managed policies for IAM role %s, error: %s", d.Id(), err)
	}
	d.Set("managed_policy_arns", managedPolicies)

	return nil
}

func resourceAwsIamRoleUpdate(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	if d.HasChange("assume_role_policy") {
		assumeRolePolicyInput := &iam.UpdateAssumeRolePolicyInput{
			RoleName:       aws.String(d.Id()),
			PolicyDocument: aws.String(d.Get("assume_role_policy").(string)),
		}
		_, err := iamconn.UpdateAssumeRolePolicy(assumeRolePolicyInput)
		if err != nil {
			if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
				d.SetId("")
				return nil
			}
			return fmt.Errorf("Error Updating IAM Role (%s) Assume Role Policy: %s", d.Id(), err)
		}
	}

	if d.HasChange("description") {
		roleDescriptionInput := &iam.UpdateRoleDescriptionInput{
			RoleName:    aws.String(d.Id()),
			Description: aws.String(d.Get("description").(string)),
		}
		_, err := iamconn.UpdateRoleDescription(roleDescriptionInput)
		if err != nil {
			if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
				d.SetId("")
				return nil
			}
			return fmt.Errorf("Error Updating IAM Role (%s) Assume Role Policy: %s", d.Id(), err)
		}
	}

	if d.HasChange("max_session_duration") {
		roleMaxDurationInput := &iam.UpdateRoleInput{
			RoleName:           aws.String(d.Id()),
			MaxSessionDuration: aws.Int64(int64(d.Get("max_session_duration").(int))),
		}
		_, err := iamconn.UpdateRole(roleMaxDurationInput)
		if err != nil {
			if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
				d.SetId("")
				return nil
			}
			return fmt.Errorf("Error Updating IAM Role (%s) Max Session Duration: %s", d.Id(), err)
		}
	}

	if d.HasChange("permissions_boundary") {
		permissionsBoundary := d.Get("permissions_boundary").(string)
		if permissionsBoundary != "" {
			input := &iam.PutRolePermissionsBoundaryInput{
				PermissionsBoundary: aws.String(permissionsBoundary),
				RoleName:            aws.String(d.Id()),
			}
			_, err := iamconn.PutRolePermissionsBoundary(input)
			if err != nil {
				return fmt.Errorf("error updating IAM Role permissions boundary: %s", err)
			}
		} else {
			input := &iam.DeleteRolePermissionsBoundaryInput{
				RoleName: aws.String(d.Id()),
			}
			_, err := iamconn.DeleteRolePermissionsBoundary(input)
			if err != nil {
				return fmt.Errorf("error deleting IAM Role permissions boundary: %s", err)
			}
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.IamRoleUpdateTags(iamconn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating IAM Role (%s) tags: %s", d.Id(), err)
		}
	}

	if d.HasChange("inline_policy") {
		roleName := d.Get("name").(string)
		o, n := d.GetChange("inline_policy")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		remove := os.Difference(ns).List()
		add := ns.Difference(os).List()

		var policyNames []*string
		for _, policy := range remove {
			tfMap, ok := policy.(map[string]interface{})

			if !ok {
				continue
			}

			if v, ok := tfMap["name"].(string); ok && v != "" {
				policyNames = append(policyNames, aws.String(tfMap["name"].(string)))
			}
		}
		if err := deleteIamRolePolicies(iamconn, roleName, policyNames); err != nil {
			return fmt.Errorf("unable to delete inline policies: %w", err)
		}

		policies := expandIamInlinePolicies(roleName, add)
		if err := addIamInlinePolicies(policies, meta); err != nil {
			return err
		}
	}

	if d.HasChange("managed_policy_arns") {
		roleName := d.Get("name").(string)

		o, n := d.GetChange("managed_policy_arns")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		remove := expandStringSet(os.Difference(ns))
		add := expandStringSet(ns.Difference(os))

		if err := deleteIamRolePolicyAttachments(iamconn, roleName, remove); err != nil {
			return fmt.Errorf("unable to detach policies: %w", err)
		}

		if err := addIamManagedPolicies(roleName, add, meta); err != nil {
			return err
		}
	}

	return resourceAwsIamRoleRead(d, meta)
}

func resourceAwsIamRoleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	hasInline := false
	if v, ok := d.GetOk("inline_policy"); ok && v.(*schema.Set).Len() > 0 {
		hasInline = true
	}

	hasManaged := false
	if v, ok := d.GetOk("managed_policy_arns"); ok && v.(*schema.Set).Len() > 0 {
		hasManaged = true
	}

	err := deleteIamRole(conn, d.Id(), d.Get("force_detach_policies").(bool), hasInline, hasManaged)
	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting IAM Role (%s): %w", d.Id(), err)
	}

	return nil
}

func deleteIamRole(conn *iam.IAM, roleName string, forceDetach, hasInline, hasManaged bool) error {
	if err := deleteIamRoleInstanceProfiles(conn, roleName); err != nil {
		return fmt.Errorf("unable to detach instance profiles: %w", err)
	}

	if forceDetach || hasManaged {
		managedPolicies, err := readIamRolePolicyAttachments(conn, roleName)
		if err != nil {
			return err
		}

		if err := deleteIamRolePolicyAttachments(conn, roleName, managedPolicies); err != nil {
			return fmt.Errorf("unable to detach policies: %w", err)
		}
	}

	if forceDetach || hasInline {
		inlinePolicies, err := readIamRolePolicyNames(conn, roleName)
		if err != nil {
			return err
		}

		if err := deleteIamRolePolicies(conn, roleName, inlinePolicies); err != nil {
			return fmt.Errorf("unable to delete inline policies: %w", err)
		}
	}

	deleteRoleInput := &iam.DeleteRoleInput{
		RoleName: aws.String(roleName),
	}
	err := resource.Retry(waiter.PropagationTimeout, func() *resource.RetryError {
		_, err := conn.DeleteRole(deleteRoleInput)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, iam.ErrCodeDeleteConflictException) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DeleteRole(deleteRoleInput)
	}

	return err
}

func deleteIamRoleInstanceProfiles(conn *iam.IAM, roleName string) error {
	resp, err := conn.ListInstanceProfilesForRole(&iam.ListInstanceProfilesForRoleInput{
		RoleName: aws.String(roleName),
	})
	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil
	}
	if err != nil {
		return err
	}

	// Loop and remove this Role from any Profiles
	for _, i := range resp.InstanceProfiles {
		input := &iam.RemoveRoleFromInstanceProfileInput{
			InstanceProfileName: i.InstanceProfileName,
			RoleName:            aws.String(roleName),
		}

		_, err := conn.RemoveRoleFromInstanceProfile(input)
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			continue
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func readIamRolePolicyAttachments(conn *iam.IAM, roleName string) ([]*string, error) {
	managedPolicies := make([]*string, 0)
	input := &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	}

	err := conn.ListAttachedRolePoliciesPages(input, func(page *iam.ListAttachedRolePoliciesOutput, lastPage bool) bool {
		for _, v := range page.AttachedPolicies {
			managedPolicies = append(managedPolicies, v.PolicyArn)
		}
		return !lastPage
	})
	if err != nil && !tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, err
	}

	return managedPolicies, nil
}

func deleteIamRolePolicyAttachments(conn *iam.IAM, roleName string, managedPolicies []*string) error {
	for _, parn := range managedPolicies {
		input := &iam.DetachRolePolicyInput{
			PolicyArn: parn,
			RoleName:  aws.String(roleName),
		}

		_, err := conn.DetachRolePolicy(input)
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			continue
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func readIamRolePolicyNames(conn *iam.IAM, roleName string) ([]*string, error) {
	inlinePolicies := make([]*string, 0)
	input := &iam.ListRolePoliciesInput{
		RoleName: aws.String(roleName),
	}

	err := conn.ListRolePoliciesPages(input, func(page *iam.ListRolePoliciesOutput, lastPage bool) bool {
		inlinePolicies = append(inlinePolicies, page.PolicyNames...)
		return !lastPage
	})

	if err != nil && !tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, err
	}

	return inlinePolicies, nil
}

func deleteIamRolePolicies(conn *iam.IAM, roleName string, policyNames []*string) error {
	for _, name := range policyNames {
		if len(aws.StringValue(name)) == 0 {
			continue
		}

		input := &iam.DeleteRolePolicyInput{
			PolicyName: name,
			RoleName:   aws.String(roleName),
		}

		_, err := conn.DeleteRolePolicy(input)
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return nil
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func flattenIamInlinePolicy(apiObject *iam.PutRolePolicyInput) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["name"] = aws.StringValue(apiObject.PolicyName)
	tfMap["policy"] = aws.StringValue(apiObject.PolicyDocument)

	return tfMap
}

func flattenIamInlinePolicies(apiObjects []*iam.PutRolePolicyInput) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenIamInlinePolicy(apiObject))
	}

	return tfList
}

func expandIamInlinePolicy(roleName string, tfMap map[string]interface{}) *iam.PutRolePolicyInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &iam.PutRolePolicyInput{
		RoleName: aws.String(roleName),
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.PolicyName = aws.String(v)
	}

	if v, ok := tfMap["policy"].(string); ok && v != "" {
		apiObject.PolicyDocument = aws.String(v)
	}

	return apiObject
}

func expandIamInlinePolicies(roleName string, tfList []interface{}) []*iam.PutRolePolicyInput {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*iam.PutRolePolicyInput

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandIamInlinePolicy(roleName, tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func addIamInlinePolicies(policies []*iam.PutRolePolicyInput, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	var errs *multierror.Error
	for _, policy := range policies {
		if len(aws.StringValue(policy.PolicyName)) == 0 || len(aws.StringValue(policy.PolicyDocument)) == 0 {
			continue
		}

		if _, err := conn.PutRolePolicy(policy); err != nil {
			newErr := fmt.Errorf("creating inline policy (%s): %w", aws.StringValue(policy.PolicyName), err)
			log.Printf("[ERROR] %s", newErr)
			errs = multierror.Append(errs, newErr)
		}
	}

	return errs.ErrorOrNil()
}

func addIamManagedPolicies(roleName string, policies []*string, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	var errs *multierror.Error
	for _, arn := range policies {
		if err := attachPolicyToRole(conn, roleName, aws.StringValue(arn)); err != nil {
			newErr := fmt.Errorf("attaching managed policy (%s): %w", aws.StringValue(arn), err)
			log.Printf("[ERROR] %s", newErr)
			errs = multierror.Append(errs, newErr)
		}
	}

	return errs.ErrorOrNil()
}

func readIamInlinePolicies(roleName string, meta interface{}) ([]*iam.PutRolePolicyInput, error) {
	conn := meta.(*AWSClient).iamconn

	policyNames, err := readIamRolePolicyNames(conn, roleName)
	if err != nil {
		return nil, err
	}

	var apiObjects []*iam.PutRolePolicyInput
	for _, policyName := range policyNames {
		policyResp, err := conn.GetRolePolicy(&iam.GetRolePolicyInput{
			RoleName:   aws.String(roleName),
			PolicyName: policyName,
		})
		if err != nil {
			return nil, err
		}

		policy, err := url.QueryUnescape(*policyResp.PolicyDocument)
		if err != nil {
			return nil, err
		}

		apiObject := &iam.PutRolePolicyInput{
			RoleName:       aws.String(roleName),
			PolicyDocument: aws.String(policy),
			PolicyName:     policyName,
		}

		apiObjects = append(apiObjects, apiObject)
	}

	if len(apiObjects) == 0 {
		apiObjects = append(apiObjects, &iam.PutRolePolicyInput{
			PolicyDocument: aws.String(""),
			PolicyName:     aws.String(""),
		})
	}

	return apiObjects, nil
}
