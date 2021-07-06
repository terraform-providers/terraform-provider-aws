package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	tfeks "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/eks"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/eks/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/eks/waiter"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsEksCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEksClusterCreate,
		Read:   resourceAwsEksClusterRead,
		Update: resourceAwsEksClusterUpdate,
		Delete: resourceAwsEksClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: customdiff.Sequence(
			SetTagsDiff,
			customdiff.ForceNewIfChange("encryption_config", func(_ context.Context, old, new, meta interface{}) bool {
				// You cannot disable envelope encryption after enabling it. This action is irreversible.
				return len(old.([]interface{})) == 1 && len(new.([]interface{})) == 0
			}),
		),

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
			"enabled_cluster_log_types": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(eks.LogType_Values(), true),
				},
				Set: schema.HashString,
			},
			"encryption_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"provider": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key_arn": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"resources": {
							Type:     schema.TypeSet,
							MinItems: 1,
							Required: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(tfeks.Resources_Values(), false),
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
			"kubernetes_network_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service_ipv4_cidr": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
							ValidateFunc: validation.All(
								validation.IsCIDRNetwork(12, 24),
								validation.StringMatch(regexp.MustCompile(`^(10|172\.(1[6-9]|2[0-9]|3[0-1])|192\.168)\..*`), "must be within 10.0.0.0/8, 172.16.0.0/12, or 192.168.0.0/16"),
							),
						},
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateEKSClusterName,
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
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
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
						"public_access_cidrs": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validateCIDRNetworkAddress,
							},
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
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsEksClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).eksconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))
	name := d.Get("name").(string)

	input := &eks.CreateClusterInput{
		EncryptionConfig:   expandEksEncryptionConfig(d.Get("encryption_config").([]interface{})),
		Logging:            expandEksLoggingTypes(d.Get("enabled_cluster_log_types").(*schema.Set)),
		Name:               aws.String(name),
		ResourcesVpcConfig: expandEksVpcConfigRequest(d.Get("vpc_config").([]interface{})),
		RoleArn:            aws.String(d.Get("role_arn").(string)),
	}

	if _, ok := d.GetOk("kubernetes_network_config"); ok {
		input.KubernetesNetworkConfig = expandEksNetworkConfigRequest(d.Get("kubernetes_network_config").([]interface{}))
	}

	if v, ok := d.GetOk("version"); ok {
		input.Version = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().EksTags()
	}

	log.Printf("[DEBUG] Creating EKS Cluster: %s", input)
	var output *eks.CreateClusterOutput
	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		var err error

		output, err = conn.CreateCluster(input)

		// InvalidParameterException: roleArn, arn:aws:iam::123456789012:role/XXX, does not exist
		if tfawserr.ErrMessageContains(err, eks.ErrCodeInvalidParameterException, "does not exist") {
			return resource.RetryableError(err)
		}

		// InvalidParameterException: Error in role params
		if tfawserr.ErrMessageContains(err, eks.ErrCodeInvalidParameterException, "Error in role params") {
			return resource.RetryableError(err)
		}

		if tfawserr.ErrMessageContains(err, eks.ErrCodeInvalidParameterException, "Role could not be assumed because the trusted entity is not correct") {
			return resource.RetryableError(err)
		}

		// InvalidParameterException: The provided role doesn't have the Amazon EKS Managed Policies associated with it. Please ensure the following policy is attached: arn:aws:iam::aws:policy/AmazonEKSClusterPolicy
		if tfawserr.ErrMessageContains(err, eks.ErrCodeInvalidParameterException, "The provided role doesn't have the Amazon EKS Managed Policies associated with it") {
			return resource.RetryableError(err)
		}

		// InvalidParameterException: IAM role's policy must include the `ec2:DescribeSubnets` action
		if tfawserr.ErrMessageContains(err, eks.ErrCodeInvalidParameterException, "IAM role's policy must include") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateCluster(input)
	}

	if err != nil {
		return fmt.Errorf("error creating EKS Cluster (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.Cluster.Name))

	_, err = waiter.ClusterCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("error waiting for EKS Cluster (%s) to create: %w", d.Id(), err)
	}

	return resourceAwsEksClusterRead(d, meta)
}

func resourceAwsEksClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).eksconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	cluster, err := finder.ClusterByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EKS Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EKS Cluster (%s): %w", d.Id(), err)
	}

	d.Set("arn", cluster.Arn)

	if err := d.Set("certificate_authority", flattenEksCertificate(cluster.CertificateAuthority)); err != nil {
		return fmt.Errorf("error setting certificate_authority: %w", err)
	}

	d.Set("created_at", aws.TimeValue(cluster.CreatedAt).String())

	if err := d.Set("enabled_cluster_log_types", flattenEksEnabledLogTypes(cluster.Logging)); err != nil {
		return fmt.Errorf("error setting enabled_cluster_log_types: %w", err)
	}

	if err := d.Set("encryption_config", flattenEksEncryptionConfig(cluster.EncryptionConfig)); err != nil {
		return fmt.Errorf("error setting encryption_config: %w", err)
	}

	d.Set("endpoint", cluster.Endpoint)

	if err := d.Set("identity", flattenEksIdentity(cluster.Identity)); err != nil {
		return fmt.Errorf("error setting identity: %w", err)
	}

	if err := d.Set("kubernetes_network_config", flattenEksNetworkConfig(cluster.KubernetesNetworkConfig)); err != nil {
		return fmt.Errorf("error setting kubernetes_network_config: %w", err)
	}

	d.Set("name", cluster.Name)
	d.Set("platform_version", cluster.PlatformVersion)
	d.Set("role_arn", cluster.RoleArn)
	d.Set("status", cluster.Status)
	d.Set("version", cluster.Version)

	if err := d.Set("vpc_config", flattenEksVpcConfigResponse(cluster.ResourcesVpcConfig)); err != nil {
		return fmt.Errorf("error setting vpc_config: %w", err)
	}

	tags := keyvaluetags.EksKeyValueTags(cluster.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsEksClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).eksconn

	// Do any version update first.
	if d.HasChange("version") {
		input := &eks.UpdateClusterVersionInput{
			Name:    aws.String(d.Id()),
			Version: aws.String(d.Get("version").(string)),
		}

		log.Printf("[DEBUG] Updating EKS Cluster (%s) version: %s", d.Id(), input)
		output, err := conn.UpdateClusterVersion(input)

		if err != nil {
			return fmt.Errorf("error updating EKS Cluster (%s) version: %w", d.Id(), err)
		}

		updateID := aws.StringValue(output.Update.Id)

		_, err = waiter.ClusterUpdateSuccessful(conn, d.Id(), updateID, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return fmt.Errorf("error waiting for EKS Cluster (%s) version update (%s): %w", d.Id(), updateID, err)
		}
	}

	if d.HasChange("encryption_config") {
		input := &eks.AssociateEncryptionConfigInput{
			ClusterName:      aws.String(d.Id()),
			EncryptionConfig: expandEksEncryptionConfig(d.Get("encryption_config").([]interface{})),
		}

		log.Printf("[DEBUG] Associating EKS Cluster (%s) encryption config: %s", d.Id(), input)
		output, err := conn.AssociateEncryptionConfig(input)

		if err != nil {
			return fmt.Errorf("error associating EKS Cluster (%s) encryption config: %w", d.Id(), err)
		}

		updateID := aws.StringValue(output.Update.Id)

		_, err = waiter.ClusterUpdateSuccessful(conn, d.Id(), updateID, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return fmt.Errorf("error waiting for EKS Cluster (%s) encryption config association (%s): %w", d.Id(), updateID, err)
		}
	}

	if d.HasChange("enabled_cluster_log_types") {
		input := &eks.UpdateClusterConfigInput{
			Logging: expandEksLoggingTypes(d.Get("enabled_cluster_log_types").(*schema.Set)),
			Name:    aws.String(d.Id()),
		}

		log.Printf("[DEBUG] Updating EKS Cluster (%s) logging: %s", d.Id(), input)
		output, err := conn.UpdateClusterConfig(input)

		if err != nil {
			return fmt.Errorf("error updating EKS Cluster (%s) logging: %w", d.Id(), err)
		}

		updateID := aws.StringValue(output.Update.Id)

		_, err = waiter.ClusterUpdateSuccessful(conn, d.Id(), updateID, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return fmt.Errorf("error waiting for EKS Cluster (%s) logging update (%s): %w", d.Id(), updateID, err)
		}
	}

	if d.HasChanges("vpc_config.0.endpoint_private_access", "vpc_config.0.endpoint_public_access", "vpc_config.0.public_access_cidrs") {
		input := &eks.UpdateClusterConfigInput{
			Name:               aws.String(d.Id()),
			ResourcesVpcConfig: expandEksVpcConfigUpdateRequest(d.Get("vpc_config").([]interface{})),
		}

		log.Printf("[DEBUG] Updating EKS Cluster (%s) VPC config: %s", d.Id(), input)
		output, err := conn.UpdateClusterConfig(input)

		if err != nil {
			return fmt.Errorf("error updating EKS Cluster (%s) VPC config: %w", d.Id(), err)
		}

		updateID := aws.StringValue(output.Update.Id)

		_, err = waiter.ClusterUpdateSuccessful(conn, d.Id(), updateID, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return fmt.Errorf("error waiting for EKS Cluster (%s) VPC config update (%s): %w", d.Id(), updateID, err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.EksUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceAwsEksClusterRead(d, meta)
}

func resourceAwsEksClusterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).eksconn

	log.Printf("[DEBUG] Deleting EKS Cluster: %s", d.Id())
	_, err := conn.DeleteCluster(&eks.DeleteClusterInput{
		Name: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) {
		return nil
	}

	// Sometimes the EKS API returns the ResourceNotFound error in this form:
	// ClientException: No cluster found for name: tf-acc-test-0o1f8
	if tfawserr.ErrMessageContains(err, eks.ErrCodeClientException, "No cluster found for name:") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EKS Cluster (%s): %w", d.Id(), err)
	}

	_, err = waiter.ClusterDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return fmt.Errorf("error waiting for EKS Cluster (%s) to delete: %w", d.Id(), err)
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

func expandEksNetworkConfigRequest(tfList []interface{}) *eks.KubernetesNetworkConfigRequest {
	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	apiObject := &eks.KubernetesNetworkConfigRequest{}

	if v, ok := tfMap["service_ipv4_cidr"].(string); ok && v != "" {
		apiObject.ServiceIpv4Cidr = aws.String(v)
	}

	return apiObject
}

func expandEksLoggingTypes(vEnabledLogTypes *schema.Set) *eks.Logging {
	vEksLogTypes := []interface{}{}
	for _, eksLogType := range eks.LogType_Values() {
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
		"security_group_ids":        flattenStringSet(vpcConfig.SecurityGroupIds),
		"subnet_ids":                flattenStringSet(vpcConfig.SubnetIds),
		"public_access_cidrs":       flattenStringSet(vpcConfig.PublicAccessCidrs),
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

func flattenEksNetworkConfig(apiObject *eks.KubernetesNetworkConfigResponse) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"service_ipv4_cidr": aws.StringValue(apiObject.ServiceIpv4Cidr),
	}

	return []interface{}{tfMap}
}
