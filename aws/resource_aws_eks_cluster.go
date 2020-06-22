package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

var eksLogTypes = []string{
	eks.LogTypeApi,
	eks.LogTypeAudit,
	eks.LogTypeAuthenticator,
	eks.LogTypeControllerManager,
	eks.LogTypeScheduler,
}

func resourceAwsEksCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEksClusterCreate,
		Read:   resourceAwsEksClusterRead,
		Update: resourceAwsEksClusterUpdate,
		Delete: resourceAwsEksClusterDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_authority": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encryption_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"provider": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key_arn": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
								},
							},
						},
						"resources": {
							Type:     schema.TypeSet,
							MinItems: 1,
							Required: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									"secrets",
								}, false),
							},
						},
					},
				},
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identity": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"oidc": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"issuer": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"platform_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
			"version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"vpc_config": {
				Type:     schema.TypeList,
				MinItems: 1,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster_security_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"endpoint_private_access": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"endpoint_public_access": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"security_group_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							MinItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"public_access_cidrs": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validateCIDRNetworkAddress,
							},
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"enabled_cluster_log_types": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(eksLogTypes, true),
				},
				Set: schema.HashString,
			},
		},
	}
}

func resourceAwsEksClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).eksconn
	name := d.Get("name").(string)

	input := &eks.CreateClusterInput{
		EncryptionConfig:   expandEksEncryptionConfig(d.Get("encryption_config").([]interface{})),
		Name:               aws.String(name),
		RoleArn:            aws.String(d.Get("role_arn").(string)),
		ResourcesVpcConfig: expandEksVpcConfigRequest(d.Get("vpc_config").([]interface{})),
		Logging:            expandEksLoggingTypes(d.Get("enabled_cluster_log_types").(*schema.Set)),
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.Tags = keyvaluetags.New(v).IgnoreAws().EksTags()
	}

	if v, ok := d.GetOk("version"); ok && v.(string) != "" {
		input.Version = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EKS Cluster: %s", input)
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := conn.CreateCluster(input)
		if err != nil {
			// InvalidParameterException: roleArn, arn:aws:iam::123456789012:role/XXX, does not exist
			if isAWSErr(err, eks.ErrCodeInvalidParameterException, "does not exist") {
				return resource.RetryableError(err)
			}
			// InvalidParameterException: Error in role params
			if isAWSErr(err, eks.ErrCodeInvalidParameterException, "Error in role params") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, eks.ErrCodeInvalidParameterException, "Role could not be assumed because the trusted entity is not correct") {
				return resource.RetryableError(err)
			}
			// InvalidParameterException: The provided role doesn't have the Amazon EKS Managed Policies associated with it. Please ensure the following policies [arn:aws:iam::aws:policy/AmazonEKSClusterPolicy, arn:aws:iam::aws:policy/AmazonEKSServicePolicy] are attached
			if isAWSErr(err, eks.ErrCodeInvalidParameterException, "The provided role doesn't have the Amazon EKS Managed Policies associated with it") {
				return resource.RetryableError(err)
			}
			// InvalidParameterException: IAM role's policy must include the `ec2:DescribeSubnets` action
			if isAWSErr(err, eks.ErrCodeInvalidParameterException, "IAM role's policy must include") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.CreateCluster(input)
	}
	if err != nil {
		return fmt.Errorf("error creating EKS Cluster (%s): %s", name, err)
	}

	d.SetId(name)

	stateConf := resource.StateChangeConf{
		Pending: []string{eks.ClusterStatusCreating},
		Target:  []string{eks.ClusterStatusActive},
		Timeout: d.Timeout(schema.TimeoutCreate),
		Refresh: refreshEksClusterStatus(conn, name),
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}

	return resourceAwsEksClusterRead(d, meta)
}

func resourceAwsEksClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).eksconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &eks.DescribeClusterInput{
		Name: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading EKS Cluster: %s", input)
	output, err := conn.DescribeCluster(input)
	if err != nil {
		if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] EKS Cluster (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading EKS Cluster (%s): %s", d.Id(), err)
	}

	cluster := output.Cluster
	if cluster == nil {
		log.Printf("[WARN] EKS Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", cluster.Arn)

	if err := d.Set("certificate_authority", flattenEksCertificate(cluster.CertificateAuthority)); err != nil {
		return fmt.Errorf("error setting certificate_authority: %s", err)
	}

	d.Set("created_at", aws.TimeValue(cluster.CreatedAt).String())

	if err := d.Set("encryption_config", flattenEksEncryptionConfig(cluster.EncryptionConfig)); err != nil {
		return fmt.Errorf("error setting encryption_config: %s", err)
	}

	d.Set("endpoint", cluster.Endpoint)

	if err := d.Set("identity", flattenEksIdentity(cluster.Identity)); err != nil {
		return fmt.Errorf("error setting identity: %s", err)
	}

	d.Set("name", cluster.Name)
	d.Set("platform_version", cluster.PlatformVersion)
	d.Set("role_arn", cluster.RoleArn)
	d.Set("status", cluster.Status)

	if err := d.Set("tags", keyvaluetags.EksKeyValueTags(cluster.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("version", cluster.Version)
	if err := d.Set("enabled_cluster_log_types", flattenEksEnabledLogTypes(cluster.Logging)); err != nil {
		return fmt.Errorf("error setting enabled_cluster_log_types: %s", err)
	}

	if err := d.Set("vpc_config", flattenEksVpcConfigResponse(cluster.ResourcesVpcConfig)); err != nil {
		return fmt.Errorf("error setting vpc_config: %s", err)
	}

	return nil
}

func resourceAwsEksClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).eksconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.EksUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	if d.HasChange("version") {
		input := &eks.UpdateClusterVersionInput{
			Name:    aws.String(d.Id()),
			Version: aws.String(d.Get("version").(string)),
		}

		log.Printf("[DEBUG] Updating EKS Cluster (%s) version: %s", d.Id(), input)
		output, err := conn.UpdateClusterVersion(input)

		if err != nil {
			return fmt.Errorf("error updating EKS Cluster (%s) version: %s", d.Id(), err)
		}

		if output == nil || output.Update == nil || output.Update.Id == nil {
			return fmt.Errorf("error determining EKS Cluster (%s) version update ID: empty response", d.Id())
		}

		updateID := aws.StringValue(output.Update.Id)

		err = waitForUpdateEksCluster(conn, d.Id(), updateID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return fmt.Errorf("error waiting for EKS Cluster (%s) version update (%s): %s", d.Id(), updateID, err)
		}
	}

	if d.HasChange("enabled_cluster_log_types") {
		_, v := d.GetChange("enabled_cluster_log_types")
		input := &eks.UpdateClusterConfigInput{
			Name:    aws.String(d.Id()),
			Logging: expandEksLoggingTypes(v.(*schema.Set)),
		}

		log.Printf("[DEBUG] Updating EKS Cluster (%s) logging: %s", d.Id(), input)
		output, err := conn.UpdateClusterConfig(input)

		if err != nil {
			return fmt.Errorf("error updating EKS Cluster (%s) logging: %s", d.Id(), err)
		}

		if output == nil || output.Update == nil || output.Update.Id == nil {
			return fmt.Errorf("error determining EKS Cluster (%s) logging update ID: empty response", d.Id())
		}

		updateID := aws.StringValue(output.Update.Id)

		err = waitForUpdateEksCluster(conn, d.Id(), updateID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return fmt.Errorf("error waiting for EKS Cluster (%s) logging update (%s): %s", d.Id(), updateID, err)
		}
	}

	if d.HasChanges("vpc_config.0.endpoint_private_access", "vpc_config.0.endpoint_public_access", "vpc_config.0.public_access_cidrs") {
		input := &eks.UpdateClusterConfigInput{
			Name:               aws.String(d.Id()),
			ResourcesVpcConfig: expandEksVpcConfigUpdateRequest(d.Get("vpc_config").([]interface{})),
		}

		log.Printf("[DEBUG] Updating EKS Cluster (%s) config: %s", d.Id(), input)
		output, err := conn.UpdateClusterConfig(input)

		if err != nil {
			return fmt.Errorf("error updating EKS Cluster (%s) config: %s", d.Id(), err)
		}

		if output == nil || output.Update == nil || output.Update.Id == nil {
			return fmt.Errorf("error determining EKS Cluster (%s) config update ID: empty response", d.Id())
		}

		updateID := aws.StringValue(output.Update.Id)

		err = waitForUpdateEksCluster(conn, d.Id(), updateID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return fmt.Errorf("error waiting for EKS Cluster (%s) config update (%s): %s", d.Id(), updateID, err)
		}
	}

	return resourceAwsEksClusterRead(d, meta)
}

func resourceAwsEksClusterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).eksconn

	log.Printf("[DEBUG] Deleting EKS Cluster: %s", d.Id())
	err := deleteEksCluster(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error deleting EKS Cluster (%s): %s", d.Id(), err)
	}

	err = waitForDeleteEksCluster(conn, d.Id(), d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return fmt.Errorf("error waiting for EKS Cluster (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func deleteEksCluster(conn *eks.EKS, clusterName string) error {
	input := &eks.DeleteClusterInput{
		Name: aws.String(clusterName),
	}

	_, err := conn.DeleteCluster(input)
	if err != nil {
		if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		// Sometimes the EKS API returns the ResourceNotFound error in this form:
		// ClientException: No cluster found for name: tf-acc-test-0o1f8
		if isAWSErr(err, eks.ErrCodeClientException, "No cluster found for name:") {
			return nil
		}
		return err
	}

	return nil
}

func expandEksEncryptionConfig(tfList []interface{}) []*eks.EncryptionConfig {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*eks.EncryptionConfig

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := &eks.EncryptionConfig{
			Provider: expandEksProvider(tfMap["provider"].([]interface{})),
		}

		if v, ok := tfMap["resources"].(*schema.Set); ok && v.Len() > 0 {
			apiObject.Resources = expandStringSet(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandEksProvider(tfList []interface{}) *eks.Provider {
	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	apiObject := &eks.Provider{}

	if v, ok := tfMap["key_arn"].(string); ok && v != "" {
		apiObject.KeyArn = aws.String(v)
	}

	return apiObject
}

func expandEksVpcConfigRequest(l []interface{}) *eks.VpcConfigRequest {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	vpcConfigRequest := &eks.VpcConfigRequest{
		EndpointPrivateAccess: aws.Bool(m["endpoint_private_access"].(bool)),
		EndpointPublicAccess:  aws.Bool(m["endpoint_public_access"].(bool)),
		SecurityGroupIds:      expandStringSet(m["security_group_ids"].(*schema.Set)),
		SubnetIds:             expandStringSet(m["subnet_ids"].(*schema.Set)),
	}

	if v, ok := m["public_access_cidrs"].(*schema.Set); ok && v.Len() > 0 {
		vpcConfigRequest.PublicAccessCidrs = expandStringSet(v)
	}

	return vpcConfigRequest
}

func expandEksVpcConfigUpdateRequest(l []interface{}) *eks.VpcConfigRequest {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	vpcConfigRequest := &eks.VpcConfigRequest{
		EndpointPrivateAccess: aws.Bool(m["endpoint_private_access"].(bool)),
		EndpointPublicAccess:  aws.Bool(m["endpoint_public_access"].(bool)),
	}

	if v, ok := m["public_access_cidrs"].(*schema.Set); ok && v.Len() > 0 {
		vpcConfigRequest.PublicAccessCidrs = expandStringSet(v)
	}

	return vpcConfigRequest
}

func expandEksLoggingTypes(vEnabledLogTypes *schema.Set) *eks.Logging {
	vEksLogTypes := []interface{}{}
	for _, eksLogType := range eksLogTypes {
		vEksLogTypes = append(vEksLogTypes, eksLogType)
	}
	vAllLogTypes := schema.NewSet(schema.HashString, vEksLogTypes)

	return &eks.Logging{
		ClusterLogging: []*eks.LogSetup{
			{
				Enabled: aws.Bool(true),
				Types:   expandStringSet(vEnabledLogTypes),
			},
			{
				Enabled: aws.Bool(false),
				Types:   expandStringSet(vAllLogTypes.Difference(vEnabledLogTypes)),
			},
		},
	}
}

func flattenEksCertificate(certificate *eks.Certificate) []map[string]interface{} {
	if certificate == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"data": aws.StringValue(certificate.Data),
	}

	return []map[string]interface{}{m}
}

func flattenEksIdentity(identity *eks.Identity) []map[string]interface{} {
	if identity == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"oidc": flattenEksOidc(identity.Oidc),
	}

	return []map[string]interface{}{m}
}

func flattenEksOidc(oidc *eks.OIDC) []map[string]interface{} {
	if oidc == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"issuer": aws.StringValue(oidc.Issuer),
	}

	return []map[string]interface{}{m}
}

func flattenEksEncryptionConfig(apiObjects []*eks.EncryptionConfig) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"provider":  flattenEksProvider(apiObject.Provider),
			"resources": aws.StringValueSlice(apiObject.Resources),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenEksProvider(apiObject *eks.Provider) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"key_arn": aws.StringValue(apiObject.KeyArn),
	}

	return []interface{}{tfMap}
}

func flattenEksVpcConfigResponse(vpcConfig *eks.VpcConfigResponse) []map[string]interface{} {
	if vpcConfig == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"cluster_security_group_id": aws.StringValue(vpcConfig.ClusterSecurityGroupId),
		"endpoint_private_access":   aws.BoolValue(vpcConfig.EndpointPrivateAccess),
		"endpoint_public_access":    aws.BoolValue(vpcConfig.EndpointPublicAccess),
		"security_group_ids":        schema.NewSet(schema.HashString, flattenStringList(vpcConfig.SecurityGroupIds)),
		"subnet_ids":                schema.NewSet(schema.HashString, flattenStringList(vpcConfig.SubnetIds)),
		"public_access_cidrs":       schema.NewSet(schema.HashString, flattenStringList(vpcConfig.PublicAccessCidrs)),
		"vpc_id":                    aws.StringValue(vpcConfig.VpcId),
	}

	return []map[string]interface{}{m}
}

func flattenEksEnabledLogTypes(logging *eks.Logging) *schema.Set {
	enabledLogTypes := []*string{}

	if logging != nil {
		logSetups := logging.ClusterLogging
		for _, logSetup := range logSetups {
			if logSetup == nil || !aws.BoolValue(logSetup.Enabled) {
				continue
			}

			enabledLogTypes = append(enabledLogTypes, logSetup.Types...)
		}
	}

	return flattenStringSet(enabledLogTypes)
}

func refreshEksClusterStatus(conn *eks.EKS, clusterName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.DescribeCluster(&eks.DescribeClusterInput{
			Name: aws.String(clusterName),
		})
		if err != nil {
			return 42, "", err
		}
		cluster := output.Cluster
		if cluster == nil {
			return cluster, "", fmt.Errorf("EKS Cluster (%s) missing", clusterName)
		}
		return cluster, aws.StringValue(cluster.Status), nil
	}
}

func refreshEksUpdateStatus(conn *eks.EKS, clusterName, updateID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &eks.DescribeUpdateInput{
			Name:     aws.String(clusterName),
			UpdateId: aws.String(updateID),
		}

		output, err := conn.DescribeUpdate(input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.Update == nil {
			return nil, "", fmt.Errorf("EKS Cluster (%s) update (%s) missing", clusterName, updateID)
		}

		return output.Update, aws.StringValue(output.Update.Status), nil
	}
}

func waitForDeleteEksCluster(conn *eks.EKS, clusterName string, timeout time.Duration) error {
	stateConf := resource.StateChangeConf{
		Pending: []string{
			eks.ClusterStatusActive,
			eks.ClusterStatusDeleting,
		},
		Target:  []string{""},
		Timeout: timeout,
		Refresh: refreshEksClusterStatus(conn, clusterName),
	}
	cluster, err := stateConf.WaitForState()
	if err != nil {
		if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		// Sometimes the EKS API returns the ResourceNotFound error in this form:
		// ClientException: No cluster found for name: tf-acc-test-0o1f8
		if isAWSErr(err, eks.ErrCodeClientException, "No cluster found for name:") {
			return nil
		}
	}
	if cluster == nil {
		return nil
	}
	return err
}

func waitForUpdateEksCluster(conn *eks.EKS, clusterName, updateID string, timeout time.Duration) error {
	stateConf := resource.StateChangeConf{
		Pending: []string{eks.UpdateStatusInProgress},
		Target: []string{
			eks.UpdateStatusCancelled,
			eks.UpdateStatusFailed,
			eks.UpdateStatusSuccessful,
		},
		Timeout: timeout,
		Refresh: refreshEksUpdateStatus(conn, clusterName, updateID),
	}
	updateRaw, err := stateConf.WaitForState()

	if err != nil {
		return err
	}

	update := updateRaw.(*eks.Update)

	if aws.StringValue(update.Status) == eks.UpdateStatusSuccessful {
		return nil
	}

	var detailedErrors []string
	for i, updateError := range update.Errors {
		detailedErrors = append(detailedErrors, fmt.Sprintf("Error %d: Code: %s / Message: %s", i+1, aws.StringValue(updateError.ErrorCode), aws.StringValue(updateError.ErrorMessage)))
	}

	return fmt.Errorf("EKS Cluster (%s) update (%s) status (%s) not successful: Errors:\n%s", clusterName, updateID, aws.StringValue(update.Status), strings.Join(detailedErrors, "\n"))
}
