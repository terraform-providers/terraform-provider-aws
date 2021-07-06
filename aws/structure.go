package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/aws/aws-sdk-go/service/dax"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/macie"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/worklink"
	"github.com/beevik/etree"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/mitchellh/copystructure"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
	"gopkg.in/yaml.v2"
)

// Takes the result of flatmap.Expand for an array of listeners and
// returns ELB API compatible objects
func expandListeners(configured []interface{}) ([]*elb.Listener, error) {
	listeners := make([]*elb.Listener, 0, len(configured))

	// Loop over our configured listeners and create
	// an array of aws-sdk-go compatible objects
	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})

		ip := int64(data["instance_port"].(int))
		lp := int64(data["lb_port"].(int))
		l := &elb.Listener{
			InstancePort:     &ip,
			InstanceProtocol: aws.String(data["instance_protocol"].(string)),
			LoadBalancerPort: &lp,
			Protocol:         aws.String(data["lb_protocol"].(string)),
		}

		if v, ok := data["ssl_certificate_id"]; ok {
			l.SSLCertificateId = aws.String(v.(string))
		}

		var valid bool
		if l.SSLCertificateId != nil && *l.SSLCertificateId != "" {
			// validate the protocol is correct
			for _, p := range []string{"https", "ssl"} {
				if (strings.ToLower(*l.InstanceProtocol) == p) || (strings.ToLower(*l.Protocol) == p) {
					valid = true
				}
			}
		} else {
			valid = true
		}

		if valid {
			listeners = append(listeners, l)
		} else {
			return nil, fmt.Errorf("ELB Listener: ssl_certificate_id may be set only when protocol is 'https' or 'ssl'")
		}
	}

	return listeners, nil
}

// Takes the result of flatmap. Expand for an array of load balancers and
// returns ecs.LoadBalancer compatible objects
func expandEcsLoadBalancers(configured []interface{}) []*ecs.LoadBalancer {
	loadBalancers := make([]*ecs.LoadBalancer, 0, len(configured))

	// Loop over our configured load balancers and create
	// an array of aws-sdk-go compatible objects
	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})

		l := &ecs.LoadBalancer{
			ContainerName: aws.String(data["container_name"].(string)),
			ContainerPort: aws.Int64(int64(data["container_port"].(int))),
		}

		if v, ok := data["elb_name"]; ok && v.(string) != "" {
			l.LoadBalancerName = aws.String(v.(string))
		}
		if v, ok := data["target_group_arn"]; ok && v.(string) != "" {
			l.TargetGroupArn = aws.String(v.(string))
		}

		loadBalancers = append(loadBalancers, l)
	}

	return loadBalancers
}

// Takes the result of flatmap.Expand for an array of ingress/egress security
// group rules and returns EC2 API compatible objects. This function will error
// if it finds invalid permissions input, namely a protocol of "-1" with either
// to_port or from_port set to a non-zero value.
func expandIPPerms(
	group *ec2.SecurityGroup, configured []interface{}) ([]*ec2.IpPermission, error) {
	vpc := group.VpcId != nil && *group.VpcId != ""

	perms := make([]*ec2.IpPermission, len(configured))
	for i, mRaw := range configured {
		var perm ec2.IpPermission
		m := mRaw.(map[string]interface{})

		perm.FromPort = aws.Int64(int64(m["from_port"].(int)))
		perm.ToPort = aws.Int64(int64(m["to_port"].(int)))
		perm.IpProtocol = aws.String(m["protocol"].(string))

		// When protocol is "-1", AWS won't store any ports for the
		// rule, but also won't error if the user specifies ports other
		// than '0'. Force the user to make a deliberate '0' port
		// choice when specifying a "-1" protocol, and tell them about
		// AWS's behavior in the error message.
		if *perm.IpProtocol == "-1" && (*perm.FromPort != 0 || *perm.ToPort != 0) {
			return nil, fmt.Errorf(
				"from_port (%d) and to_port (%d) must both be 0 to use the 'ALL' \"-1\" protocol!",
				*perm.FromPort, *perm.ToPort)
		}

		var groups []string
		if raw, ok := m["security_groups"]; ok {
			list := raw.(*schema.Set).List()
			for _, v := range list {
				groups = append(groups, v.(string))
			}
		}
		if v, ok := m["self"]; ok && v.(bool) {
			if vpc {
				groups = append(groups, *group.GroupId)
			} else {
				groups = append(groups, *group.GroupName)
			}
		}

		if len(groups) > 0 {
			perm.UserIdGroupPairs = make([]*ec2.UserIdGroupPair, len(groups))
			for i, name := range groups {
				ownerId, id := "", name
				if items := strings.Split(id, "/"); len(items) > 1 {
					ownerId, id = items[0], items[1]
				}

				perm.UserIdGroupPairs[i] = &ec2.UserIdGroupPair{
					GroupId: aws.String(id),
				}

				if ownerId != "" {
					perm.UserIdGroupPairs[i].UserId = aws.String(ownerId)
				}

				if !vpc {
					perm.UserIdGroupPairs[i].GroupId = nil
					perm.UserIdGroupPairs[i].GroupName = aws.String(id)
				}
			}
		}

		if raw, ok := m["cidr_blocks"]; ok {
			list := raw.([]interface{})
			for _, v := range list {
				perm.IpRanges = append(perm.IpRanges, &ec2.IpRange{CidrIp: aws.String(v.(string))})
			}
		}
		if raw, ok := m["ipv6_cidr_blocks"]; ok {
			list := raw.([]interface{})
			for _, v := range list {
				perm.Ipv6Ranges = append(perm.Ipv6Ranges, &ec2.Ipv6Range{CidrIpv6: aws.String(v.(string))})
			}
		}

		if raw, ok := m["prefix_list_ids"]; ok {
			list := raw.([]interface{})
			for _, v := range list {
				perm.PrefixListIds = append(perm.PrefixListIds, &ec2.PrefixListId{PrefixListId: aws.String(v.(string))})
			}
		}

		if raw, ok := m["description"]; ok {
			description := raw.(string)
			if description != "" {
				for _, v := range perm.IpRanges {
					v.Description = aws.String(description)
				}
				for _, v := range perm.Ipv6Ranges {
					v.Description = aws.String(description)
				}
				for _, v := range perm.PrefixListIds {
					v.Description = aws.String(description)
				}
				for _, v := range perm.UserIdGroupPairs {
					v.Description = aws.String(description)
				}
			}
		}

		perms[i] = &perm
	}

	return perms, nil
}

// Takes the result of flatmap.Expand for an array of parameters and
// returns Parameter API compatible objects
func expandParameters(configured []interface{}) []*rds.Parameter {
	var parameters []*rds.Parameter

	// Loop over our configured parameters and create
	// an array of aws-sdk-go compatible objects
	for _, pRaw := range configured {
		data := pRaw.(map[string]interface{})

		if data["name"].(string) == "" {
			continue
		}

		p := &rds.Parameter{
			ApplyMethod:    aws.String(data["apply_method"].(string)),
			ParameterName:  aws.String(data["name"].(string)),
			ParameterValue: aws.String(data["value"].(string)),
		}

		parameters = append(parameters, p)
	}

	return parameters
}

func expandRedshiftParameters(configured []interface{}) []*redshift.Parameter {
	var parameters []*redshift.Parameter

	// Loop over our configured parameters and create
	// an array of aws-sdk-go compatible objects
	for _, pRaw := range configured {
		data := pRaw.(map[string]interface{})

		if data["name"].(string) == "" {
			continue
		}

		p := &redshift.Parameter{
			ParameterName:  aws.String(data["name"].(string)),
			ParameterValue: aws.String(data["value"].(string)),
		}

		parameters = append(parameters, p)
	}

	return parameters
}

// Takes the result of flatmap.Expand for an array of parameters and
// returns Parameter API compatible objects
func expandDocDBParameters(configured []interface{}) []*docdb.Parameter {
	parameters := make([]*docdb.Parameter, 0, len(configured))

	// Loop over our configured parameters and create
	// an array of aws-sdk-go compatible objects
	for _, pRaw := range configured {
		data := pRaw.(map[string]interface{})

		p := &docdb.Parameter{
			ApplyMethod:    aws.String(data["apply_method"].(string)),
			ParameterName:  aws.String(data["name"].(string)),
			ParameterValue: aws.String(data["value"].(string)),
		}

		parameters = append(parameters, p)
	}

	return parameters
}

func expandOptionConfiguration(configured []interface{}) []*rds.OptionConfiguration {
	var option []*rds.OptionConfiguration

	for _, pRaw := range configured {
		data := pRaw.(map[string]interface{})

		o := &rds.OptionConfiguration{
			OptionName: aws.String(data["option_name"].(string)),
		}

		if raw, ok := data["port"]; ok {
			port := raw.(int)
			if port != 0 {
				o.Port = aws.Int64(int64(port))
			}
		}

		if raw, ok := data["db_security_group_memberships"]; ok {
			memberships := expandStringSet(raw.(*schema.Set))
			if len(memberships) > 0 {
				o.DBSecurityGroupMemberships = memberships
			}
		}

		if raw, ok := data["vpc_security_group_memberships"]; ok {
			memberships := expandStringSet(raw.(*schema.Set))
			if len(memberships) > 0 {
				o.VpcSecurityGroupMemberships = memberships
			}
		}

		if raw, ok := data["option_settings"]; ok {
			o.OptionSettings = expandOptionSetting(raw.(*schema.Set).List())
		}

		if raw, ok := data["version"]; ok && raw.(string) != "" {
			o.OptionVersion = aws.String(raw.(string))
		}

		option = append(option, o)
	}

	return option
}

func expandOptionSetting(list []interface{}) []*rds.OptionSetting {
	options := make([]*rds.OptionSetting, 0, len(list))

	for _, oRaw := range list {
		data := oRaw.(map[string]interface{})

		o := &rds.OptionSetting{
			Name:  aws.String(data["name"].(string)),
			Value: aws.String(data["value"].(string)),
		}

		options = append(options, o)
	}

	return options
}

// Takes the result of flatmap.Expand for an array of parameters and
// returns Parameter API compatible objects
func expandNeptuneParameters(configured []interface{}) []*neptune.Parameter {
	parameters := make([]*neptune.Parameter, 0, len(configured))

	// Loop over our configured parameters and create
	// an array of aws-sdk-go compatible objects
	for _, pRaw := range configured {
		data := pRaw.(map[string]interface{})

		p := &neptune.Parameter{
			ApplyMethod:    aws.String(data["apply_method"].(string)),
			ParameterName:  aws.String(data["name"].(string)),
			ParameterValue: aws.String(data["value"].(string)),
		}

		parameters = append(parameters, p)
	}

	return parameters
}

// Flattens an access log into something that flatmap.Flatten() can handle
func flattenAccessLog(l *elb.AccessLog) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)

	if l == nil {
		return nil
	}

	r := make(map[string]interface{})
	if l.S3BucketName != nil {
		r["bucket"] = *l.S3BucketName
	}

	if l.S3BucketPrefix != nil {
		r["bucket_prefix"] = *l.S3BucketPrefix
	}

	if l.EmitInterval != nil {
		r["interval"] = *l.EmitInterval
	}

	if l.Enabled != nil {
		r["enabled"] = *l.Enabled
	}

	result = append(result, r)

	return result
}

// Takes the result of flatmap.Expand for an array of step adjustments and
// returns a []*autoscaling.StepAdjustment.
func expandStepAdjustments(configured []interface{}) ([]*autoscaling.StepAdjustment, error) {
	var adjustments []*autoscaling.StepAdjustment

	// Loop over our configured step adjustments and create an array
	// of aws-sdk-go compatible objects. We're forced to convert strings
	// to floats here because there's no way to detect whether or not
	// an uninitialized, optional schema element is "0.0" deliberately.
	// With strings, we can test for "", which is definitely an empty
	// struct value.
	for _, raw := range configured {
		data := raw.(map[string]interface{})
		a := &autoscaling.StepAdjustment{
			ScalingAdjustment: aws.Int64(int64(data["scaling_adjustment"].(int))),
		}
		if data["metric_interval_lower_bound"] != "" {
			bound := data["metric_interval_lower_bound"]
			switch bound := bound.(type) {
			case string:
				f, err := strconv.ParseFloat(bound, 64)
				if err != nil {
					return nil, fmt.Errorf(
						"metric_interval_lower_bound must be a float value represented as a string")
				}
				a.MetricIntervalLowerBound = aws.Float64(f)
			default:
				return nil, fmt.Errorf(
					"metric_interval_lower_bound isn't a string. This is a bug. Please file an issue.")
			}
		}
		if data["metric_interval_upper_bound"] != "" {
			bound := data["metric_interval_upper_bound"]
			switch bound := bound.(type) {
			case string:
				f, err := strconv.ParseFloat(bound, 64)
				if err != nil {
					return nil, fmt.Errorf(
						"metric_interval_upper_bound must be a float value represented as a string")
				}
				a.MetricIntervalUpperBound = aws.Float64(f)
			default:
				return nil, fmt.Errorf(
					"metric_interval_upper_bound isn't a string. This is a bug. Please file an issue.")
			}
		}
		adjustments = append(adjustments, a)
	}

	return adjustments, nil
}

// Flattens a health check into something that flatmap.Flatten()
// can handle
func flattenHealthCheck(check *elb.HealthCheck) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)

	chk := make(map[string]interface{})
	chk["unhealthy_threshold"] = *check.UnhealthyThreshold
	chk["healthy_threshold"] = *check.HealthyThreshold
	chk["target"] = *check.Target
	chk["timeout"] = *check.Timeout
	chk["interval"] = *check.Interval

	result = append(result, chk)

	return result
}

// Flattens an array of UserSecurityGroups into a []*GroupIdentifier
func flattenSecurityGroups(list []*ec2.UserIdGroupPair, ownerId *string) []*GroupIdentifier {
	result := make([]*GroupIdentifier, 0, len(list))
	for _, g := range list {
		var userId *string
		if g.UserId != nil && *g.UserId != "" && (ownerId == nil || *ownerId != *g.UserId) {
			userId = g.UserId
		}
		// userid nil here for same vpc groups

		vpc := g.GroupName == nil || *g.GroupName == ""
		var id *string
		if vpc {
			id = g.GroupId
		} else {
			id = g.GroupName
		}

		// id is groupid for vpcs
		// id is groupname for non vpc (classic)

		if userId != nil {
			id = aws.String(*userId + "/" + *id)
		}

		if vpc {
			result = append(result, &GroupIdentifier{
				GroupId:     id,
				Description: g.Description,
			})
		} else {
			result = append(result, &GroupIdentifier{
				GroupId:     g.GroupId,
				GroupName:   id,
				Description: g.Description,
			})
		}
	}
	return result
}

// Flattens an array of Instances into a []string
func flattenInstances(list []*elb.Instance) []string {
	result := make([]string, 0, len(list))
	for _, i := range list {
		result = append(result, *i.InstanceId)
	}
	return result
}

// Expands an array of String Instance IDs into a []Instances
func expandInstanceString(list []interface{}) []*elb.Instance {
	result := make([]*elb.Instance, 0, len(list))
	for _, i := range list {
		result = append(result, &elb.Instance{InstanceId: aws.String(i.(string))})
	}
	return result
}

// Flattens an array of Backend Descriptions into a a map of instance_port to policy names.
func flattenBackendPolicies(backends []*elb.BackendServerDescription) map[int64][]string {
	policies := make(map[int64][]string)
	for _, i := range backends {
		for _, p := range i.PolicyNames {
			policies[*i.InstancePort] = append(policies[*i.InstancePort], *p)
		}
		sort.Strings(policies[*i.InstancePort])
	}
	return policies
}

// Flattens an array of Listeners into a []map[string]interface{}
func flattenListeners(list []*elb.ListenerDescription) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, i := range list {
		l := map[string]interface{}{
			"instance_port":     *i.Listener.InstancePort,
			"instance_protocol": strings.ToLower(*i.Listener.InstanceProtocol),
			"lb_port":           *i.Listener.LoadBalancerPort,
			"lb_protocol":       strings.ToLower(*i.Listener.Protocol),
		}
		// SSLCertificateID is optional, and may be nil
		if i.Listener.SSLCertificateId != nil {
			l["ssl_certificate_id"] = *i.Listener.SSLCertificateId
		}
		result = append(result, l)
	}
	return result
}

// Flattens an array of ECS LoadBalancers into a []map[string]interface{}
func flattenEcsLoadBalancers(list []*ecs.LoadBalancer) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, loadBalancer := range list {
		l := map[string]interface{}{
			"container_name": *loadBalancer.ContainerName,
			"container_port": *loadBalancer.ContainerPort,
		}

		if loadBalancer.LoadBalancerName != nil {
			l["elb_name"] = *loadBalancer.LoadBalancerName
		}

		if loadBalancer.TargetGroupArn != nil {
			l["target_group_arn"] = *loadBalancer.TargetGroupArn
		}

		result = append(result, l)
	}
	return result
}

// Flattens an array of Options into a []map[string]interface{}
func flattenOptions(apiOptions []*rds.Option, optionConfigurations []*rds.OptionConfiguration) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, apiOption := range apiOptions {
		if apiOption == nil || apiOption.OptionName == nil {
			continue
		}

		var configuredOption *rds.OptionConfiguration

		for _, optionConfiguration := range optionConfigurations {
			if aws.StringValue(apiOption.OptionName) == aws.StringValue(optionConfiguration.OptionName) {
				configuredOption = optionConfiguration
				break
			}
		}

		dbSecurityGroupMemberships := make([]interface{}, 0)
		for _, db := range apiOption.DBSecurityGroupMemberships {
			if db != nil {
				dbSecurityGroupMemberships = append(dbSecurityGroupMemberships, aws.StringValue(db.DBSecurityGroupName))
			}
		}

		optionSettings := make([]interface{}, 0)
		for _, apiOptionSetting := range apiOption.OptionSettings {
			// The RDS API responds with all settings. Omit settings that match default value,
			// but only if unconfigured. This is to prevent operators from continually needing
			// to continually update their Terraform configurations to match new option settings
			// when added by the API.
			var configuredOptionSetting *rds.OptionSetting

			if configuredOption != nil {
				for _, configuredOptionOptionSetting := range configuredOption.OptionSettings {
					if aws.StringValue(apiOptionSetting.Name) == aws.StringValue(configuredOptionOptionSetting.Name) {
						configuredOptionSetting = configuredOptionOptionSetting
						break
					}
				}
			}

			if configuredOptionSetting == nil && aws.StringValue(apiOptionSetting.Value) == aws.StringValue(apiOptionSetting.DefaultValue) {
				continue
			}

			optionSetting := map[string]interface{}{
				"name":  aws.StringValue(apiOptionSetting.Name),
				"value": aws.StringValue(apiOptionSetting.Value),
			}

			// Some values, like passwords, are sent back from the API as ****.
			// Set the response to match the configuration to prevent an unexpected difference
			if configuredOptionSetting != nil && aws.StringValue(apiOptionSetting.Value) == "****" {
				optionSetting["value"] = aws.StringValue(configuredOptionSetting.Value)
			}

			optionSettings = append(optionSettings, optionSetting)
		}
		optionSettingsResource := &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"value": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		}

		vpcSecurityGroupMemberships := make([]interface{}, 0)
		for _, vpc := range apiOption.VpcSecurityGroupMemberships {
			if vpc != nil {
				vpcSecurityGroupMemberships = append(vpcSecurityGroupMemberships, aws.StringValue(vpc.VpcSecurityGroupId))
			}
		}

		r := map[string]interface{}{
			"db_security_group_memberships":  schema.NewSet(schema.HashString, dbSecurityGroupMemberships),
			"option_name":                    aws.StringValue(apiOption.OptionName),
			"option_settings":                schema.NewSet(schema.HashResource(optionSettingsResource), optionSettings),
			"port":                           aws.Int64Value(apiOption.Port),
			"version":                        aws.StringValue(apiOption.OptionVersion),
			"vpc_security_group_memberships": schema.NewSet(schema.HashString, vpcSecurityGroupMemberships),
		}

		result = append(result, r)
	}

	return result
}

// Flattens an array of Parameters into a []map[string]interface{}
func flattenParameters(list []*rds.Parameter) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, i := range list {
		if i.ParameterName != nil {
			r := make(map[string]interface{})
			r["name"] = strings.ToLower(*i.ParameterName)
			// Default empty string, guard against nil parameter values
			r["value"] = ""
			if i.ParameterValue != nil {
				r["value"] = strings.ToLower(*i.ParameterValue)
			}
			if i.ApplyMethod != nil {
				r["apply_method"] = strings.ToLower(*i.ApplyMethod)
			}

			result = append(result, r)
		}
	}
	return result
}

// Flattens an array of Redshift Parameters into a []map[string]interface{}
func flattenRedshiftParameters(list []*redshift.Parameter) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, i := range list {
		result = append(result, map[string]interface{}{
			"name":  strings.ToLower(*i.ParameterName),
			"value": strings.ToLower(*i.ParameterValue),
		})
	}
	return result
}

// Flattens an array of Parameters into a []map[string]interface{}
func flattenNeptuneParameters(list []*neptune.Parameter) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, i := range list {
		if i.ParameterValue != nil {
			result = append(result, map[string]interface{}{
				"apply_method": aws.StringValue(i.ApplyMethod),
				"name":         aws.StringValue(i.ParameterName),
				"value":        aws.StringValue(i.ParameterValue),
			})
		}
	}
	return result
}

// Flattens an array of Parameters into a []map[string]interface{}
func flattenDocDBParameters(list []*docdb.Parameter, parameterList []interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, i := range list {
		if i.ParameterValue != nil {
			name := aws.StringValue(i.ParameterName)

			// Check if any non-user parameters are specified in the configuration.
			parameterFound := false
			for _, configParameter := range parameterList {
				if configParameter.(map[string]interface{})["name"] == name {
					parameterFound = true
				}
			}

			// Skip parameters that are not user defined or specified in the configuration.
			if aws.StringValue(i.Source) != "user" && !parameterFound {
				continue
			}

			result = append(result, map[string]interface{}{
				"apply_method": aws.StringValue(i.ApplyMethod),
				"name":         aws.StringValue(i.ParameterName),
				"value":        aws.StringValue(i.ParameterValue),
			})
		}
	}
	return result
}

// Takes the result of flatmap.Expand for an array of strings
// and returns a []*string
func expandStringList(configured []interface{}) []*string {
	vs := make([]*string, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(string)
		if ok && val != "" {
			vs = append(vs, aws.String(v.(string)))
		}
	}
	return vs
}

func expandStringListKeepEmpty(configured []interface{}) []*string {
	vs := make([]*string, 0, len(configured))
	for _, v := range configured {
		if val, ok := v.(string); ok {
			vs = append(vs, aws.String(val))
		}
	}
	return vs
}

// Takes the result of flatmap.Expand for an array of int64
// and returns a []*int64
func expandInt64List(configured []interface{}) []*int64 {
	vs := make([]*int64, 0, len(configured))
	for _, v := range configured {
		vs = append(vs, aws.Int64(int64(v.(int))))
	}
	return vs
}

// Expands a map of string to interface to a map of string to *float
func expandFloat64Map(m map[string]interface{}) map[string]*float64 {
	float64Map := make(map[string]*float64, len(m))
	for k, v := range m {
		float64Map[k] = aws.Float64(v.(float64))
	}
	return float64Map
}

// Expands a map of string to interface to a map of string to *string
func expandStringMap(m map[string]interface{}) map[string]*string {
	stringMap := make(map[string]*string, len(m))
	for k, v := range m {
		stringMap[k] = aws.String(v.(string))
	}
	return stringMap
}

// Takes the result of schema.Set of strings and returns a []*string
func expandStringSet(configured *schema.Set) []*string {
	return expandStringList(configured.List()) // nosemgrep: helper-schema-Set-extraneous-expandStringList-with-List
}

// Takes the result of schema.Set of strings and returns a []*int64
func expandInt64Set(configured *schema.Set) []*int64 {
	return expandInt64List(configured.List())
}

// Takes list of pointers to strings. Expand to an array
// of raw strings and returns a []interface{}
// to keep compatibility w/ schema.NewSetschema.NewSet
func flattenStringList(list []*string) []interface{} {
	vs := make([]interface{}, 0, len(list))
	for _, v := range list {
		vs = append(vs, *v)
	}
	return vs
}

func flattenStringSet(list []*string) *schema.Set {
	return schema.NewSet(schema.HashString, flattenStringList(list)) // nosemgrep: helper-schema-Set-extraneous-NewSet-with-flattenStringList
}

// hashStringCaseInsensitive hashes strings in a case insensitive manner.
// If you want a Set of strings and are case inensitive, this is the SchemaSetFunc you want.
func hashStringCaseInsensitive(v interface{}) int {
	return hashcode.String(strings.ToLower(v.(string)))
}

func flattenCaseInsensitiveStringSet(list []*string) *schema.Set {
	return schema.NewSet(hashStringCaseInsensitive, flattenStringList(list))
}

// Takes list of pointers to int64s. Expand to an array
// of raw ints and returns a []interface{}
// to keep compatibility w/ schema.NewSet
func flattenInt64List(list []*int64) []interface{} {
	vs := make([]interface{}, 0, len(list))
	for _, v := range list {
		vs = append(vs, int(aws.Int64Value(v)))
	}
	return vs
}

func flattenInt64Set(list []*int64) *schema.Set {
	return schema.NewSet(schema.HashInt, flattenInt64List(list))
}

//Flattens an array of private ip addresses into a []string, where the elements returned are the IP strings e.g. "192.168.0.0"
func flattenNetworkInterfacesPrivateIPAddresses(dtos []*ec2.NetworkInterfacePrivateIpAddress) []string {
	ips := make([]string, 0, len(dtos))
	for _, v := range dtos {
		ip := *v.PrivateIpAddress
		ips = append(ips, ip)
	}
	return ips
}

//Flattens security group identifiers into a []string, where the elements returned are the GroupIDs
func flattenGroupIdentifiers(dtos []*ec2.GroupIdentifier) []string {
	ids := make([]string, 0, len(dtos))
	for _, v := range dtos {
		group_id := *v.GroupId
		ids = append(ids, group_id)
	}
	return ids
}

//Expands an array of IPs into a ec2 Private IP Address Spec
func expandPrivateIPAddresses(ips []interface{}) []*ec2.PrivateIpAddressSpecification {
	dtos := make([]*ec2.PrivateIpAddressSpecification, 0, len(ips))
	for i, v := range ips {
		new_private_ip := &ec2.PrivateIpAddressSpecification{
			PrivateIpAddress: aws.String(v.(string)),
		}

		new_private_ip.Primary = aws.Bool(i == 0)

		dtos = append(dtos, new_private_ip)
	}
	return dtos
}

func expandIP6Addresses(ips []interface{}) []*ec2.InstanceIpv6Address {
	dtos := make([]*ec2.InstanceIpv6Address, 0, len(ips))
	for _, v := range ips {
		ipv6Address := &ec2.InstanceIpv6Address{
			Ipv6Address: aws.String(v.(string)),
		}

		dtos = append(dtos, ipv6Address)
	}
	return dtos
}

//Flattens network interface attachment into a map[string]interface
func flattenAttachment(a *ec2.NetworkInterfaceAttachment) map[string]interface{} {
	att := make(map[string]interface{})
	if a.InstanceId != nil {
		att["instance"] = *a.InstanceId
	}
	if a.DeviceIndex != nil {
		att["device_index"] = *a.DeviceIndex
	}
	if a.AttachmentId != nil {
		att["attachment_id"] = *a.AttachmentId
	}
	return att
}

func flattenEc2AttributeValues(l []*ec2.AttributeValue) []string {
	values := make([]string, 0, len(l))
	for _, v := range l {
		values = append(values, aws.StringValue(v.Value))
	}
	return values
}

func flattenEc2NetworkInterfaceAssociation(a *ec2.NetworkInterfaceAssociation) []interface{} {
	tfMap := map[string]interface{}{}

	if a.AllocationId != nil {
		tfMap["allocation_id"] = aws.StringValue(a.AllocationId)
	}
	if a.AssociationId != nil {
		tfMap["association_id"] = aws.StringValue(a.AssociationId)
	}
	if a.CarrierIp != nil {
		tfMap["carrier_ip"] = aws.StringValue(a.CarrierIp)
	}
	if a.CustomerOwnedIp != nil {
		tfMap["customer_owned_ip"] = aws.StringValue(a.CustomerOwnedIp)
	}
	if a.IpOwnerId != nil {
		tfMap["ip_owner_id"] = aws.StringValue(a.IpOwnerId)
	}
	if a.PublicDnsName != nil {
		tfMap["public_dns_name"] = aws.StringValue(a.PublicDnsName)
	}
	if a.PublicIp != nil {
		tfMap["public_ip"] = aws.StringValue(a.PublicIp)
	}

	return []interface{}{tfMap}
}

func flattenEc2NetworkInterfaceIpv6Address(niia []*ec2.NetworkInterfaceIpv6Address) []string {
	ips := make([]string, 0, len(niia))
	for _, v := range niia {
		ips = append(ips, *v.Ipv6Address)
	}
	return ips
}

func flattenElastiCacheSecurityGroupNames(securityGroups []*elasticache.CacheSecurityGroupMembership) []string {
	result := make([]string, 0, len(securityGroups))
	for _, sg := range securityGroups {
		if sg.CacheSecurityGroupName != nil {
			result = append(result, *sg.CacheSecurityGroupName)
		}
	}
	return result
}

func flattenElastiCacheSecurityGroupIds(securityGroups []*elasticache.SecurityGroupMembership) []string {
	result := make([]string, 0, len(securityGroups))
	for _, sg := range securityGroups {
		if sg.SecurityGroupId != nil {
			result = append(result, *sg.SecurityGroupId)
		}
	}
	return result
}

func flattenDaxSecurityGroupIds(securityGroups []*dax.SecurityGroupMembership) []string {
	result := make([]string, 0, len(securityGroups))
	for _, sg := range securityGroups {
		if sg.SecurityGroupIdentifier != nil {
			result = append(result, *sg.SecurityGroupIdentifier)
		}
	}
	return result
}

// Flattens step adjustments into a list of map[string]interface.
func flattenStepAdjustments(adjustments []*autoscaling.StepAdjustment) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(adjustments))
	for _, raw := range adjustments {
		a := map[string]interface{}{
			"scaling_adjustment": aws.Int64Value(raw.ScalingAdjustment),
		}
		if raw.MetricIntervalUpperBound != nil {
			a["metric_interval_upper_bound"] = fmt.Sprintf("%g", aws.Float64Value(raw.MetricIntervalUpperBound))
		}
		if raw.MetricIntervalLowerBound != nil {
			a["metric_interval_lower_bound"] = fmt.Sprintf("%g", aws.Float64Value(raw.MetricIntervalLowerBound))
		}
		result = append(result, a)
	}
	return result
}

func flattenResourceRecords(recs []*route53.ResourceRecord, typeStr string) []string {
	strs := make([]string, 0, len(recs))
	for _, r := range recs {
		if r.Value != nil {
			s := *r.Value
			if typeStr == "TXT" || typeStr == "SPF" {
				s = expandTxtEntry(s)
			}
			strs = append(strs, s)
		}
	}
	return strs
}

func expandResourceRecords(recs []interface{}, typeStr string) []*route53.ResourceRecord {
	records := make([]*route53.ResourceRecord, 0, len(recs))
	for _, r := range recs {
		s := r.(string)
		if typeStr == "TXT" || typeStr == "SPF" {
			s = flattenTxtEntry(s)
		}
		records = append(records, &route53.ResourceRecord{Value: aws.String(s)})
	}
	return records
}

// How 'flattenTxtEntry' and 'expandTxtEntry' work.
//
// In the Route 53, TXT entries are written using quoted strings, one per line.
// Example:
//     "x=foo"
//     "bar=12"
//
// In Terraform, there are two differences:
// - We use a list of strings instead of separating strings with newlines.
// - Within each string, we dont' include the surrounding quotes.
// Example:
//     records = ["x=foo", "bar=12"]    # Instead of ["\"x=foo\", \"bar=12\""]
//
// When we pull from Route 53, `expandTxtEntry` removes the surrounding quotes;
// when we push to Route 53, `flattenTxtEntry` adds them back.
//
// One complication is that a single TXT entry can have multiple quoted strings.
// For example, here are two TXT entries, one with two quoted strings and the
// other with three.
//     "x=" "foo"
//     "ba" "r" "=12"
//
// DNS clients are expected to merge the quoted strings before interpreting the
// value.  Since `expandTxtEntry` only removes the quotes at the end we can still
// (hackily) represent the above configuration in Terraform:
//      records = ["x=\" \"foo", "ba\" \"r\" \"=12"]
//
// The primary reason to use multiple strings for an entry is that DNS (and Route
// 53) doesn't allow a quoted string to be more than 255 characters long.  If you
// want a longer TXT entry, you must use multiple quoted strings.
//
// It would be nice if this Terraform automatically split strings longer than 255
// characters.  For example, imagine "xxx..xxx" has 256 "x" characters.
//      records = ["xxx..xxx"]
// When pushing to Route 53, this could be converted to:
//      "xxx..xx" "x"
//
// This could also work when the user is already using multiple quoted strings:
//      records = ["xxx.xxx\" \"yyy..yyy"]
// When pushing to Route 53, this could be converted to:
//       "xxx..xx" "xyyy...y" "yy"
//
// If you want to add this feature, make sure to follow all the quoting rules in
// <https://tools.ietf.org/html/rfc1464#section-2>.  If you make a mistake, people
// might end up relying on that mistake so fixing it would be a breaking change.

func flattenTxtEntry(s string) string {
	return fmt.Sprintf(`"%s"`, s)
}

func expandTxtEntry(s string) string {
	last := len(s) - 1
	if last != 0 && s[0] == '"' && s[last] == '"' {
		s = s[1:last]
	}
	return s
}
func expandESCognitoOptions(c []interface{}) *elasticsearch.CognitoOptions {
	options := &elasticsearch.CognitoOptions{
		Enabled: aws.Bool(false),
	}
	if len(c) < 1 {
		return options
	}

	m := c[0].(map[string]interface{})

	if cognitoEnabled, ok := m["enabled"]; ok {
		options.Enabled = aws.Bool(cognitoEnabled.(bool))

		if cognitoEnabled.(bool) {

			if v, ok := m["user_pool_id"]; ok && v.(string) != "" {
				options.UserPoolId = aws.String(v.(string))
			}
			if v, ok := m["identity_pool_id"]; ok && v.(string) != "" {
				options.IdentityPoolId = aws.String(v.(string))
			}
			if v, ok := m["role_arn"]; ok && v.(string) != "" {
				options.RoleArn = aws.String(v.(string))
			}
		}
	}

	return options
}

func flattenESCognitoOptions(c *elasticsearch.CognitoOptions) []map[string]interface{} {
	m := map[string]interface{}{}

	m["enabled"] = aws.BoolValue(c.Enabled)

	if aws.BoolValue(c.Enabled) {
		m["identity_pool_id"] = aws.StringValue(c.IdentityPoolId)
		m["user_pool_id"] = aws.StringValue(c.UserPoolId)
		m["role_arn"] = aws.StringValue(c.RoleArn)
	}

	return []map[string]interface{}{m}
}

func expandESDomainEndpointOptions(l []interface{}) *elasticsearch.DomainEndpointOptions {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	domainEndpointOptions := &elasticsearch.DomainEndpointOptions{}

	if v, ok := m["enforce_https"].(bool); ok {
		domainEndpointOptions.EnforceHTTPS = aws.Bool(v)
	}

	if v, ok := m["tls_security_policy"].(string); ok {
		domainEndpointOptions.TLSSecurityPolicy = aws.String(v)
	}

	if customEndpointEnabled, ok := m["custom_endpoint_enabled"]; ok {
		domainEndpointOptions.CustomEndpointEnabled = aws.Bool(customEndpointEnabled.(bool))

		if customEndpointEnabled.(bool) {
			if v, ok := m["custom_endpoint"].(string); ok && v != "" {
				domainEndpointOptions.CustomEndpoint = aws.String(v)
			}

			if v, ok := m["custom_endpoint_certificate_arn"].(string); ok && v != "" {
				domainEndpointOptions.CustomEndpointCertificateArn = aws.String(v)
			}
		}
	}

	return domainEndpointOptions
}

func flattenESDomainEndpointOptions(domainEndpointOptions *elasticsearch.DomainEndpointOptions) []interface{} {
	if domainEndpointOptions == nil {
		return nil
	}

	m := map[string]interface{}{
		"enforce_https":           aws.BoolValue(domainEndpointOptions.EnforceHTTPS),
		"tls_security_policy":     aws.StringValue(domainEndpointOptions.TLSSecurityPolicy),
		"custom_endpoint_enabled": aws.BoolValue(domainEndpointOptions.CustomEndpointEnabled),
	}
	if aws.BoolValue(domainEndpointOptions.CustomEndpointEnabled) {
		if domainEndpointOptions.CustomEndpoint != nil {
			m["custom_endpoint"] = aws.StringValue(domainEndpointOptions.CustomEndpoint)
		}
		if domainEndpointOptions.CustomEndpointCertificateArn != nil {
			m["custom_endpoint_certificate_arn"] = aws.StringValue(domainEndpointOptions.CustomEndpointCertificateArn)
		}
	}

	return []interface{}{m}
}

func flattenESSnapshotOptions(snapshotOptions *elasticsearch.SnapshotOptions) []map[string]interface{} {
	if snapshotOptions == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"automated_snapshot_start_hour": int(aws.Int64Value(snapshotOptions.AutomatedSnapshotStartHour)),
	}

	return []map[string]interface{}{m}
}

func flattenESEBSOptions(o *elasticsearch.EBSOptions) []map[string]interface{} {
	m := map[string]interface{}{}

	if o.EBSEnabled != nil {
		m["ebs_enabled"] = *o.EBSEnabled
	}

	if aws.BoolValue(o.EBSEnabled) {
		if o.Iops != nil {
			m["iops"] = *o.Iops
		}
		if o.VolumeSize != nil {
			m["volume_size"] = *o.VolumeSize
		}
		if o.VolumeType != nil {
			m["volume_type"] = *o.VolumeType
		}
	}

	return []map[string]interface{}{m}
}

func expandESEBSOptions(m map[string]interface{}) *elasticsearch.EBSOptions {
	options := elasticsearch.EBSOptions{}

	if ebsEnabled, ok := m["ebs_enabled"]; ok {
		options.EBSEnabled = aws.Bool(ebsEnabled.(bool))

		if ebsEnabled.(bool) {
			if v, ok := m["iops"]; ok && v.(int) > 0 {
				options.Iops = aws.Int64(int64(v.(int)))
			}
			if v, ok := m["volume_size"]; ok && v.(int) > 0 {
				options.VolumeSize = aws.Int64(int64(v.(int)))
			}
			if v, ok := m["volume_type"]; ok && v.(string) != "" {
				options.VolumeType = aws.String(v.(string))
			}
		}
	}

	return &options
}

func flattenESEncryptAtRestOptions(o *elasticsearch.EncryptionAtRestOptions) []map[string]interface{} {
	if o == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if o.Enabled != nil {
		m["enabled"] = *o.Enabled
	}
	if o.KmsKeyId != nil {
		m["kms_key_id"] = *o.KmsKeyId
	}

	return []map[string]interface{}{m}
}

func expandESEncryptAtRestOptions(m map[string]interface{}) *elasticsearch.EncryptionAtRestOptions {
	options := elasticsearch.EncryptionAtRestOptions{}

	if v, ok := m["enabled"]; ok {
		options.Enabled = aws.Bool(v.(bool))
	}
	if v, ok := m["kms_key_id"]; ok && v.(string) != "" {
		options.KmsKeyId = aws.String(v.(string))
	}

	return &options
}

func flattenESVPCDerivedInfo(o *elasticsearch.VPCDerivedInfo) []map[string]interface{} {
	m := map[string]interface{}{}

	if o.AvailabilityZones != nil {
		m["availability_zones"] = flattenStringSet(o.AvailabilityZones)
	}
	if o.SecurityGroupIds != nil {
		m["security_group_ids"] = flattenStringSet(o.SecurityGroupIds)
	}
	if o.SubnetIds != nil {
		m["subnet_ids"] = flattenStringSet(o.SubnetIds)
	}
	if o.VPCId != nil {
		m["vpc_id"] = *o.VPCId
	}

	return []map[string]interface{}{m}
}

func expandESVPCOptions(m map[string]interface{}) *elasticsearch.VPCOptions {
	options := elasticsearch.VPCOptions{}

	if v, ok := m["security_group_ids"]; ok {
		options.SecurityGroupIds = expandStringSet(v.(*schema.Set))
	}
	if v, ok := m["subnet_ids"]; ok {
		options.SubnetIds = expandStringSet(v.(*schema.Set))
	}

	return &options
}

func expandConfigRecordingGroup(configured []interface{}) *configservice.RecordingGroup {
	recordingGroup := configservice.RecordingGroup{}
	group := configured[0].(map[string]interface{})

	if v, ok := group["all_supported"]; ok {
		recordingGroup.AllSupported = aws.Bool(v.(bool))
	}

	if v, ok := group["include_global_resource_types"]; ok {
		recordingGroup.IncludeGlobalResourceTypes = aws.Bool(v.(bool))
	}

	if v, ok := group["resource_types"]; ok {
		recordingGroup.ResourceTypes = expandStringSet(v.(*schema.Set))
	}
	return &recordingGroup
}

func flattenConfigRecordingGroup(g *configservice.RecordingGroup) []map[string]interface{} {
	m := make(map[string]interface{}, 1)

	if g.AllSupported != nil {
		m["all_supported"] = *g.AllSupported
	}

	if g.IncludeGlobalResourceTypes != nil {
		m["include_global_resource_types"] = *g.IncludeGlobalResourceTypes
	}

	if g.ResourceTypes != nil && len(g.ResourceTypes) > 0 {
		m["resource_types"] = flattenStringSet(g.ResourceTypes)
	}

	return []map[string]interface{}{m}
}

func flattenConfigSnapshotDeliveryProperties(p *configservice.ConfigSnapshotDeliveryProperties) []map[string]interface{} {
	m := make(map[string]interface{})

	if p.DeliveryFrequency != nil {
		m["delivery_frequency"] = *p.DeliveryFrequency
	}

	return []map[string]interface{}{m}
}

func pointersMapToStringList(pointers map[string]*string) map[string]interface{} {
	list := make(map[string]interface{}, len(pointers))
	for i, v := range pointers {
		list[i] = *v
	}
	return list
}

// diffStringMaps returns the set of keys and values that must be created, the set of keys
// and values that must be destroyed, and the set of keys and values that are unchanged.
func diffStringMaps(oldMap, newMap map[string]interface{}) (map[string]*string, map[string]*string, map[string]*string) {
	// First, we're creating everything we have
	create := map[string]*string{}
	for k, v := range newMap {
		create[k] = aws.String(v.(string))
	}

	// Build the maps of what to remove and what is unchanged
	remove := map[string]*string{}
	unchanged := map[string]*string{}
	for k, v := range oldMap {
		old, ok := create[k]
		if !ok || aws.StringValue(old) != v.(string) {
			// Delete it!
			remove[k] = aws.String(v.(string))
		} else if ok {
			unchanged[k] = aws.String(v.(string))
			// already present so remove from new
			delete(create, k)
		}
	}

	return create, remove, unchanged
}

func flattenDSVpcSettings(
	s *directoryservice.DirectoryVpcSettingsDescription) []map[string]interface{} {
	settings := make(map[string]interface{})

	if s == nil {
		return nil
	}

	settings["subnet_ids"] = flattenStringSet(s.SubnetIds)
	settings["vpc_id"] = aws.StringValue(s.VpcId)
	settings["availability_zones"] = flattenStringSet(s.AvailabilityZones)

	return []map[string]interface{}{settings}
}

func flattenLambdaLayers(layers []*lambda.Layer) []interface{} {
	arns := make([]*string, len(layers))
	for i, layer := range layers {
		arns[i] = layer.Arn
	}
	return flattenStringList(arns)
}

func flattenLambdaVpcConfigResponse(s *lambda.VpcConfigResponse) []map[string]interface{} {
	settings := make(map[string]interface{})

	if s == nil {
		return nil
	}

	var emptyVpc bool
	if s.VpcId == nil || *s.VpcId == "" {
		emptyVpc = true
	}
	if len(s.SubnetIds) == 0 && len(s.SecurityGroupIds) == 0 && emptyVpc {
		return nil
	}

	settings["subnet_ids"] = flattenStringSet(s.SubnetIds)
	settings["security_group_ids"] = flattenStringSet(s.SecurityGroupIds)
	if s.VpcId != nil {
		settings["vpc_id"] = *s.VpcId
	}

	return []map[string]interface{}{settings}
}

func flattenLambdaAliasRoutingConfiguration(arc *lambda.AliasRoutingConfiguration) []interface{} {
	if arc == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"additional_version_weights": aws.Float64ValueMap(arc.AdditionalVersionWeights),
	}

	return []interface{}{m}
}

func flattenDSConnectSettings(
	customerDnsIps []*string,
	s *directoryservice.DirectoryConnectSettingsDescription) []map[string]interface{} {
	if s == nil {
		return nil
	}

	settings := make(map[string]interface{})

	settings["customer_dns_ips"] = flattenStringSet(customerDnsIps)
	settings["connect_ips"] = flattenStringSet(s.ConnectIps)
	settings["customer_username"] = aws.StringValue(s.CustomerUserName)
	settings["subnet_ids"] = flattenStringSet(s.SubnetIds)
	settings["vpc_id"] = aws.StringValue(s.VpcId)
	settings["availability_zones"] = flattenStringSet(s.AvailabilityZones)

	return []map[string]interface{}{settings}
}

func expandCloudFormationParameters(params map[string]interface{}) []*cloudformation.Parameter {
	var cfParams []*cloudformation.Parameter
	for k, v := range params {
		cfParams = append(cfParams, &cloudformation.Parameter{
			ParameterKey:   aws.String(k),
			ParameterValue: aws.String(v.(string)),
		})
	}

	return cfParams
}

// flattenCloudFormationParameters is flattening list of
// *cloudformation.Parameters and only returning existing
// parameters to avoid clash with default values
func flattenCloudFormationParameters(cfParams []*cloudformation.Parameter,
	originalParams map[string]interface{}) map[string]interface{} {
	params := make(map[string]interface{}, len(cfParams))
	for _, p := range cfParams {
		_, isConfigured := originalParams[*p.ParameterKey]
		if isConfigured {
			params[*p.ParameterKey] = *p.ParameterValue
		}
	}
	return params
}

func flattenAllCloudFormationParameters(cfParams []*cloudformation.Parameter) map[string]interface{} {
	params := make(map[string]interface{}, len(cfParams))
	for _, p := range cfParams {
		params[*p.ParameterKey] = *p.ParameterValue
	}
	return params
}

func flattenCloudFormationOutputs(cfOutputs []*cloudformation.Output) map[string]string {
	outputs := make(map[string]string, len(cfOutputs))
	for _, o := range cfOutputs {
		outputs[*o.OutputKey] = *o.OutputValue
	}
	return outputs
}

func flattenAsgSuspendedProcesses(list []*autoscaling.SuspendedProcess) []string {
	strs := make([]string, 0, len(list))
	for _, r := range list {
		if r.ProcessName != nil {
			strs = append(strs, *r.ProcessName)
		}
	}
	return strs
}

func flattenAsgEnabledMetrics(list []*autoscaling.EnabledMetric) []string {
	strs := make([]string, 0, len(list))
	for _, r := range list {
		if r.Metric != nil {
			strs = append(strs, *r.Metric)
		}
	}
	return strs
}

func flattenKinesisShardLevelMetrics(list []*kinesis.EnhancedMetrics) []string {
	if len(list) == 0 {
		return []string{}
	}
	strs := make([]string, 0, len(list[0].ShardLevelMetrics))
	for _, s := range list[0].ShardLevelMetrics {
		strs = append(strs, *s)
	}
	return strs
}

func expandApiGatewayRequestResponseModelOperations(d *schema.ResourceData, key string, prefix string) []*apigateway.PatchOperation {
	operations := make([]*apigateway.PatchOperation, 0)

	oldModels, newModels := d.GetChange(key)
	oldModelMap := oldModels.(map[string]interface{})
	newModelMap := newModels.(map[string]interface{})

	for k := range oldModelMap {
		operation := apigateway.PatchOperation{
			Op:   aws.String("remove"),
			Path: aws.String(fmt.Sprintf("/%s/%s", prefix, strings.Replace(k, "/", "~1", -1))),
		}

		for nK, nV := range newModelMap {
			if nK == k {
				operation.Op = aws.String("replace")
				operation.Value = aws.String(nV.(string))
			}
		}

		operations = append(operations, &operation)
	}

	for nK, nV := range newModelMap {
		exists := false
		for k := range oldModelMap {
			if k == nK {
				exists = true
			}
		}
		if !exists {
			operation := apigateway.PatchOperation{
				Op:    aws.String("add"),
				Path:  aws.String(fmt.Sprintf("/%s/%s", prefix, strings.Replace(nK, "/", "~1", -1))),
				Value: aws.String(nV.(string)),
			}
			operations = append(operations, &operation)
		}
	}

	return operations
}

func expandApiGatewayMethodParametersOperations(d *schema.ResourceData, key string, prefix string) []*apigateway.PatchOperation {
	operations := make([]*apigateway.PatchOperation, 0)

	oldParameters, newParameters := d.GetChange(key)
	oldParametersMap := oldParameters.(map[string]interface{})
	newParametersMap := newParameters.(map[string]interface{})

	for k := range oldParametersMap {
		operation := apigateway.PatchOperation{
			Op:   aws.String("remove"),
			Path: aws.String(fmt.Sprintf("/%s/%s", prefix, k)),
		}

		for nK, nV := range newParametersMap {
			b, ok := nV.(bool)
			if !ok {
				value, _ := strconv.ParseBool(nV.(string))
				b = value
			}
			if nK == k {
				operation.Op = aws.String("replace")
				operation.Value = aws.String(strconv.FormatBool(b))
			}
		}

		operations = append(operations, &operation)
	}

	for nK, nV := range newParametersMap {
		exists := false
		for k := range oldParametersMap {
			if k == nK {
				exists = true
			}
		}
		if !exists {
			b, ok := nV.(bool)
			if !ok {
				value, _ := strconv.ParseBool(nV.(string))
				b = value
			}
			operation := apigateway.PatchOperation{
				Op:    aws.String("add"),
				Path:  aws.String(fmt.Sprintf("/%s/%s", prefix, nK)),
				Value: aws.String(strconv.FormatBool(b)),
			}
			operations = append(operations, &operation)
		}
	}

	return operations
}

func flattenBeanstalkAsg(list []*elasticbeanstalk.AutoScalingGroup) []string {
	strs := make([]string, 0, len(list))
	for _, r := range list {
		if r.Name != nil {
			strs = append(strs, *r.Name)
		}
	}
	return strs
}

func flattenBeanstalkInstances(list []*elasticbeanstalk.Instance) []string {
	strs := make([]string, 0, len(list))
	for _, r := range list {
		if r.Id != nil {
			strs = append(strs, *r.Id)
		}
	}
	return strs
}

func flattenBeanstalkLc(list []*elasticbeanstalk.LaunchConfiguration) []string {
	strs := make([]string, 0, len(list))
	for _, r := range list {
		if r.Name != nil {
			strs = append(strs, *r.Name)
		}
	}
	return strs
}

func flattenBeanstalkElb(list []*elasticbeanstalk.LoadBalancer) []string {
	strs := make([]string, 0, len(list))
	for _, r := range list {
		if r.Name != nil {
			strs = append(strs, *r.Name)
		}
	}
	return strs
}

func flattenBeanstalkSqs(list []*elasticbeanstalk.Queue) []string {
	strs := make([]string, 0, len(list))
	for _, r := range list {
		if r.URL != nil {
			strs = append(strs, *r.URL)
		}
	}
	return strs
}

func flattenBeanstalkTrigger(list []*elasticbeanstalk.Trigger) []string {
	strs := make([]string, 0, len(list))
	for _, r := range list {
		if r.Name != nil {
			strs = append(strs, *r.Name)
		}
	}
	return strs
}

// There are several parts of the AWS API that will sort lists of strings,
// causing diffs between resources that use lists. This avoids a bit of
// code duplication for pre-sorts that can be used for things like hash
// functions, etc.
func sortInterfaceSlice(in []interface{}) []interface{} {
	a := []string{}
	b := []interface{}{}
	for _, v := range in {
		a = append(a, v.(string))
	}

	sort.Strings(a)

	for _, v := range a {
		b = append(b, v)
	}

	return b
}

// This function sorts List A to look like a list found in the tf file.
func sortListBasedonTFFile(in []string, d *schema.ResourceData) ([]string, error) {
	listName := "layer_ids"
	if attributeCount, ok := d.Get(listName + ".#").(int); ok {
		for i := 0; i < attributeCount; i++ {
			currAttributeId := d.Get(listName + "." + strconv.Itoa(i))
			for j := 0; j < len(in); j++ {
				if currAttributeId == in[j] {
					in[i], in[j] = in[j], in[i]
				}
			}
		}
		return in, nil
	}
	return in, fmt.Errorf("Could not find list: %s", listName)
}

func flattenApiGatewayThrottleSettings(settings *apigateway.ThrottleSettings) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)

	if settings != nil {
		r := make(map[string]interface{})
		if settings.BurstLimit != nil {
			r["burst_limit"] = *settings.BurstLimit
		}

		if settings.RateLimit != nil {
			r["rate_limit"] = *settings.RateLimit
		}

		result = append(result, r)
	}

	return result
}

// TODO: refactor some of these helper functions and types in the terraform/helper packages

// Takes the result of flatmap.Expand for an array of policy attributes and
// returns ELB API compatible objects
func expandPolicyAttributes(configured []interface{}) []*elb.PolicyAttribute {
	attributes := make([]*elb.PolicyAttribute, 0, len(configured))

	// Loop over our configured attributes and create
	// an array of aws-sdk-go compatible objects
	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})

		a := &elb.PolicyAttribute{
			AttributeName:  aws.String(data["name"].(string)),
			AttributeValue: aws.String(data["value"].(string)),
		}

		attributes = append(attributes, a)

	}

	return attributes
}

// Flattens an array of PolicyAttributes into a []interface{}
func flattenPolicyAttributes(list []*elb.PolicyAttributeDescription) []interface{} {
	attributes := []interface{}{}
	for _, attrdef := range list {
		attribute := map[string]string{
			"name":  *attrdef.AttributeName,
			"value": *attrdef.AttributeValue,
		}

		attributes = append(attributes, attribute)

	}

	return attributes
}

func expandConfigAccountAggregationSources(configured []interface{}) []*configservice.AccountAggregationSource {
	var results []*configservice.AccountAggregationSource
	for _, item := range configured {
		detail := item.(map[string]interface{})
		source := configservice.AccountAggregationSource{
			AllAwsRegions: aws.Bool(detail["all_regions"].(bool)),
		}

		if v, ok := detail["account_ids"]; ok {
			accountIDs := v.([]interface{})
			if len(accountIDs) > 0 {
				source.AccountIds = expandStringList(accountIDs)
			}
		}

		if v, ok := detail["regions"]; ok {
			regions := v.([]interface{})
			if len(regions) > 0 {
				source.AwsRegions = expandStringList(regions)
			}
		}

		results = append(results, &source)
	}
	return results
}

func expandConfigOrganizationAggregationSource(configured map[string]interface{}) *configservice.OrganizationAggregationSource {
	source := configservice.OrganizationAggregationSource{
		AllAwsRegions: aws.Bool(configured["all_regions"].(bool)),
		RoleArn:       aws.String(configured["role_arn"].(string)),
	}

	if v, ok := configured["regions"]; ok {
		regions := v.([]interface{})
		if len(regions) > 0 {
			source.AwsRegions = expandStringList(regions)
		}
	}

	return &source
}

func flattenConfigAccountAggregationSources(sources []*configservice.AccountAggregationSource) []interface{} {
	var result []interface{}

	if len(sources) == 0 {
		return result
	}

	source := sources[0]
	m := make(map[string]interface{})
	m["account_ids"] = flattenStringList(source.AccountIds)
	m["all_regions"] = aws.BoolValue(source.AllAwsRegions)
	m["regions"] = flattenStringList(source.AwsRegions)
	result = append(result, m)
	return result
}

func flattenConfigOrganizationAggregationSource(source *configservice.OrganizationAggregationSource) []interface{} {
	var result []interface{}

	if source == nil {
		return result
	}

	m := make(map[string]interface{})
	m["all_regions"] = aws.BoolValue(source.AllAwsRegions)
	m["regions"] = flattenStringList(source.AwsRegions)
	m["role_arn"] = aws.StringValue(source.RoleArn)
	result = append(result, m)
	return result
}

func flattenConfigRuleSource(source *configservice.Source) []interface{} {
	var result []interface{}
	m := make(map[string]interface{})
	m["owner"] = *source.Owner
	m["source_identifier"] = *source.SourceIdentifier
	if len(source.SourceDetails) > 0 {
		m["source_detail"] = schema.NewSet(configRuleSourceDetailsHash, flattenConfigRuleSourceDetails(source.SourceDetails))
	}
	result = append(result, m)
	return result
}

func flattenConfigRuleSourceDetails(details []*configservice.SourceDetail) []interface{} {
	var items []interface{}
	for _, d := range details {
		m := make(map[string]interface{})
		if d.MessageType != nil {
			m["message_type"] = *d.MessageType
		}
		if d.EventSource != nil {
			m["event_source"] = *d.EventSource
		}
		if d.MaximumExecutionFrequency != nil {
			m["maximum_execution_frequency"] = *d.MaximumExecutionFrequency
		}

		items = append(items, m)
	}

	return items
}

func expandConfigRuleSource(configured []interface{}) *configservice.Source {
	cfg := configured[0].(map[string]interface{})
	source := configservice.Source{
		Owner:            aws.String(cfg["owner"].(string)),
		SourceIdentifier: aws.String(cfg["source_identifier"].(string)),
	}
	if details, ok := cfg["source_detail"]; ok {
		source.SourceDetails = expandConfigRuleSourceDetails(details.(*schema.Set))
	}
	return &source
}

func expandConfigRuleSourceDetails(configured *schema.Set) []*configservice.SourceDetail {
	var results []*configservice.SourceDetail

	for _, item := range configured.List() {
		detail := item.(map[string]interface{})
		src := configservice.SourceDetail{}

		if msgType, ok := detail["message_type"].(string); ok && msgType != "" {
			src.MessageType = aws.String(msgType)
		}
		if eventSource, ok := detail["event_source"].(string); ok && eventSource != "" {
			src.EventSource = aws.String(eventSource)
		}
		if maxExecFreq, ok := detail["maximum_execution_frequency"].(string); ok && maxExecFreq != "" {
			src.MaximumExecutionFrequency = aws.String(maxExecFreq)
		}

		results = append(results, &src)
	}

	return results
}

func flattenConfigRuleScope(scope *configservice.Scope) []interface{} {
	var items []interface{}

	m := make(map[string]interface{})
	if scope.ComplianceResourceId != nil {
		m["compliance_resource_id"] = *scope.ComplianceResourceId
	}
	if scope.ComplianceResourceTypes != nil {
		m["compliance_resource_types"] = flattenStringSet(scope.ComplianceResourceTypes)
	}
	if scope.TagKey != nil {
		m["tag_key"] = *scope.TagKey
	}
	if scope.TagValue != nil {
		m["tag_value"] = *scope.TagValue
	}

	items = append(items, m)
	return items
}

func expandConfigRuleScope(l []interface{}) *configservice.Scope {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	configured := l[0].(map[string]interface{})
	scope := &configservice.Scope{}

	if v, ok := configured["compliance_resource_id"].(string); ok && v != "" {
		scope.ComplianceResourceId = aws.String(v)
	}
	if v, ok := configured["compliance_resource_types"]; ok {
		l := v.(*schema.Set)
		if l.Len() > 0 {
			scope.ComplianceResourceTypes = expandStringSet(l)
		}
	}
	if v, ok := configured["tag_key"].(string); ok && v != "" {
		scope.TagKey = aws.String(v)
	}
	if v, ok := configured["tag_value"].(string); ok && v != "" {
		scope.TagValue = aws.String(v)
	}

	return scope
}

// Takes a value containing YAML string and passes it through
// the YAML parser. Returns either a parsing
// error or original YAML string.
func checkYamlString(yamlString interface{}) (string, error) {
	var y interface{}

	if yamlString == nil || yamlString.(string) == "" {
		return "", nil
	}

	s := yamlString.(string)

	err := yaml.Unmarshal([]byte(s), &y)

	return s, err
}

func normalizeJsonOrYamlString(templateString interface{}) (string, error) {
	if looksLikeJsonString(templateString) {
		return structure.NormalizeJsonString(templateString.(string))
	}

	return checkYamlString(templateString)
}

func buildApiGatewayInvokeURL(client *AWSClient, restApiId, stageName string) string {
	hostname := client.RegionalHostname(fmt.Sprintf("%s.execute-api", restApiId))
	return fmt.Sprintf("https://%s/%s", hostname, stageName)
}

func expandCognitoSupportedLoginProviders(config map[string]interface{}) map[string]*string {
	m := map[string]*string{}
	for k, v := range config {
		s := v.(string)
		m[k] = &s
	}
	return m
}

func flattenCognitoSupportedLoginProviders(config map[string]*string) map[string]string {
	m := map[string]string{}
	for k, v := range config {
		m[k] = *v
	}
	return m
}

func expandCognitoIdentityProviders(s *schema.Set) []*cognitoidentity.Provider {
	ips := make([]*cognitoidentity.Provider, 0)

	for _, v := range s.List() {
		s := v.(map[string]interface{})

		ip := &cognitoidentity.Provider{}

		if sv, ok := s["client_id"].(string); ok {
			ip.ClientId = aws.String(sv)
		}

		if sv, ok := s["provider_name"].(string); ok {
			ip.ProviderName = aws.String(sv)
		}

		if sv, ok := s["server_side_token_check"].(bool); ok {
			ip.ServerSideTokenCheck = aws.Bool(sv)
		}

		ips = append(ips, ip)
	}

	return ips
}

func flattenCognitoIdentityProviders(ips []*cognitoidentity.Provider) []map[string]interface{} {
	values := make([]map[string]interface{}, 0)

	for _, v := range ips {
		ip := make(map[string]interface{})

		if v == nil {
			return nil
		}

		if v.ClientId != nil {
			ip["client_id"] = *v.ClientId
		}

		if v.ProviderName != nil {
			ip["provider_name"] = *v.ProviderName
		}

		if v.ServerSideTokenCheck != nil {
			ip["server_side_token_check"] = *v.ServerSideTokenCheck
		}

		values = append(values, ip)
	}

	return values
}

func expandCognitoResourceServerScope(inputs []interface{}) []*cognitoidentityprovider.ResourceServerScopeType {
	configs := make([]*cognitoidentityprovider.ResourceServerScopeType, len(inputs))
	for i, input := range inputs {
		param := input.(map[string]interface{})
		config := &cognitoidentityprovider.ResourceServerScopeType{}

		if v, ok := param["scope_description"]; ok {
			config.ScopeDescription = aws.String(v.(string))
		}

		if v, ok := param["scope_name"]; ok {
			config.ScopeName = aws.String(v.(string))
		}

		configs[i] = config
	}

	return configs
}

func flattenCognitoResourceServerScope(inputs []*cognitoidentityprovider.ResourceServerScopeType) []map[string]interface{} {
	values := make([]map[string]interface{}, 0)

	for _, input := range inputs {
		if input == nil {
			continue
		}
		var value = map[string]interface{}{
			"scope_name":        aws.StringValue(input.ScopeName),
			"scope_description": aws.StringValue(input.ScopeDescription),
		}
		values = append(values, value)
	}
	return values
}

func sliceContainsMap(l []interface{}, m map[string]interface{}) (int, bool) {
	for i, t := range l {
		if reflect.DeepEqual(m, t.(map[string]interface{})) {
			return i, true
		}
	}

	return -1, false
}

func expandAwsSsmTargets(in []interface{}) []*ssm.Target {
	targets := make([]*ssm.Target, 0)

	for _, tConfig := range in {
		config := tConfig.(map[string]interface{})

		target := &ssm.Target{
			Key:    aws.String(config["key"].(string)),
			Values: expandStringList(config["values"].([]interface{})),
		}

		targets = append(targets, target)
	}

	return targets
}

func flattenAwsSsmParameters(parameters map[string][]*string) map[string]string {
	result := make(map[string]string)
	for p, values := range parameters {
		var vs []string
		for _, vPtr := range values {
			if v := aws.StringValue(vPtr); v != "" {
				vs = append(vs, v)
			}
		}
		result[p] = strings.Join(vs, ",")
	}
	return result
}

func flattenAwsSsmTargets(targets []*ssm.Target) []map[string]interface{} {
	if len(targets) == 0 {
		return nil
	}

	result := make([]map[string]interface{}, 0, len(targets))
	for _, target := range targets {
		t := make(map[string]interface{}, 1)
		t["key"] = *target.Key
		t["values"] = flattenStringList(target.Values)

		result = append(result, t)
	}

	return result
}

func expandFieldToMatch(d map[string]interface{}) *waf.FieldToMatch {
	ftm := &waf.FieldToMatch{
		Type: aws.String(d["type"].(string)),
	}
	if data, ok := d["data"].(string); ok && data != "" {
		ftm.Data = aws.String(data)
	}
	return ftm
}

func flattenFieldToMatch(fm *waf.FieldToMatch) []interface{} {
	m := make(map[string]interface{})
	if fm.Data != nil {
		m["data"] = *fm.Data
	}
	if fm.Type != nil {
		m["type"] = *fm.Type
	}
	return []interface{}{m}
}

func diffWafWebAclRules(oldR, newR []interface{}) []*waf.WebACLUpdate {
	updates := make([]*waf.WebACLUpdate, 0)

	for _, or := range oldR {
		aclRule := or.(map[string]interface{})

		if idx, contains := sliceContainsMap(newR, aclRule); contains {
			newR = append(newR[:idx], newR[idx+1:]...)
			continue
		}
		updates = append(updates, expandWafWebAclUpdate(waf.ChangeActionDelete, aclRule))
	}

	for _, nr := range newR {
		aclRule := nr.(map[string]interface{})
		updates = append(updates, expandWafWebAclUpdate(waf.ChangeActionInsert, aclRule))
	}
	return updates
}

func expandWafWebAclUpdate(updateAction string, aclRule map[string]interface{}) *waf.WebACLUpdate {
	var rule *waf.ActivatedRule

	switch aclRule["type"].(string) {
	case waf.WafRuleTypeGroup:
		rule = &waf.ActivatedRule{
			OverrideAction: expandWafOverrideAction(aclRule["override_action"].([]interface{})),
			Priority:       aws.Int64(int64(aclRule["priority"].(int))),
			RuleId:         aws.String(aclRule["rule_id"].(string)),
			Type:           aws.String(aclRule["type"].(string)),
		}
	default:
		rule = &waf.ActivatedRule{
			Action:   expandWafAction(aclRule["action"].([]interface{})),
			Priority: aws.Int64(int64(aclRule["priority"].(int))),
			RuleId:   aws.String(aclRule["rule_id"].(string)),
			Type:     aws.String(aclRule["type"].(string)),
		}
	}

	update := &waf.WebACLUpdate{
		Action:        aws.String(updateAction),
		ActivatedRule: rule,
	}

	return update
}

func expandWafAction(l []interface{}) *waf.WafAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &waf.WafAction{
		Type: aws.String(m["type"].(string)),
	}
}

func expandWafOverrideAction(l []interface{}) *waf.WafOverrideAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &waf.WafOverrideAction{
		Type: aws.String(m["type"].(string)),
	}
}

func flattenWafAction(n *waf.WafAction) []map[string]interface{} {
	if n == nil {
		return nil
	}

	result := map[string]interface{}{
		"type": aws.StringValue(n.Type),
	}

	return []map[string]interface{}{result}
}

func flattenWafWebAclRules(ts []*waf.ActivatedRule) []map[string]interface{} {
	out := make([]map[string]interface{}, len(ts))
	for i, r := range ts {
		m := make(map[string]interface{})

		switch aws.StringValue(r.Type) {
		case waf.WafRuleTypeGroup:
			actionMap := map[string]interface{}{
				"type": aws.StringValue(r.OverrideAction.Type),
			}
			m["override_action"] = []map[string]interface{}{actionMap}
		default:
			actionMap := map[string]interface{}{
				"type": aws.StringValue(r.Action.Type),
			}
			m["action"] = []map[string]interface{}{actionMap}
		}

		m["priority"] = int(aws.Int64Value(r.Priority))
		m["rule_id"] = aws.StringValue(r.RuleId)
		m["type"] = aws.StringValue(r.Type)
		out[i] = m
	}
	return out
}

func flattenWorkLinkNetworkConfigResponse(c *worklink.DescribeCompanyNetworkConfigurationOutput) []map[string]interface{} {
	config := make(map[string]interface{})

	if c == nil {
		return nil
	}

	if len(c.SubnetIds) == 0 && len(c.SecurityGroupIds) == 0 && aws.StringValue(c.VpcId) == "" {
		return nil
	}

	config["subnet_ids"] = flattenStringSet(c.SubnetIds)
	config["security_group_ids"] = flattenStringSet(c.SecurityGroupIds)
	config["vpc_id"] = aws.StringValue(c.VpcId)

	return []map[string]interface{}{config}
}

func flattenWorkLinkIdentityProviderConfigResponse(c *worklink.DescribeIdentityProviderConfigurationOutput) []map[string]interface{} {
	config := make(map[string]interface{})

	if c.IdentityProviderType == nil && c.IdentityProviderSamlMetadata == nil {
		return nil
	}

	if c.IdentityProviderType != nil {
		config["type"] = aws.StringValue(c.IdentityProviderType)
	}
	if c.IdentityProviderSamlMetadata != nil {
		config["saml_metadata"] = aws.StringValue(c.IdentityProviderSamlMetadata)
	}

	return []map[string]interface{}{config}
}

// escapeJsonPointer escapes string per RFC 6901
// so it can be used as path in JSON patch operations
func escapeJsonPointer(path string) string {
	path = strings.Replace(path, "~", "~0", -1)
	path = strings.Replace(path, "/", "~1", -1)
	return path
}

// Like ec2.GroupIdentifier but with additional rule description.
type GroupIdentifier struct {
	// The ID of the security group.
	GroupId *string

	// The name of the security group.
	GroupName *string

	Description *string
}

func expandCognitoIdentityPoolRoles(config map[string]interface{}) map[string]*string {
	m := map[string]*string{}
	for k, v := range config {
		s := v.(string)
		m[k] = &s
	}
	return m
}

func flattenCognitoIdentityPoolRoles(config map[string]*string) map[string]string {
	m := map[string]string{}
	for k, v := range config {
		m[k] = *v
	}
	return m
}

func expandCognitoIdentityPoolRoleMappingsAttachment(rms []interface{}) map[string]*cognitoidentity.RoleMapping {
	values := make(map[string]*cognitoidentity.RoleMapping)

	if len(rms) == 0 {
		return values
	}

	for _, v := range rms {
		rm := v.(map[string]interface{})
		key := rm["identity_provider"].(string)

		roleMapping := &cognitoidentity.RoleMapping{
			Type: aws.String(rm["type"].(string)),
		}

		if sv, ok := rm["ambiguous_role_resolution"].(string); ok {
			roleMapping.AmbiguousRoleResolution = aws.String(sv)
		}

		if mr, ok := rm["mapping_rule"].([]interface{}); ok && len(mr) > 0 {
			rct := &cognitoidentity.RulesConfigurationType{}
			mappingRules := make([]*cognitoidentity.MappingRule, 0)

			for _, r := range mr {
				rule := r.(map[string]interface{})
				mr := &cognitoidentity.MappingRule{
					Claim:     aws.String(rule["claim"].(string)),
					MatchType: aws.String(rule["match_type"].(string)),
					RoleARN:   aws.String(rule["role_arn"].(string)),
					Value:     aws.String(rule["value"].(string)),
				}

				mappingRules = append(mappingRules, mr)
			}

			rct.Rules = mappingRules
			roleMapping.RulesConfiguration = rct
		}

		values[key] = roleMapping
	}

	return values
}

func flattenCognitoIdentityPoolRoleMappingsAttachment(rms map[string]*cognitoidentity.RoleMapping) []map[string]interface{} {
	roleMappings := make([]map[string]interface{}, 0)

	if rms == nil {
		return roleMappings
	}

	for k, v := range rms {
		m := make(map[string]interface{})

		if v == nil {
			return nil
		}

		if v.Type != nil {
			m["type"] = *v.Type
		}

		if v.AmbiguousRoleResolution != nil {
			m["ambiguous_role_resolution"] = *v.AmbiguousRoleResolution
		}

		if v.RulesConfiguration != nil && v.RulesConfiguration.Rules != nil {
			m["mapping_rule"] = flattenCognitoIdentityPoolRolesAttachmentMappingRules(v.RulesConfiguration.Rules)
		}

		m["identity_provider"] = k
		roleMappings = append(roleMappings, m)
	}

	return roleMappings
}

func flattenCognitoIdentityPoolRolesAttachmentMappingRules(d []*cognitoidentity.MappingRule) []interface{} {
	rules := make([]interface{}, 0)

	for _, rule := range d {
		r := make(map[string]interface{})
		r["claim"] = *rule.Claim
		r["match_type"] = *rule.MatchType
		r["role_arn"] = *rule.RoleARN
		r["value"] = *rule.Value

		rules = append(rules, r)
	}

	return rules
}

func flattenRedshiftLogging(ls *redshift.LoggingStatus) []interface{} {
	if ls == nil {
		return []interface{}{}
	}

	cfg := make(map[string]interface{})
	cfg["enable"] = aws.BoolValue(ls.LoggingEnabled)
	if ls.BucketName != nil {
		cfg["bucket_name"] = *ls.BucketName
	}
	if ls.S3KeyPrefix != nil {
		cfg["s3_key_prefix"] = *ls.S3KeyPrefix
	}
	return []interface{}{cfg}
}

func flattenRedshiftSnapshotCopy(scs *redshift.ClusterSnapshotCopyStatus) []interface{} {
	if scs == nil {
		return []interface{}{}
	}

	cfg := make(map[string]interface{})
	if scs.DestinationRegion != nil {
		cfg["destination_region"] = *scs.DestinationRegion
	}
	if scs.RetentionPeriod != nil {
		cfg["retention_period"] = *scs.RetentionPeriod
	}
	if scs.SnapshotCopyGrantName != nil {
		cfg["grant_name"] = *scs.SnapshotCopyGrantName
	}

	return []interface{}{cfg}
}

// cannonicalXML reads XML in a string and re-writes it canonically, used for
// comparing XML for logical equivalency
func canonicalXML(s string) (string, error) {
	doc := etree.NewDocument()
	doc.WriteSettings.CanonicalEndTags = true
	if err := doc.ReadFromString(s); err != nil {
		return "", err
	}

	rawString, err := doc.WriteToString()
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`\s`)
	results := re.ReplaceAllString(rawString, "")
	return results, nil
}

func flattenResourceLifecycleConfig(rlc *elasticbeanstalk.ApplicationResourceLifecycleConfig) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)

	anything_enabled := false
	appversion_lifecycle := make(map[string]interface{})

	if rlc.ServiceRole != nil {
		appversion_lifecycle["service_role"] = *rlc.ServiceRole
	}

	if vlc := rlc.VersionLifecycleConfig; vlc != nil {
		if mar := vlc.MaxAgeRule; mar != nil && *mar.Enabled {
			anything_enabled = true
			appversion_lifecycle["max_age_in_days"] = *mar.MaxAgeInDays
			appversion_lifecycle["delete_source_from_s3"] = *mar.DeleteSourceFromS3
		}
		if mcr := vlc.MaxCountRule; mcr != nil && *mcr.Enabled {
			anything_enabled = true
			appversion_lifecycle["max_count"] = *mcr.MaxCount
			appversion_lifecycle["delete_source_from_s3"] = *mcr.DeleteSourceFromS3
		}
	}

	if anything_enabled {
		result = append(result, appversion_lifecycle)
	}

	return result
}

func stripNonKeyAttributes(in map[string]interface{}) (map[string]interface{}, error) {
	mapCopy, err := copystructure.Copy(in)
	if err != nil {
		return nil, err
	}

	m := mapCopy.(map[string]interface{})

	delete(m, "non_key_attributes")

	return m, nil
}

// checkIfNonKeyAttributesChanged returns true if non_key_attributes between old map and new map are different
func checkIfNonKeyAttributesChanged(oldMap, newMap map[string]interface{}) bool {
	oldNonKeyAttributes, oldNkaExists := oldMap["non_key_attributes"].(*schema.Set)
	newNonKeyAttributes, newNkaExists := newMap["non_key_attributes"].(*schema.Set)

	if oldNkaExists && newNkaExists {
		return !oldNonKeyAttributes.Equal(newNonKeyAttributes)
	}

	return oldNkaExists != newNkaExists
}

func stripCapacityAttributes(in map[string]interface{}) (map[string]interface{}, error) {
	mapCopy, err := copystructure.Copy(in)
	if err != nil {
		return nil, err
	}

	m := mapCopy.(map[string]interface{})

	delete(m, "write_capacity")
	delete(m, "read_capacity")

	return m, nil
}

// Expanders + flatteners

func expandDynamoDbTableItemAttributes(input string) (map[string]*dynamodb.AttributeValue, error) {
	var attributes map[string]*dynamodb.AttributeValue

	dec := json.NewDecoder(strings.NewReader(input))
	err := dec.Decode(&attributes)
	if err != nil {
		return nil, fmt.Errorf("Decoding failed: %s", err)
	}

	return attributes, nil
}

func flattenDynamoDbTableItemAttributes(attrs map[string]*dynamodb.AttributeValue) (string, error) {
	buf := bytes.NewBufferString("")
	encoder := json.NewEncoder(buf)
	err := encoder.Encode(attrs)
	if err != nil {
		return "", fmt.Errorf("Encoding failed: %s", err)
	}

	var rawData map[string]map[string]interface{}

	// Reserialize so we get rid of the nulls
	decoder := json.NewDecoder(strings.NewReader(buf.String()))
	err = decoder.Decode(&rawData)
	if err != nil {
		return "", fmt.Errorf("Decoding failed: %s", err)
	}

	for _, value := range rawData {
		for typeName, typeVal := range value {
			if typeVal == nil {
				delete(value, typeName)
			}
		}
	}

	rawBuffer := bytes.NewBufferString("")
	rawEncoder := json.NewEncoder(rawBuffer)
	err = rawEncoder.Encode(rawData)
	if err != nil {
		return "", fmt.Errorf("Re-encoding failed: %s", err)
	}

	return rawBuffer.String(), nil
}

func expandIotThingTypeProperties(config map[string]interface{}) *iot.ThingTypeProperties {
	properties := &iot.ThingTypeProperties{
		SearchableAttributes: expandStringSet(config["searchable_attributes"].(*schema.Set)),
	}

	if v, ok := config["description"]; ok && v.(string) != "" {
		properties.ThingTypeDescription = aws.String(v.(string))
	}

	return properties
}

func flattenIotThingTypeProperties(s *iot.ThingTypeProperties) []map[string]interface{} {
	m := map[string]interface{}{
		"description":           "",
		"searchable_attributes": flattenStringSet(nil),
	}

	if s == nil {
		return []map[string]interface{}{m}
	}

	m["description"] = aws.StringValue(s.ThingTypeDescription)
	m["searchable_attributes"] = flattenStringSet(s.SearchableAttributes)

	return []map[string]interface{}{m}
}

func expandLaunchTemplateSpecification(specs []interface{}) (*autoscaling.LaunchTemplateSpecification, error) {
	if len(specs) < 1 {
		return nil, nil
	}

	spec := specs[0].(map[string]interface{})

	idValue, idOk := spec["id"]
	nameValue, nameOk := spec["name"]

	if idValue == "" && nameValue == "" {
		return nil, fmt.Errorf("One of `id` or `name` must be set for `launch_template`")
	}

	result := &autoscaling.LaunchTemplateSpecification{}

	// DescribeAutoScalingGroups returns both name and id but LaunchTemplateSpecification
	// allows only one of them to be set
	if idOk && idValue != "" {
		result.LaunchTemplateId = aws.String(idValue.(string))
	} else if nameOk && nameValue != "" {
		result.LaunchTemplateName = aws.String(nameValue.(string))
	}

	if v, ok := spec["version"]; ok && v != "" {
		result.Version = aws.String(v.(string))
	}

	return result, nil
}

func flattenLaunchTemplateSpecification(lt *autoscaling.LaunchTemplateSpecification) []map[string]interface{} {
	if lt == nil {
		return []map[string]interface{}{}
	}

	attrs := map[string]interface{}{}
	result := make([]map[string]interface{}, 0)

	// id and name are always returned by DescribeAutoscalingGroups
	attrs["id"] = *lt.LaunchTemplateId
	attrs["name"] = *lt.LaunchTemplateName

	// version is returned only if it was previosly set
	if lt.Version != nil {
		attrs["version"] = *lt.Version
	} else {
		attrs["version"] = nil
	}

	result = append(result, attrs)

	return result
}

func flattenVpcPeeringConnectionOptions(options *ec2.VpcPeeringConnectionOptionsDescription) []interface{} {
	// When the VPC Peering Connection is pending acceptance,
	// the details about accepter and/or requester peering
	// options would not be included in the response.
	if options == nil {
		return []interface{}{}
	}

	return []interface{}{map[string]interface{}{
		"allow_remote_vpc_dns_resolution":  aws.BoolValue(options.AllowDnsResolutionFromRemoteVpc),
		"allow_classic_link_to_remote_vpc": aws.BoolValue(options.AllowEgressFromLocalClassicLinkToRemoteVpc),
		"allow_vpc_to_remote_classic_link": aws.BoolValue(options.AllowEgressFromLocalVpcToRemoteClassicLink),
	}}
}

func expandVpcPeeringConnectionOptions(vOptions []interface{}, crossRegionPeering bool) *ec2.PeeringConnectionOptionsRequest {
	if len(vOptions) == 0 || vOptions[0] == nil {
		return nil
	}

	mOptions := vOptions[0].(map[string]interface{})

	options := &ec2.PeeringConnectionOptionsRequest{}

	if v, ok := mOptions["allow_remote_vpc_dns_resolution"].(bool); ok {
		options.AllowDnsResolutionFromRemoteVpc = aws.Bool(v)
	}
	if !crossRegionPeering {
		if v, ok := mOptions["allow_classic_link_to_remote_vpc"].(bool); ok {
			options.AllowEgressFromLocalClassicLinkToRemoteVpc = aws.Bool(v)
		}
		if v, ok := mOptions["allow_vpc_to_remote_classic_link"].(bool); ok {
			options.AllowEgressFromLocalVpcToRemoteClassicLink = aws.Bool(v)
		}
	}

	return options
}

func expandDxRouteFilterPrefixes(vPrefixes *schema.Set) []*directconnect.RouteFilterPrefix {
	routeFilterPrefixes := []*directconnect.RouteFilterPrefix{}

	for _, vPrefix := range vPrefixes.List() {
		routeFilterPrefixes = append(routeFilterPrefixes, &directconnect.RouteFilterPrefix{
			Cidr: aws.String(vPrefix.(string)),
		})
	}

	return routeFilterPrefixes
}

func flattenDxRouteFilterPrefixes(routeFilterPrefixes []*directconnect.RouteFilterPrefix) *schema.Set {
	vPrefixes := []interface{}{}

	for _, routeFilterPrefix := range routeFilterPrefixes {
		vPrefixes = append(vPrefixes, aws.StringValue(routeFilterPrefix.Cidr))
	}

	return schema.NewSet(schema.HashString, vPrefixes)
}

func expandMacieClassificationType(d *schema.ResourceData) *macie.ClassificationType {
	continuous := macie.S3ContinuousClassificationTypeFull
	oneTime := macie.S3OneTimeClassificationTypeNone
	if v := d.Get("classification_type").([]interface{}); len(v) > 0 {
		m := v[0].(map[string]interface{})
		continuous = m["continuous"].(string)
		oneTime = m["one_time"].(string)
	}

	return &macie.ClassificationType{
		Continuous: aws.String(continuous),
		OneTime:    aws.String(oneTime),
	}
}

func expandMacieClassificationTypeUpdate(d *schema.ResourceData) *macie.ClassificationTypeUpdate {
	continuous := macie.S3ContinuousClassificationTypeFull
	oneTime := macie.S3OneTimeClassificationTypeNone
	if v := d.Get("classification_type").([]interface{}); len(v) > 0 {
		m := v[0].(map[string]interface{})
		continuous = m["continuous"].(string)
		oneTime = m["one_time"].(string)
	}

	return &macie.ClassificationTypeUpdate{
		Continuous: aws.String(continuous),
		OneTime:    aws.String(oneTime),
	}
}

func flattenMacieClassificationType(classificationType *macie.ClassificationType) []map[string]interface{} {
	if classificationType == nil {
		return []map[string]interface{}{}
	}
	m := map[string]interface{}{
		"continuous": aws.StringValue(classificationType.Continuous),
		"one_time":   aws.StringValue(classificationType.OneTime),
	}
	return []map[string]interface{}{m}
}

func expandDaxParameterGroupParameterNameValue(config []interface{}) []*dax.ParameterNameValue {
	if len(config) == 0 {
		return nil
	}
	results := make([]*dax.ParameterNameValue, 0, len(config))
	for _, raw := range config {
		m := raw.(map[string]interface{})
		pnv := &dax.ParameterNameValue{
			ParameterName:  aws.String(m["name"].(string)),
			ParameterValue: aws.String(m["value"].(string)),
		}
		results = append(results, pnv)
	}
	return results
}

func flattenDaxParameterGroupParameters(params []*dax.Parameter) []map[string]interface{} {
	if len(params) == 0 {
		return nil
	}
	results := make([]map[string]interface{}, 0)
	for _, p := range params {
		m := map[string]interface{}{
			"name":  aws.StringValue(p.ParameterName),
			"value": aws.StringValue(p.ParameterValue),
		}
		results = append(results, m)
	}
	return results
}

func expandDaxEncryptAtRestOptions(m map[string]interface{}) *dax.SSESpecification {
	options := dax.SSESpecification{}

	if v, ok := m["enabled"]; ok {
		options.Enabled = aws.Bool(v.(bool))
	}

	return &options
}

func flattenDaxEncryptAtRestOptions(options *dax.SSEDescription) []map[string]interface{} {
	m := map[string]interface{}{
		"enabled": false,
	}

	if options == nil {
		return []map[string]interface{}{m}
	}

	m["enabled"] = (aws.StringValue(options.Status) == dax.SSEStatusEnabled)

	return []map[string]interface{}{m}
}

func expandRdsClusterScalingConfiguration(l []interface{}) *rds.ScalingConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	scalingConfiguration := &rds.ScalingConfiguration{
		AutoPause:             aws.Bool(m["auto_pause"].(bool)),
		MaxCapacity:           aws.Int64(int64(m["max_capacity"].(int))),
		MinCapacity:           aws.Int64(int64(m["min_capacity"].(int))),
		SecondsUntilAutoPause: aws.Int64(int64(m["seconds_until_auto_pause"].(int))),
	}

	if vTimeoutAction, ok := m["timeout_action"].(string); ok && vTimeoutAction != "" {
		scalingConfiguration.TimeoutAction = aws.String(vTimeoutAction)
	}

	return scalingConfiguration
}

func flattenRdsScalingConfigurationInfo(scalingConfigurationInfo *rds.ScalingConfigurationInfo) []interface{} {
	if scalingConfigurationInfo == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"auto_pause":               aws.BoolValue(scalingConfigurationInfo.AutoPause),
		"max_capacity":             aws.Int64Value(scalingConfigurationInfo.MaxCapacity),
		"min_capacity":             aws.Int64Value(scalingConfigurationInfo.MinCapacity),
		"seconds_until_auto_pause": aws.Int64Value(scalingConfigurationInfo.SecondsUntilAutoPause),
		"timeout_action":           aws.StringValue(scalingConfigurationInfo.TimeoutAction),
	}

	return []interface{}{m}
}

func expandAppmeshMeshSpec(vSpec []interface{}) *appmesh.MeshSpec {
	spec := &appmesh.MeshSpec{}

	if len(vSpec) == 0 || vSpec[0] == nil {
		// Empty Spec is allowed.
		return spec
	}
	mSpec := vSpec[0].(map[string]interface{})

	if vEgressFilter, ok := mSpec["egress_filter"].([]interface{}); ok && len(vEgressFilter) > 0 && vEgressFilter[0] != nil {
		mEgressFilter := vEgressFilter[0].(map[string]interface{})

		if vType, ok := mEgressFilter["type"].(string); ok && vType != "" {
			spec.EgressFilter = &appmesh.EgressFilter{
				Type: aws.String(vType),
			}
		}
	}

	return spec
}

func flattenAppmeshMeshSpec(spec *appmesh.MeshSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	mSpec := map[string]interface{}{}

	if spec.EgressFilter != nil {
		mSpec["egress_filter"] = []interface{}{
			map[string]interface{}{
				"type": aws.StringValue(spec.EgressFilter.Type),
			},
		}
	}

	return []interface{}{mSpec}
}

func expandAppmeshVirtualRouterSpec(vSpec []interface{}) *appmesh.VirtualRouterSpec {
	spec := &appmesh.VirtualRouterSpec{}

	if len(vSpec) == 0 || vSpec[0] == nil {
		// Empty Spec is allowed.
		return spec
	}
	mSpec := vSpec[0].(map[string]interface{})

	if vListeners, ok := mSpec["listener"].([]interface{}); ok && len(vListeners) > 0 && vListeners[0] != nil {
		listeners := []*appmesh.VirtualRouterListener{}

		for _, vListener := range vListeners {
			listener := &appmesh.VirtualRouterListener{}

			mListener := vListener.(map[string]interface{})

			if vPortMapping, ok := mListener["port_mapping"].([]interface{}); ok && len(vPortMapping) > 0 && vPortMapping[0] != nil {
				mPortMapping := vPortMapping[0].(map[string]interface{})

				listener.PortMapping = &appmesh.PortMapping{}

				if vPort, ok := mPortMapping["port"].(int); ok && vPort > 0 {
					listener.PortMapping.Port = aws.Int64(int64(vPort))
				}
				if vProtocol, ok := mPortMapping["protocol"].(string); ok && vProtocol != "" {
					listener.PortMapping.Protocol = aws.String(vProtocol)
				}
			}
			listeners = append(listeners, listener)
		}
		spec.Listeners = listeners
	}

	return spec
}

func flattenAppmeshVirtualRouterSpec(spec *appmesh.VirtualRouterSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}
	mSpec := make(map[string]interface{})
	if spec.Listeners != nil && spec.Listeners[0] != nil {
		// Per schema definition, set at most 1 Listener
		listener := spec.Listeners[0]
		mListener := make(map[string]interface{})
		if listener.PortMapping != nil {
			mPortMapping := map[string]interface{}{
				"port":     int(aws.Int64Value(listener.PortMapping.Port)),
				"protocol": aws.StringValue(listener.PortMapping.Protocol),
			}
			mListener["port_mapping"] = []interface{}{mPortMapping}
		}
		mSpec["listener"] = []interface{}{mListener}
	}

	return []interface{}{mSpec}
}

func expandAppmeshVirtualNodeSpec(vSpec []interface{}) *appmesh.VirtualNodeSpec {
	spec := &appmesh.VirtualNodeSpec{}

	if len(vSpec) == 0 || vSpec[0] == nil {
		// Empty Spec is allowed.
		return spec
	}
	mSpec := vSpec[0].(map[string]interface{})

	if vBackends, ok := mSpec["backend"].(*schema.Set); ok && vBackends.Len() > 0 {
		backends := []*appmesh.Backend{}

		for _, vBackend := range vBackends.List() {
			backend := &appmesh.Backend{}

			mBackend := vBackend.(map[string]interface{})

			if vVirtualService, ok := mBackend["virtual_service"].([]interface{}); ok && len(vVirtualService) > 0 && vVirtualService[0] != nil {
				virtualService := &appmesh.VirtualServiceBackend{}

				mVirtualService := vVirtualService[0].(map[string]interface{})

				if vVirtualServiceName, ok := mVirtualService["virtual_service_name"].(string); ok {
					virtualService.VirtualServiceName = aws.String(vVirtualServiceName)
				}

				if vClientPolicy, ok := mVirtualService["client_policy"].([]interface{}); ok {
					virtualService.ClientPolicy = expandAppmeshClientPolicy(vClientPolicy)
				}

				backend.VirtualService = virtualService
			}

			backends = append(backends, backend)
		}

		spec.Backends = backends
	}

	if vBackendDefaults, ok := mSpec["backend_defaults"].([]interface{}); ok && len(vBackendDefaults) > 0 && vBackendDefaults[0] != nil {
		backendDefaults := &appmesh.BackendDefaults{}

		mBackendDefaults := vBackendDefaults[0].(map[string]interface{})

		if vClientPolicy, ok := mBackendDefaults["client_policy"].([]interface{}); ok {
			backendDefaults.ClientPolicy = expandAppmeshClientPolicy(vClientPolicy)
		}

		spec.BackendDefaults = backendDefaults
	}

	if vListeners, ok := mSpec["listener"].([]interface{}); ok && len(vListeners) > 0 && vListeners[0] != nil {
		listeners := []*appmesh.Listener{}

		for _, vListener := range vListeners {
			listener := &appmesh.Listener{}

			mListener := vListener.(map[string]interface{})

			if vConnectionPool, ok := mListener["connection_pool"].([]interface{}); ok && len(vConnectionPool) > 0 && vConnectionPool[0] != nil {
				mConnectionPool := vConnectionPool[0].(map[string]interface{})

				connectionPool := &appmesh.VirtualNodeConnectionPool{}

				if vGrpcConnectionPool, ok := mConnectionPool["grpc"].([]interface{}); ok && len(vGrpcConnectionPool) > 0 && vGrpcConnectionPool[0] != nil {
					mGrpcConnectionPool := vGrpcConnectionPool[0].(map[string]interface{})

					grpcConnectionPool := &appmesh.VirtualNodeGrpcConnectionPool{}

					if vMaxRequests, ok := mGrpcConnectionPool["max_requests"].(int); ok && vMaxRequests > 0 {
						grpcConnectionPool.MaxRequests = aws.Int64(int64(vMaxRequests))
					}

					connectionPool.Grpc = grpcConnectionPool
				}

				if vHttpConnectionPool, ok := mConnectionPool["http"].([]interface{}); ok && len(vHttpConnectionPool) > 0 && vHttpConnectionPool[0] != nil {
					mHttpConnectionPool := vHttpConnectionPool[0].(map[string]interface{})

					httpConnectionPool := &appmesh.VirtualNodeHttpConnectionPool{}

					if vMaxConnections, ok := mHttpConnectionPool["max_connections"].(int); ok && vMaxConnections > 0 {
						httpConnectionPool.MaxConnections = aws.Int64(int64(vMaxConnections))
					}
					if vMaxPendingRequests, ok := mHttpConnectionPool["max_pending_requests"].(int); ok && vMaxPendingRequests > 0 {
						httpConnectionPool.MaxPendingRequests = aws.Int64(int64(vMaxPendingRequests))
					}

					connectionPool.Http = httpConnectionPool
				}

				if vHttp2ConnectionPool, ok := mConnectionPool["http2"].([]interface{}); ok && len(vHttp2ConnectionPool) > 0 && vHttp2ConnectionPool[0] != nil {
					mHttp2ConnectionPool := vHttp2ConnectionPool[0].(map[string]interface{})

					http2ConnectionPool := &appmesh.VirtualNodeHttp2ConnectionPool{}

					if vMaxRequests, ok := mHttp2ConnectionPool["max_requests"].(int); ok && vMaxRequests > 0 {
						http2ConnectionPool.MaxRequests = aws.Int64(int64(vMaxRequests))
					}

					connectionPool.Http2 = http2ConnectionPool
				}

				if vTcpConnectionPool, ok := mConnectionPool["tcp"].([]interface{}); ok && len(vTcpConnectionPool) > 0 && vTcpConnectionPool[0] != nil {
					mTcpConnectionPool := vTcpConnectionPool[0].(map[string]interface{})

					tcpConnectionPool := &appmesh.VirtualNodeTcpConnectionPool{}

					if vMaxConnections, ok := mTcpConnectionPool["max_connections"].(int); ok && vMaxConnections > 0 {
						tcpConnectionPool.MaxConnections = aws.Int64(int64(vMaxConnections))
					}

					connectionPool.Tcp = tcpConnectionPool
				}

				listener.ConnectionPool = connectionPool
			}

			if vHealthCheck, ok := mListener["health_check"].([]interface{}); ok && len(vHealthCheck) > 0 && vHealthCheck[0] != nil {
				healthCheck := &appmesh.HealthCheckPolicy{}

				mHealthCheck := vHealthCheck[0].(map[string]interface{})

				if vHealthyThreshold, ok := mHealthCheck["healthy_threshold"].(int); ok && vHealthyThreshold > 0 {
					healthCheck.HealthyThreshold = aws.Int64(int64(vHealthyThreshold))
				}
				if vIntervalMillis, ok := mHealthCheck["interval_millis"].(int); ok && vIntervalMillis > 0 {
					healthCheck.IntervalMillis = aws.Int64(int64(vIntervalMillis))
				}
				if vPath, ok := mHealthCheck["path"].(string); ok && vPath != "" {
					healthCheck.Path = aws.String(vPath)
				}
				if vPort, ok := mHealthCheck["port"].(int); ok && vPort > 0 {
					healthCheck.Port = aws.Int64(int64(vPort))
				}
				if vProtocol, ok := mHealthCheck["protocol"].(string); ok && vProtocol != "" {
					healthCheck.Protocol = aws.String(vProtocol)
				}
				if vTimeoutMillis, ok := mHealthCheck["timeout_millis"].(int); ok && vTimeoutMillis > 0 {
					healthCheck.TimeoutMillis = aws.Int64(int64(vTimeoutMillis))
				}
				if vUnhealthyThreshold, ok := mHealthCheck["unhealthy_threshold"].(int); ok && vUnhealthyThreshold > 0 {
					healthCheck.UnhealthyThreshold = aws.Int64(int64(vUnhealthyThreshold))
				}

				listener.HealthCheck = healthCheck
			}

			if vOutlierDetection, ok := mListener["outlier_detection"].([]interface{}); ok && len(vOutlierDetection) > 0 && vOutlierDetection[0] != nil {
				outlierDetection := &appmesh.OutlierDetection{}

				mOutlierDetection := vOutlierDetection[0].(map[string]interface{})

				if vMaxEjectionPercent, ok := mOutlierDetection["max_ejection_percent"].(int); ok && vMaxEjectionPercent > 0 {
					outlierDetection.MaxEjectionPercent = aws.Int64(int64(vMaxEjectionPercent))
				}
				if vMaxServerErrors, ok := mOutlierDetection["max_server_errors"].(int); ok && vMaxServerErrors > 0 {
					outlierDetection.MaxServerErrors = aws.Int64(int64(vMaxServerErrors))
				}

				if vBaseEjectionDuration, ok := mOutlierDetection["base_ejection_duration"].([]interface{}); ok {
					outlierDetection.BaseEjectionDuration = expandAppmeshDuration(vBaseEjectionDuration)
				}

				if vInterval, ok := mOutlierDetection["interval"].([]interface{}); ok {
					outlierDetection.Interval = expandAppmeshDuration(vInterval)
				}

				listener.OutlierDetection = outlierDetection
			}

			if vPortMapping, ok := mListener["port_mapping"].([]interface{}); ok && len(vPortMapping) > 0 && vPortMapping[0] != nil {
				portMapping := &appmesh.PortMapping{}

				mPortMapping := vPortMapping[0].(map[string]interface{})

				if vPort, ok := mPortMapping["port"].(int); ok && vPort > 0 {
					portMapping.Port = aws.Int64(int64(vPort))
				}
				if vProtocol, ok := mPortMapping["protocol"].(string); ok && vProtocol != "" {
					portMapping.Protocol = aws.String(vProtocol)
				}

				listener.PortMapping = portMapping
			}

			if vTimeout, ok := mListener["timeout"].([]interface{}); ok && len(vTimeout) > 0 && vTimeout[0] != nil {
				mTimeout := vTimeout[0].(map[string]interface{})

				listenerTimeout := &appmesh.ListenerTimeout{}

				if vGrpcTimeout, ok := mTimeout["grpc"].([]interface{}); ok {
					listenerTimeout.Grpc = expandAppmeshGrpcTimeout(vGrpcTimeout)
				}

				if vHttpTimeout, ok := mTimeout["http"].([]interface{}); ok {
					listenerTimeout.Http = expandAppmeshHttpTimeout(vHttpTimeout)
				}

				if vHttp2Timeout, ok := mTimeout["http2"].([]interface{}); ok {
					listenerTimeout.Http2 = expandAppmeshHttpTimeout(vHttp2Timeout)
				}

				if vTcpTimeout, ok := mTimeout["tcp"].([]interface{}); ok {
					listenerTimeout.Tcp = expandAppmeshTcpTimeout(vTcpTimeout)
				}

				listener.Timeout = listenerTimeout
			}

			if vTls, ok := mListener["tls"].([]interface{}); ok && len(vTls) > 0 && vTls[0] != nil {
				tls := &appmesh.ListenerTls{}

				mTls := vTls[0].(map[string]interface{})

				if vMode, ok := mTls["mode"].(string); ok && vMode != "" {
					tls.Mode = aws.String(vMode)
				}

				if vCertificate, ok := mTls["certificate"].([]interface{}); ok && len(vCertificate) > 0 && vCertificate[0] != nil {
					certificate := &appmesh.ListenerTlsCertificate{}

					mCertificate := vCertificate[0].(map[string]interface{})

					if vAcm, ok := mCertificate["acm"].([]interface{}); ok && len(vAcm) > 0 && vAcm[0] != nil {
						acm := &appmesh.ListenerTlsAcmCertificate{}

						mAcm := vAcm[0].(map[string]interface{})

						if vCertificateArn, ok := mAcm["certificate_arn"].(string); ok && vCertificateArn != "" {
							acm.CertificateArn = aws.String(vCertificateArn)
						}

						certificate.Acm = acm
					}

					if vFile, ok := mCertificate["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
						file := &appmesh.ListenerTlsFileCertificate{}

						mFile := vFile[0].(map[string]interface{})

						if vCertificateChain, ok := mFile["certificate_chain"].(string); ok && vCertificateChain != "" {
							file.CertificateChain = aws.String(vCertificateChain)
						}
						if vPrivateKey, ok := mFile["private_key"].(string); ok && vPrivateKey != "" {
							file.PrivateKey = aws.String(vPrivateKey)
						}

						certificate.File = file
					}

					if vSds, ok := mCertificate["sds"].([]interface{}); ok && len(vSds) > 0 && vSds[0] != nil {
						sds := &appmesh.ListenerTlsSdsCertificate{}

						mSds := vSds[0].(map[string]interface{})

						if vSecretName, ok := mSds["secret_name"].(string); ok && vSecretName != "" {
							sds.SecretName = aws.String(vSecretName)
						}

						certificate.Sds = sds
					}

					tls.Certificate = certificate
				}

				if vValidation, ok := mTls["validation"].([]interface{}); ok && len(vValidation) > 0 && vValidation[0] != nil {
					validation := &appmesh.ListenerTlsValidationContext{}

					mValidation := vValidation[0].(map[string]interface{})

					if vSubjectAlternativeNames, ok := mValidation["subject_alternative_names"].([]interface{}); ok && len(vSubjectAlternativeNames) > 0 && vSubjectAlternativeNames[0] != nil {
						subjectAlternativeNames := &appmesh.SubjectAlternativeNames{}

						mSubjectAlternativeNames := vSubjectAlternativeNames[0].(map[string]interface{})

						if vMatch, ok := mSubjectAlternativeNames["match"].([]interface{}); ok && len(vMatch) > 0 && vMatch[0] != nil {
							match := &appmesh.SubjectAlternativeNameMatchers{}

							mMatch := vMatch[0].(map[string]interface{})

							if vExact, ok := mMatch["exact"].(*schema.Set); ok && vExact.Len() > 0 {
								match.Exact = expandStringSet(vExact)
							}

							subjectAlternativeNames.Match = match
						}

						validation.SubjectAlternativeNames = subjectAlternativeNames
					}

					if vTrust, ok := mValidation["trust"].([]interface{}); ok && len(vTrust) > 0 && vTrust[0] != nil {
						trust := &appmesh.ListenerTlsValidationContextTrust{}

						mTrust := vTrust[0].(map[string]interface{})

						if vFile, ok := mTrust["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
							file := &appmesh.TlsValidationContextFileTrust{}

							mFile := vFile[0].(map[string]interface{})

							if vCertificateChain, ok := mFile["certificate_chain"].(string); ok && vCertificateChain != "" {
								file.CertificateChain = aws.String(vCertificateChain)
							}

							trust.File = file
						}

						if vSds, ok := mTrust["sds"].([]interface{}); ok && len(vSds) > 0 && vSds[0] != nil {
							sds := &appmesh.TlsValidationContextSdsTrust{}

							mSds := vSds[0].(map[string]interface{})

							if vSecretName, ok := mSds["secret_name"].(string); ok && vSecretName != "" {
								sds.SecretName = aws.String(vSecretName)
							}

							trust.Sds = sds
						}

						validation.Trust = trust
					}

					tls.Validation = validation
				}

				listener.Tls = tls
			}

			listeners = append(listeners, listener)
		}

		spec.Listeners = listeners
	}

	if vLogging, ok := mSpec["logging"].([]interface{}); ok && len(vLogging) > 0 && vLogging[0] != nil {
		logging := &appmesh.Logging{}

		mLogging := vLogging[0].(map[string]interface{})

		if vAccessLog, ok := mLogging["access_log"].([]interface{}); ok && len(vAccessLog) > 0 && vAccessLog[0] != nil {
			accessLog := &appmesh.AccessLog{}

			mAccessLog := vAccessLog[0].(map[string]interface{})

			if vFile, ok := mAccessLog["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
				file := &appmesh.FileAccessLog{}

				mFile := vFile[0].(map[string]interface{})

				if vPath, ok := mFile["path"].(string); ok && vPath != "" {
					file.Path = aws.String(vPath)
				}

				accessLog.File = file
			}

			logging.AccessLog = accessLog
		}

		spec.Logging = logging
	}

	if vServiceDiscovery, ok := mSpec["service_discovery"].([]interface{}); ok && len(vServiceDiscovery) > 0 && vServiceDiscovery[0] != nil {
		serviceDiscovery := &appmesh.ServiceDiscovery{}

		mServiceDiscovery := vServiceDiscovery[0].(map[string]interface{})

		if vAwsCloudMap, ok := mServiceDiscovery["aws_cloud_map"].([]interface{}); ok && len(vAwsCloudMap) > 0 && vAwsCloudMap[0] != nil {
			awsCloudMap := &appmesh.AwsCloudMapServiceDiscovery{}

			mAwsCloudMap := vAwsCloudMap[0].(map[string]interface{})

			if vAttributes, ok := mAwsCloudMap["attributes"].(map[string]interface{}); ok && len(vAttributes) > 0 {
				attributes := []*appmesh.AwsCloudMapInstanceAttribute{}

				for k, v := range vAttributes {
					attributes = append(attributes, &appmesh.AwsCloudMapInstanceAttribute{
						Key:   aws.String(k),
						Value: aws.String(v.(string)),
					})
				}

				awsCloudMap.Attributes = attributes
			}
			if vNamespaceName, ok := mAwsCloudMap["namespace_name"].(string); ok && vNamespaceName != "" {
				awsCloudMap.NamespaceName = aws.String(vNamespaceName)
			}
			if vServiceName, ok := mAwsCloudMap["service_name"].(string); ok && vServiceName != "" {
				awsCloudMap.ServiceName = aws.String(vServiceName)
			}

			serviceDiscovery.AwsCloudMap = awsCloudMap
		}

		if vDns, ok := mServiceDiscovery["dns"].([]interface{}); ok && len(vDns) > 0 && vDns[0] != nil {
			dns := &appmesh.DnsServiceDiscovery{}

			mDns := vDns[0].(map[string]interface{})

			if vHostname, ok := mDns["hostname"].(string); ok && vHostname != "" {
				dns.Hostname = aws.String(vHostname)
			}

			serviceDiscovery.Dns = dns
		}

		spec.ServiceDiscovery = serviceDiscovery
	}

	return spec
}

func flattenAppmeshVirtualNodeSpec(spec *appmesh.VirtualNodeSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	mSpec := map[string]interface{}{}

	if backends := spec.Backends; backends != nil {
		vBackends := []interface{}{}

		for _, backend := range backends {
			mBackend := map[string]interface{}{}

			if virtualService := backend.VirtualService; virtualService != nil {
				mVirtualService := map[string]interface{}{
					"client_policy":        flattenAppmeshClientPolicy(virtualService.ClientPolicy),
					"virtual_service_name": aws.StringValue(virtualService.VirtualServiceName),
				}

				mBackend["virtual_service"] = []interface{}{mVirtualService}
			}

			vBackends = append(vBackends, mBackend)
		}

		mSpec["backend"] = vBackends
	}

	if backendDefaults := spec.BackendDefaults; backendDefaults != nil {
		mBackendDefaults := map[string]interface{}{
			"client_policy": flattenAppmeshClientPolicy(backendDefaults.ClientPolicy),
		}

		mSpec["backend_defaults"] = []interface{}{mBackendDefaults}
	}

	if spec.Listeners != nil && spec.Listeners[0] != nil {
		// Per schema definition, set at most 1 Listener
		listener := spec.Listeners[0]
		mListener := map[string]interface{}{}

		if connectionPool := listener.ConnectionPool; connectionPool != nil {
			mConnectionPool := map[string]interface{}{}

			if grpcConnectionPool := connectionPool.Grpc; grpcConnectionPool != nil {
				mGrpcConnectionPool := map[string]interface{}{
					"max_requests": int(aws.Int64Value(grpcConnectionPool.MaxRequests)),
				}
				mConnectionPool["grpc"] = []interface{}{mGrpcConnectionPool}
			}

			if httpConnectionPool := connectionPool.Http; httpConnectionPool != nil {
				mHttpConnectionPool := map[string]interface{}{
					"max_connections":      int(aws.Int64Value(httpConnectionPool.MaxConnections)),
					"max_pending_requests": int(aws.Int64Value(httpConnectionPool.MaxPendingRequests)),
				}
				mConnectionPool["http"] = []interface{}{mHttpConnectionPool}
			}

			if http2ConnectionPool := connectionPool.Http2; http2ConnectionPool != nil {
				mHttp2ConnectionPool := map[string]interface{}{
					"max_requests": int(aws.Int64Value(http2ConnectionPool.MaxRequests)),
				}
				mConnectionPool["http2"] = []interface{}{mHttp2ConnectionPool}
			}

			if tcpConnectionPool := connectionPool.Tcp; tcpConnectionPool != nil {
				mTcpConnectionPool := map[string]interface{}{
					"max_connections": int(aws.Int64Value(tcpConnectionPool.MaxConnections)),
				}
				mConnectionPool["tcp"] = []interface{}{mTcpConnectionPool}
			}

			mListener["connection_pool"] = []interface{}{mConnectionPool}
		}

		if healthCheck := listener.HealthCheck; healthCheck != nil {
			mHealthCheck := map[string]interface{}{
				"healthy_threshold":   int(aws.Int64Value(healthCheck.HealthyThreshold)),
				"interval_millis":     int(aws.Int64Value(healthCheck.IntervalMillis)),
				"path":                aws.StringValue(healthCheck.Path),
				"port":                int(aws.Int64Value(healthCheck.Port)),
				"protocol":            aws.StringValue(healthCheck.Protocol),
				"timeout_millis":      int(aws.Int64Value(healthCheck.TimeoutMillis)),
				"unhealthy_threshold": int(aws.Int64Value(healthCheck.UnhealthyThreshold)),
			}
			mListener["health_check"] = []interface{}{mHealthCheck}
		}

		if outlierDetection := listener.OutlierDetection; outlierDetection != nil {
			mOutlierDetection := map[string]interface{}{
				"base_ejection_duration": flattenAppmeshDuration(outlierDetection.BaseEjectionDuration),
				"interval":               flattenAppmeshDuration(outlierDetection.Interval),
				"max_ejection_percent":   int(aws.Int64Value(outlierDetection.MaxEjectionPercent)),
				"max_server_errors":      int(aws.Int64Value(outlierDetection.MaxServerErrors)),
			}
			mListener["outlier_detection"] = []interface{}{mOutlierDetection}
		}

		if portMapping := listener.PortMapping; portMapping != nil {
			mPortMapping := map[string]interface{}{
				"port":     int(aws.Int64Value(portMapping.Port)),
				"protocol": aws.StringValue(portMapping.Protocol),
			}
			mListener["port_mapping"] = []interface{}{mPortMapping}
		}

		if listenerTimeout := listener.Timeout; listenerTimeout != nil {
			mListenerTimeout := map[string]interface{}{
				"grpc":  flattenAppmeshGrpcTimeout(listenerTimeout.Grpc),
				"http":  flattenAppmeshHttpTimeout(listenerTimeout.Http),
				"http2": flattenAppmeshHttpTimeout(listenerTimeout.Http2),
				"tcp":   flattenAppmeshTcpTimeout(listenerTimeout.Tcp),
			}
			mListener["timeout"] = []interface{}{mListenerTimeout}
		}

		if tls := listener.Tls; tls != nil {
			mTls := map[string]interface{}{
				"mode": aws.StringValue(tls.Mode),
			}

			if certificate := tls.Certificate; certificate != nil {
				mCertificate := map[string]interface{}{}

				if acm := certificate.Acm; acm != nil {
					mAcm := map[string]interface{}{
						"certificate_arn": aws.StringValue(acm.CertificateArn),
					}

					mCertificate["acm"] = []interface{}{mAcm}
				}

				if file := certificate.File; file != nil {
					mFile := map[string]interface{}{
						"certificate_chain": aws.StringValue(file.CertificateChain),
						"private_key":       aws.StringValue(file.PrivateKey),
					}

					mCertificate["file"] = []interface{}{mFile}
				}

				if sds := certificate.Sds; sds != nil {
					mSds := map[string]interface{}{
						"secret_name": aws.StringValue(sds.SecretName),
					}

					mCertificate["sds"] = []interface{}{mSds}
				}

				mTls["certificate"] = []interface{}{mCertificate}
			}

			if validation := tls.Validation; validation != nil {
				mValidation := map[string]interface{}{}

				if subjectAlternativeNames := validation.SubjectAlternativeNames; subjectAlternativeNames != nil {
					mSubjectAlternativeNames := map[string]interface{}{}

					if match := subjectAlternativeNames.Match; match != nil {
						mMatch := map[string]interface{}{
							"exact": flattenStringSet(match.Exact),
						}

						mSubjectAlternativeNames["match"] = []interface{}{mMatch}
					}

					mValidation["subject_alternative_names"] = []interface{}{mSubjectAlternativeNames}
				}

				if trust := validation.Trust; trust != nil {
					mTrust := map[string]interface{}{}

					if file := trust.File; file != nil {
						mFile := map[string]interface{}{
							"certificate_chain": aws.StringValue(file.CertificateChain),
						}

						mTrust["file"] = []interface{}{mFile}
					}

					if sds := trust.Sds; sds != nil {
						mSds := map[string]interface{}{
							"secret_name": aws.StringValue(sds.SecretName),
						}

						mTrust["sds"] = []interface{}{mSds}
					}

					mValidation["trust"] = []interface{}{mTrust}
				}

				mTls["validation"] = []interface{}{mValidation}
			}

			mListener["tls"] = []interface{}{mTls}
		}

		mSpec["listener"] = []interface{}{mListener}
	}

	if logging := spec.Logging; logging != nil {
		mLogging := map[string]interface{}{}

		if accessLog := logging.AccessLog; accessLog != nil {
			mAccessLog := map[string]interface{}{}

			if file := accessLog.File; file != nil {
				mAccessLog["file"] = []interface{}{
					map[string]interface{}{
						"path": aws.StringValue(file.Path),
					},
				}
			}

			mLogging["access_log"] = []interface{}{mAccessLog}
		}

		mSpec["logging"] = []interface{}{mLogging}
	}

	if serviceDiscovery := spec.ServiceDiscovery; serviceDiscovery != nil {
		mServiceDiscovery := map[string]interface{}{}

		if awsCloudMap := serviceDiscovery.AwsCloudMap; awsCloudMap != nil {
			vAttributes := map[string]interface{}{}

			for _, attribute := range awsCloudMap.Attributes {
				vAttributes[aws.StringValue(attribute.Key)] = aws.StringValue(attribute.Value)
			}

			mServiceDiscovery["aws_cloud_map"] = []interface{}{
				map[string]interface{}{
					"attributes":     vAttributes,
					"namespace_name": aws.StringValue(awsCloudMap.NamespaceName),
					"service_name":   aws.StringValue(awsCloudMap.ServiceName),
				},
			}
		}

		if dns := serviceDiscovery.Dns; dns != nil {
			mServiceDiscovery["dns"] = []interface{}{
				map[string]interface{}{
					"hostname": aws.StringValue(dns.Hostname),
				},
			}
		}

		mSpec["service_discovery"] = []interface{}{mServiceDiscovery}
	}

	return []interface{}{mSpec}
}

func expandAppmeshClientPolicy(vClientPolicy []interface{}) *appmesh.ClientPolicy {
	if len(vClientPolicy) == 0 || vClientPolicy[0] == nil {
		return nil
	}

	clientPolicy := &appmesh.ClientPolicy{}

	mClientPolicy := vClientPolicy[0].(map[string]interface{})

	if vTls, ok := mClientPolicy["tls"].([]interface{}); ok && len(vTls) > 0 && vTls[0] != nil {
		tls := &appmesh.ClientPolicyTls{}

		mTls := vTls[0].(map[string]interface{})

		if vCertificate, ok := mTls["certificate"].([]interface{}); ok && len(vCertificate) > 0 && vCertificate[0] != nil {
			certificate := &appmesh.ClientTlsCertificate{}

			mCertificate := vCertificate[0].(map[string]interface{})

			if vFile, ok := mCertificate["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
				file := &appmesh.ListenerTlsFileCertificate{}

				mFile := vFile[0].(map[string]interface{})

				if vCertificateChain, ok := mFile["certificate_chain"].(string); ok && vCertificateChain != "" {
					file.CertificateChain = aws.String(vCertificateChain)
				}
				if vPrivateKey, ok := mFile["private_key"].(string); ok && vPrivateKey != "" {
					file.PrivateKey = aws.String(vPrivateKey)
				}

				certificate.File = file
			}

			if vSds, ok := mCertificate["sds"].([]interface{}); ok && len(vSds) > 0 && vSds[0] != nil {
				sds := &appmesh.ListenerTlsSdsCertificate{}

				mSds := vSds[0].(map[string]interface{})

				if vSecretName, ok := mSds["secret_name"].(string); ok && vSecretName != "" {
					sds.SecretName = aws.String(vSecretName)
				}

				certificate.Sds = sds
			}

			tls.Certificate = certificate
		}

		if vEnforce, ok := mTls["enforce"].(bool); ok {
			tls.Enforce = aws.Bool(vEnforce)
		}

		if vPorts, ok := mTls["ports"].(*schema.Set); ok && vPorts.Len() > 0 {
			tls.Ports = expandInt64Set(vPorts)
		}

		if vValidation, ok := mTls["validation"].([]interface{}); ok && len(vValidation) > 0 && vValidation[0] != nil {
			validation := &appmesh.TlsValidationContext{}

			mValidation := vValidation[0].(map[string]interface{})

			if vSubjectAlternativeNames, ok := mValidation["subject_alternative_names"].([]interface{}); ok && len(vSubjectAlternativeNames) > 0 && vSubjectAlternativeNames[0] != nil {
				subjectAlternativeNames := &appmesh.SubjectAlternativeNames{}

				mSubjectAlternativeNames := vSubjectAlternativeNames[0].(map[string]interface{})

				if vMatch, ok := mSubjectAlternativeNames["match"].([]interface{}); ok && len(vMatch) > 0 && vMatch[0] != nil {
					match := &appmesh.SubjectAlternativeNameMatchers{}

					mMatch := vMatch[0].(map[string]interface{})

					if vExact, ok := mMatch["exact"].(*schema.Set); ok && vExact.Len() > 0 {
						match.Exact = expandStringSet(vExact)
					}

					subjectAlternativeNames.Match = match
				}

				validation.SubjectAlternativeNames = subjectAlternativeNames
			}

			if vTrust, ok := mValidation["trust"].([]interface{}); ok && len(vTrust) > 0 && vTrust[0] != nil {
				trust := &appmesh.TlsValidationContextTrust{}

				mTrust := vTrust[0].(map[string]interface{})

				if vAcm, ok := mTrust["acm"].([]interface{}); ok && len(vAcm) > 0 && vAcm[0] != nil {
					acm := &appmesh.TlsValidationContextAcmTrust{}

					mAcm := vAcm[0].(map[string]interface{})

					if vCertificateAuthorityArns, ok := mAcm["certificate_authority_arns"].(*schema.Set); ok && vCertificateAuthorityArns.Len() > 0 {
						acm.CertificateAuthorityArns = expandStringSet(vCertificateAuthorityArns)
					}

					trust.Acm = acm
				}

				if vFile, ok := mTrust["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
					file := &appmesh.TlsValidationContextFileTrust{}

					mFile := vFile[0].(map[string]interface{})

					if vCertificateChain, ok := mFile["certificate_chain"].(string); ok && vCertificateChain != "" {
						file.CertificateChain = aws.String(vCertificateChain)
					}

					trust.File = file
				}

				if vSds, ok := mTrust["sds"].([]interface{}); ok && len(vSds) > 0 && vSds[0] != nil {
					sds := &appmesh.TlsValidationContextSdsTrust{}

					mSds := vSds[0].(map[string]interface{})

					if vSecretName, ok := mSds["secret_name"].(string); ok && vSecretName != "" {
						sds.SecretName = aws.String(vSecretName)
					}

					trust.Sds = sds
				}

				validation.Trust = trust
			}

			tls.Validation = validation
		}

		clientPolicy.Tls = tls
	}

	return clientPolicy
}

func flattenAppmeshClientPolicy(clientPolicy *appmesh.ClientPolicy) []interface{} {
	if clientPolicy == nil {
		return []interface{}{}
	}

	mClientPolicy := map[string]interface{}{}

	if tls := clientPolicy.Tls; tls != nil {
		mTls := map[string]interface{}{
			"enforce": aws.BoolValue(tls.Enforce),
			"ports":   flattenInt64Set(tls.Ports),
		}

		if certificate := tls.Certificate; certificate != nil {
			mCertificate := map[string]interface{}{}

			if file := certificate.File; file != nil {
				mFile := map[string]interface{}{
					"certificate_chain": aws.StringValue(file.CertificateChain),
					"private_key":       aws.StringValue(file.PrivateKey),
				}

				mCertificate["file"] = []interface{}{mFile}
			}

			if sds := certificate.Sds; sds != nil {
				mSds := map[string]interface{}{
					"secret_name": aws.StringValue(sds.SecretName),
				}

				mCertificate["sds"] = []interface{}{mSds}
			}

			mTls["certificate"] = []interface{}{mCertificate}
		}

		if validation := tls.Validation; validation != nil {
			mValidation := map[string]interface{}{}

			if subjectAlternativeNames := validation.SubjectAlternativeNames; subjectAlternativeNames != nil {
				mSubjectAlternativeNames := map[string]interface{}{}

				if match := subjectAlternativeNames.Match; match != nil {
					mMatch := map[string]interface{}{
						"exact": flattenStringSet(match.Exact),
					}

					mSubjectAlternativeNames["match"] = []interface{}{mMatch}
				}

				mValidation["subject_alternative_names"] = []interface{}{mSubjectAlternativeNames}
			}

			if trust := validation.Trust; trust != nil {
				mTrust := map[string]interface{}{}

				if acm := trust.Acm; acm != nil {
					mAcm := map[string]interface{}{
						"certificate_authority_arns": flattenStringSet(acm.CertificateAuthorityArns),
					}

					mTrust["acm"] = []interface{}{mAcm}
				}

				if file := trust.File; file != nil {
					mFile := map[string]interface{}{
						"certificate_chain": aws.StringValue(file.CertificateChain),
					}

					mTrust["file"] = []interface{}{mFile}
				}

				if sds := trust.Sds; sds != nil {
					mSds := map[string]interface{}{
						"secret_name": aws.StringValue(sds.SecretName),
					}

					mTrust["sds"] = []interface{}{mSds}
				}

				mValidation["trust"] = []interface{}{mTrust}
			}

			mTls["validation"] = []interface{}{mValidation}
		}

		mClientPolicy["tls"] = []interface{}{mTls}
	}

	return []interface{}{mClientPolicy}
}

func expandAppmeshVirtualServiceSpec(vSpec []interface{}) *appmesh.VirtualServiceSpec {
	spec := &appmesh.VirtualServiceSpec{}

	if len(vSpec) == 0 || vSpec[0] == nil {
		// Empty Spec is allowed.
		return spec
	}
	mSpec := vSpec[0].(map[string]interface{})

	if vProvider, ok := mSpec["provider"].([]interface{}); ok && len(vProvider) > 0 && vProvider[0] != nil {
		mProvider := vProvider[0].(map[string]interface{})

		spec.Provider = &appmesh.VirtualServiceProvider{}

		if vVirtualNode, ok := mProvider["virtual_node"].([]interface{}); ok && len(vVirtualNode) > 0 && vVirtualNode[0] != nil {
			mVirtualNode := vVirtualNode[0].(map[string]interface{})

			if vVirtualNodeName, ok := mVirtualNode["virtual_node_name"].(string); ok && vVirtualNodeName != "" {
				spec.Provider.VirtualNode = &appmesh.VirtualNodeServiceProvider{
					VirtualNodeName: aws.String(vVirtualNodeName),
				}
			}
		}

		if vVirtualRouter, ok := mProvider["virtual_router"].([]interface{}); ok && len(vVirtualRouter) > 0 && vVirtualRouter[0] != nil {
			mVirtualRouter := vVirtualRouter[0].(map[string]interface{})

			if vVirtualRouterName, ok := mVirtualRouter["virtual_router_name"].(string); ok && vVirtualRouterName != "" {
				spec.Provider.VirtualRouter = &appmesh.VirtualRouterServiceProvider{
					VirtualRouterName: aws.String(vVirtualRouterName),
				}
			}
		}
	}

	return spec
}

func flattenAppmeshVirtualServiceSpec(spec *appmesh.VirtualServiceSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	mSpec := map[string]interface{}{}

	if spec.Provider != nil {
		mProvider := map[string]interface{}{}

		if spec.Provider.VirtualNode != nil {
			mProvider["virtual_node"] = []interface{}{
				map[string]interface{}{
					"virtual_node_name": aws.StringValue(spec.Provider.VirtualNode.VirtualNodeName),
				},
			}
		}

		if spec.Provider.VirtualRouter != nil {
			mProvider["virtual_router"] = []interface{}{
				map[string]interface{}{
					"virtual_router_name": aws.StringValue(spec.Provider.VirtualRouter.VirtualRouterName),
				},
			}
		}

		mSpec["provider"] = []interface{}{mProvider}
	}

	return []interface{}{mSpec}
}

func expandAppmeshRouteSpec(vSpec []interface{}) *appmesh.RouteSpec {
	spec := &appmesh.RouteSpec{}

	if len(vSpec) == 0 || vSpec[0] == nil {
		// Empty Spec is allowed.
		return spec
	}
	mSpec := vSpec[0].(map[string]interface{})

	if vGrpcRoute, ok := mSpec["grpc_route"].([]interface{}); ok {
		spec.GrpcRoute = expandAppmeshGrpcRoute(vGrpcRoute)
	}

	if vHttp2Route, ok := mSpec["http2_route"].([]interface{}); ok {
		spec.Http2Route = expandAppmeshHttpRoute(vHttp2Route)
	}

	if vHttpRoute, ok := mSpec["http_route"].([]interface{}); ok {
		spec.HttpRoute = expandAppmeshHttpRoute(vHttpRoute)
	}

	if vPriority, ok := mSpec["priority"].(int); ok && vPriority > 0 {
		spec.Priority = aws.Int64(int64(vPriority))
	}

	if vTcpRoute, ok := mSpec["tcp_route"].([]interface{}); ok {
		spec.TcpRoute = expandAppmeshTcpRoute(vTcpRoute)
	}

	return spec
}

func expandAppmeshGrpcRoute(vGrpcRoute []interface{}) *appmesh.GrpcRoute {
	if len(vGrpcRoute) == 0 || vGrpcRoute[0] == nil {
		return nil
	}

	mGrpcRoute := vGrpcRoute[0].(map[string]interface{})

	grpcRoute := &appmesh.GrpcRoute{}

	if vGrpcRouteAction, ok := mGrpcRoute["action"].([]interface{}); ok && len(vGrpcRouteAction) > 0 && vGrpcRouteAction[0] != nil {
		mGrpcRouteAction := vGrpcRouteAction[0].(map[string]interface{})

		if vWeightedTargets, ok := mGrpcRouteAction["weighted_target"].(*schema.Set); ok && vWeightedTargets.Len() > 0 {
			weightedTargets := []*appmesh.WeightedTarget{}

			for _, vWeightedTarget := range vWeightedTargets.List() {
				weightedTarget := &appmesh.WeightedTarget{}

				mWeightedTarget := vWeightedTarget.(map[string]interface{})

				if vVirtualNode, ok := mWeightedTarget["virtual_node"].(string); ok && vVirtualNode != "" {
					weightedTarget.VirtualNode = aws.String(vVirtualNode)
				}
				if vWeight, ok := mWeightedTarget["weight"].(int); ok {
					weightedTarget.Weight = aws.Int64(int64(vWeight))
				}

				weightedTargets = append(weightedTargets, weightedTarget)
			}

			grpcRoute.Action = &appmesh.GrpcRouteAction{
				WeightedTargets: weightedTargets,
			}
		}
	}

	if vGrpcRouteMatch, ok := mGrpcRoute["match"].([]interface{}); ok {
		grpcRouteMatch := &appmesh.GrpcRouteMatch{}

		// Empty match is allowed.
		// https://github.com/hashicorp/terraform-provider-aws/issues/16816.

		if len(vGrpcRouteMatch) > 0 && vGrpcRouteMatch[0] != nil {
			mGrpcRouteMatch := vGrpcRouteMatch[0].(map[string]interface{})

			if vMethodName, ok := mGrpcRouteMatch["method_name"].(string); ok && vMethodName != "" {
				grpcRouteMatch.MethodName = aws.String(vMethodName)
			}
			if vServiceName, ok := mGrpcRouteMatch["service_name"].(string); ok && vServiceName != "" {
				grpcRouteMatch.ServiceName = aws.String(vServiceName)
			}

			if vGrpcRouteMetadatas, ok := mGrpcRouteMatch["metadata"].(*schema.Set); ok && vGrpcRouteMetadatas.Len() > 0 {
				grpcRouteMetadatas := []*appmesh.GrpcRouteMetadata{}

				for _, vGrpcRouteMetadata := range vGrpcRouteMetadatas.List() {
					grpcRouteMetadata := &appmesh.GrpcRouteMetadata{}

					mGrpcRouteMetadata := vGrpcRouteMetadata.(map[string]interface{})

					if vInvert, ok := mGrpcRouteMetadata["invert"].(bool); ok {
						grpcRouteMetadata.Invert = aws.Bool(vInvert)
					}
					if vName, ok := mGrpcRouteMetadata["name"].(string); ok && vName != "" {
						grpcRouteMetadata.Name = aws.String(vName)
					}

					if vMatch, ok := mGrpcRouteMetadata["match"].([]interface{}); ok && len(vMatch) > 0 && vMatch[0] != nil {
						grpcRouteMetadata.Match = &appmesh.GrpcRouteMetadataMatchMethod{}

						mMatch := vMatch[0].(map[string]interface{})

						if vExact, ok := mMatch["exact"].(string); ok && vExact != "" {
							grpcRouteMetadata.Match.Exact = aws.String(vExact)
						}
						if vPrefix, ok := mMatch["prefix"].(string); ok && vPrefix != "" {
							grpcRouteMetadata.Match.Prefix = aws.String(vPrefix)
						}
						if vRegex, ok := mMatch["regex"].(string); ok && vRegex != "" {
							grpcRouteMetadata.Match.Regex = aws.String(vRegex)
						}
						if vSuffix, ok := mMatch["suffix"].(string); ok && vSuffix != "" {
							grpcRouteMetadata.Match.Suffix = aws.String(vSuffix)
						}

						if vRange, ok := mMatch["range"].([]interface{}); ok && len(vRange) > 0 && vRange[0] != nil {
							grpcRouteMetadata.Match.Range = &appmesh.MatchRange{}

							mRange := vRange[0].(map[string]interface{})

							if vEnd, ok := mRange["end"].(int); ok && vEnd > 0 {
								grpcRouteMetadata.Match.Range.End = aws.Int64(int64(vEnd))
							}
							if vStart, ok := mRange["start"].(int); ok && vStart > 0 {
								grpcRouteMetadata.Match.Range.Start = aws.Int64(int64(vStart))
							}
						}
					}

					grpcRouteMetadatas = append(grpcRouteMetadatas, grpcRouteMetadata)
				}

				grpcRouteMatch.Metadata = grpcRouteMetadatas
			}
		}

		grpcRoute.Match = grpcRouteMatch
	}

	if vGrpcRetryPolicy, ok := mGrpcRoute["retry_policy"].([]interface{}); ok && len(vGrpcRetryPolicy) > 0 && vGrpcRetryPolicy[0] != nil {
		grpcRetryPolicy := &appmesh.GrpcRetryPolicy{}

		mGrpcRetryPolicy := vGrpcRetryPolicy[0].(map[string]interface{})

		if vMaxRetries, ok := mGrpcRetryPolicy["max_retries"].(int); ok && vMaxRetries > 0 {
			grpcRetryPolicy.MaxRetries = aws.Int64(int64(vMaxRetries))
		}

		if vGrpcRetryEvents, ok := mGrpcRetryPolicy["grpc_retry_events"].(*schema.Set); ok && vGrpcRetryEvents.Len() > 0 {
			grpcRetryPolicy.GrpcRetryEvents = expandStringSet(vGrpcRetryEvents)
		}

		if vHttpRetryEvents, ok := mGrpcRetryPolicy["http_retry_events"].(*schema.Set); ok && vHttpRetryEvents.Len() > 0 {
			grpcRetryPolicy.HttpRetryEvents = expandStringSet(vHttpRetryEvents)
		}

		if vPerRetryTimeout, ok := mGrpcRetryPolicy["per_retry_timeout"].([]interface{}); ok {
			grpcRetryPolicy.PerRetryTimeout = expandAppmeshDuration(vPerRetryTimeout)
		}

		if vTcpRetryEvents, ok := mGrpcRetryPolicy["tcp_retry_events"].(*schema.Set); ok && vTcpRetryEvents.Len() > 0 {
			grpcRetryPolicy.TcpRetryEvents = expandStringSet(vTcpRetryEvents)
		}

		grpcRoute.RetryPolicy = grpcRetryPolicy
	}

	if vGrpcTimeout, ok := mGrpcRoute["timeout"].([]interface{}); ok {
		grpcRoute.Timeout = expandAppmeshGrpcTimeout(vGrpcTimeout)
	}

	return grpcRoute
}

func expandAppmeshGrpcTimeout(vGrpcTimeout []interface{}) *appmesh.GrpcTimeout {
	if len(vGrpcTimeout) == 0 || vGrpcTimeout[0] == nil {
		return nil
	}

	grpcTimeout := &appmesh.GrpcTimeout{}

	mGrpcTimeout := vGrpcTimeout[0].(map[string]interface{})

	if vIdleTimeout, ok := mGrpcTimeout["idle"].([]interface{}); ok {
		grpcTimeout.Idle = expandAppmeshDuration(vIdleTimeout)
	}

	if vPerRequestTimeout, ok := mGrpcTimeout["per_request"].([]interface{}); ok {
		grpcTimeout.PerRequest = expandAppmeshDuration(vPerRequestTimeout)
	}

	return grpcTimeout
}

func expandAppmeshHttpRoute(vHttpRoute []interface{}) *appmesh.HttpRoute {
	if len(vHttpRoute) == 0 || vHttpRoute[0] == nil {
		return nil
	}

	mHttpRoute := vHttpRoute[0].(map[string]interface{})

	httpRoute := &appmesh.HttpRoute{}

	if vHttpRouteAction, ok := mHttpRoute["action"].([]interface{}); ok && len(vHttpRouteAction) > 0 && vHttpRouteAction[0] != nil {
		mHttpRouteAction := vHttpRouteAction[0].(map[string]interface{})

		if vWeightedTargets, ok := mHttpRouteAction["weighted_target"].(*schema.Set); ok && vWeightedTargets.Len() > 0 {
			weightedTargets := []*appmesh.WeightedTarget{}

			for _, vWeightedTarget := range vWeightedTargets.List() {
				weightedTarget := &appmesh.WeightedTarget{}

				mWeightedTarget := vWeightedTarget.(map[string]interface{})

				if vVirtualNode, ok := mWeightedTarget["virtual_node"].(string); ok && vVirtualNode != "" {
					weightedTarget.VirtualNode = aws.String(vVirtualNode)
				}
				if vWeight, ok := mWeightedTarget["weight"].(int); ok {
					weightedTarget.Weight = aws.Int64(int64(vWeight))
				}

				weightedTargets = append(weightedTargets, weightedTarget)
			}

			httpRoute.Action = &appmesh.HttpRouteAction{
				WeightedTargets: weightedTargets,
			}
		}
	}

	if vHttpRouteMatch, ok := mHttpRoute["match"].([]interface{}); ok && len(vHttpRouteMatch) > 0 && vHttpRouteMatch[0] != nil {
		httpRouteMatch := &appmesh.HttpRouteMatch{}

		mHttpRouteMatch := vHttpRouteMatch[0].(map[string]interface{})

		if vMethod, ok := mHttpRouteMatch["method"].(string); ok && vMethod != "" {
			httpRouteMatch.Method = aws.String(vMethod)
		}
		if vPrefix, ok := mHttpRouteMatch["prefix"].(string); ok && vPrefix != "" {
			httpRouteMatch.Prefix = aws.String(vPrefix)
		}
		if vScheme, ok := mHttpRouteMatch["scheme"].(string); ok && vScheme != "" {
			httpRouteMatch.Scheme = aws.String(vScheme)
		}

		if vHttpRouteHeaders, ok := mHttpRouteMatch["header"].(*schema.Set); ok && vHttpRouteHeaders.Len() > 0 {
			httpRouteHeaders := []*appmesh.HttpRouteHeader{}

			for _, vHttpRouteHeader := range vHttpRouteHeaders.List() {
				httpRouteHeader := &appmesh.HttpRouteHeader{}

				mHttpRouteHeader := vHttpRouteHeader.(map[string]interface{})

				if vInvert, ok := mHttpRouteHeader["invert"].(bool); ok {
					httpRouteHeader.Invert = aws.Bool(vInvert)
				}
				if vName, ok := mHttpRouteHeader["name"].(string); ok && vName != "" {
					httpRouteHeader.Name = aws.String(vName)
				}

				if vMatch, ok := mHttpRouteHeader["match"].([]interface{}); ok && len(vMatch) > 0 && vMatch[0] != nil {
					httpRouteHeader.Match = &appmesh.HeaderMatchMethod{}

					mMatch := vMatch[0].(map[string]interface{})

					if vExact, ok := mMatch["exact"].(string); ok && vExact != "" {
						httpRouteHeader.Match.Exact = aws.String(vExact)
					}
					if vPrefix, ok := mMatch["prefix"].(string); ok && vPrefix != "" {
						httpRouteHeader.Match.Prefix = aws.String(vPrefix)
					}
					if vRegex, ok := mMatch["regex"].(string); ok && vRegex != "" {
						httpRouteHeader.Match.Regex = aws.String(vRegex)
					}
					if vSuffix, ok := mMatch["suffix"].(string); ok && vSuffix != "" {
						httpRouteHeader.Match.Suffix = aws.String(vSuffix)
					}

					if vRange, ok := mMatch["range"].([]interface{}); ok && len(vRange) > 0 && vRange[0] != nil {
						httpRouteHeader.Match.Range = &appmesh.MatchRange{}

						mRange := vRange[0].(map[string]interface{})

						if vEnd, ok := mRange["end"].(int); ok && vEnd > 0 {
							httpRouteHeader.Match.Range.End = aws.Int64(int64(vEnd))
						}
						if vStart, ok := mRange["start"].(int); ok && vStart > 0 {
							httpRouteHeader.Match.Range.Start = aws.Int64(int64(vStart))
						}
					}
				}

				httpRouteHeaders = append(httpRouteHeaders, httpRouteHeader)
			}

			httpRouteMatch.Headers = httpRouteHeaders
		}

		httpRoute.Match = httpRouteMatch
	}

	if vHttpRetryPolicy, ok := mHttpRoute["retry_policy"].([]interface{}); ok && len(vHttpRetryPolicy) > 0 && vHttpRetryPolicy[0] != nil {
		httpRetryPolicy := &appmesh.HttpRetryPolicy{}

		mHttpRetryPolicy := vHttpRetryPolicy[0].(map[string]interface{})

		if vMaxRetries, ok := mHttpRetryPolicy["max_retries"].(int); ok && vMaxRetries > 0 {
			httpRetryPolicy.MaxRetries = aws.Int64(int64(vMaxRetries))
		}

		if vHttpRetryEvents, ok := mHttpRetryPolicy["http_retry_events"].(*schema.Set); ok && vHttpRetryEvents.Len() > 0 {
			httpRetryPolicy.HttpRetryEvents = expandStringSet(vHttpRetryEvents)
		}

		if vPerRetryTimeout, ok := mHttpRetryPolicy["per_retry_timeout"].([]interface{}); ok {
			httpRetryPolicy.PerRetryTimeout = expandAppmeshDuration(vPerRetryTimeout)
		}

		if vTcpRetryEvents, ok := mHttpRetryPolicy["tcp_retry_events"].(*schema.Set); ok && vTcpRetryEvents.Len() > 0 {
			httpRetryPolicy.TcpRetryEvents = expandStringSet(vTcpRetryEvents)
		}

		httpRoute.RetryPolicy = httpRetryPolicy
	}

	if vHttpTimeout, ok := mHttpRoute["timeout"].([]interface{}); ok {
		httpRoute.Timeout = expandAppmeshHttpTimeout(vHttpTimeout)
	}

	return httpRoute
}

func expandAppmeshHttpTimeout(vHttpTimeout []interface{}) *appmesh.HttpTimeout {
	if len(vHttpTimeout) == 0 || vHttpTimeout[0] == nil {
		return nil
	}

	httpTimeout := &appmesh.HttpTimeout{}

	mHttpTimeout := vHttpTimeout[0].(map[string]interface{})

	if vIdleTimeout, ok := mHttpTimeout["idle"].([]interface{}); ok {
		httpTimeout.Idle = expandAppmeshDuration(vIdleTimeout)
	}

	if vPerRequestTimeout, ok := mHttpTimeout["per_request"].([]interface{}); ok {
		httpTimeout.PerRequest = expandAppmeshDuration(vPerRequestTimeout)
	}

	return httpTimeout
}

func expandAppmeshTcpRoute(vTcpRoute []interface{}) *appmesh.TcpRoute {
	if len(vTcpRoute) == 0 || vTcpRoute[0] == nil {
		return nil
	}

	mTcpRoute := vTcpRoute[0].(map[string]interface{})

	tcpRoute := &appmesh.TcpRoute{}

	if vTcpRouteAction, ok := mTcpRoute["action"].([]interface{}); ok && len(vTcpRouteAction) > 0 && vTcpRouteAction[0] != nil {
		mTcpRouteAction := vTcpRouteAction[0].(map[string]interface{})

		if vWeightedTargets, ok := mTcpRouteAction["weighted_target"].(*schema.Set); ok && vWeightedTargets.Len() > 0 {
			weightedTargets := []*appmesh.WeightedTarget{}

			for _, vWeightedTarget := range vWeightedTargets.List() {
				weightedTarget := &appmesh.WeightedTarget{}

				mWeightedTarget := vWeightedTarget.(map[string]interface{})

				if vVirtualNode, ok := mWeightedTarget["virtual_node"].(string); ok && vVirtualNode != "" {
					weightedTarget.VirtualNode = aws.String(vVirtualNode)
				}
				if vWeight, ok := mWeightedTarget["weight"].(int); ok {
					weightedTarget.Weight = aws.Int64(int64(vWeight))
				}

				weightedTargets = append(weightedTargets, weightedTarget)
			}

			tcpRoute.Action = &appmesh.TcpRouteAction{
				WeightedTargets: weightedTargets,
			}
		}
	}

	if vTcpTimeout, ok := mTcpRoute["timeout"].([]interface{}); ok {
		tcpRoute.Timeout = expandAppmeshTcpTimeout(vTcpTimeout)
	}

	return tcpRoute
}

func expandAppmeshTcpTimeout(vTcpTimeout []interface{}) *appmesh.TcpTimeout {
	if len(vTcpTimeout) == 0 || vTcpTimeout[0] == nil {
		return nil
	}

	tcpTimeout := &appmesh.TcpTimeout{}

	mTcpTimeout := vTcpTimeout[0].(map[string]interface{})

	if vIdleTimeout, ok := mTcpTimeout["idle"].([]interface{}); ok {
		tcpTimeout.Idle = expandAppmeshDuration(vIdleTimeout)
	}

	return tcpTimeout
}

func expandAppmeshDuration(vDuration []interface{}) *appmesh.Duration {
	if len(vDuration) == 0 || vDuration[0] == nil {
		return nil
	}

	duration := &appmesh.Duration{}

	mDuration := vDuration[0].(map[string]interface{})

	if vUnit, ok := mDuration["unit"].(string); ok && vUnit != "" {
		duration.Unit = aws.String(vUnit)
	}
	if vValue, ok := mDuration["value"].(int); ok && vValue > 0 {
		duration.Value = aws.Int64(int64(vValue))
	}

	return duration
}

func flattenAppmeshRouteSpec(spec *appmesh.RouteSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	mSpec := map[string]interface{}{
		"grpc_route":  flattenAppmeshGrpcRoute(spec.GrpcRoute),
		"http2_route": flattenAppmeshHttpRoute(spec.Http2Route),
		"http_route":  flattenAppmeshHttpRoute(spec.HttpRoute),
		"priority":    int(aws.Int64Value(spec.Priority)),
		"tcp_route":   flattenAppmeshTcpRoute(spec.TcpRoute),
	}

	return []interface{}{mSpec}
}

func flattenAppmeshGrpcRoute(grpcRoute *appmesh.GrpcRoute) []interface{} {
	if grpcRoute == nil {
		return []interface{}{}
	}

	mGrpcRoute := map[string]interface{}{}

	if action := grpcRoute.Action; action != nil {
		if weightedTargets := action.WeightedTargets; weightedTargets != nil {
			vWeightedTargets := []interface{}{}

			for _, weightedTarget := range weightedTargets {
				mWeightedTarget := map[string]interface{}{
					"virtual_node": aws.StringValue(weightedTarget.VirtualNode),
					"weight":       int(aws.Int64Value(weightedTarget.Weight)),
				}

				vWeightedTargets = append(vWeightedTargets, mWeightedTarget)
			}

			mGrpcRoute["action"] = []interface{}{
				map[string]interface{}{
					"weighted_target": vWeightedTargets,
				},
			}
		}
	}

	if grpcRouteMatch := grpcRoute.Match; grpcRouteMatch != nil {
		vGrpcRouteMetadatas := []interface{}{}

		for _, grpcRouteMetadata := range grpcRouteMatch.Metadata {
			mGrpcRouteMetadata := map[string]interface{}{
				"invert": aws.BoolValue(grpcRouteMetadata.Invert),
				"name":   aws.StringValue(grpcRouteMetadata.Name),
			}

			if match := grpcRouteMetadata.Match; match != nil {
				mMatch := map[string]interface{}{
					"exact":  aws.StringValue(match.Exact),
					"prefix": aws.StringValue(match.Prefix),
					"regex":  aws.StringValue(match.Regex),
					"suffix": aws.StringValue(match.Suffix),
				}

				if r := match.Range; r != nil {
					mRange := map[string]interface{}{
						"end":   int(aws.Int64Value(r.End)),
						"start": int(aws.Int64Value(r.Start)),
					}

					mMatch["range"] = []interface{}{mRange}
				}

				mGrpcRouteMetadata["match"] = []interface{}{mMatch}
			}

			vGrpcRouteMetadatas = append(vGrpcRouteMetadatas, mGrpcRouteMetadata)
		}

		mGrpcRoute["match"] = []interface{}{
			map[string]interface{}{
				"metadata":     vGrpcRouteMetadatas,
				"method_name":  aws.StringValue(grpcRouteMatch.MethodName),
				"service_name": aws.StringValue(grpcRouteMatch.ServiceName),
			},
		}
	}

	if grpcRetryPolicy := grpcRoute.RetryPolicy; grpcRetryPolicy != nil {
		mGrpcRetryPolicy := map[string]interface{}{
			"grpc_retry_events": flattenStringSet(grpcRetryPolicy.GrpcRetryEvents),
			"http_retry_events": flattenStringSet(grpcRetryPolicy.HttpRetryEvents),
			"max_retries":       int(aws.Int64Value(grpcRetryPolicy.MaxRetries)),
			"per_retry_timeout": flattenAppmeshDuration(grpcRetryPolicy.PerRetryTimeout),
			"tcp_retry_events":  flattenStringSet(grpcRetryPolicy.TcpRetryEvents),
		}

		mGrpcRoute["retry_policy"] = []interface{}{mGrpcRetryPolicy}
	}

	mGrpcRoute["timeout"] = flattenAppmeshGrpcTimeout(grpcRoute.Timeout)

	return []interface{}{mGrpcRoute}
}

func flattenAppmeshGrpcTimeout(grpcTimeout *appmesh.GrpcTimeout) []interface{} {
	if grpcTimeout == nil {
		return []interface{}{}
	}

	mGrpcTimeout := map[string]interface{}{
		"idle":        flattenAppmeshDuration(grpcTimeout.Idle),
		"per_request": flattenAppmeshDuration(grpcTimeout.PerRequest),
	}

	return []interface{}{mGrpcTimeout}
}

func flattenAppmeshHttpRoute(httpRoute *appmesh.HttpRoute) []interface{} {
	if httpRoute == nil {
		return []interface{}{}
	}

	mHttpRoute := map[string]interface{}{}

	if action := httpRoute.Action; action != nil {
		if weightedTargets := action.WeightedTargets; weightedTargets != nil {
			vWeightedTargets := []interface{}{}

			for _, weightedTarget := range weightedTargets {
				mWeightedTarget := map[string]interface{}{
					"virtual_node": aws.StringValue(weightedTarget.VirtualNode),
					"weight":       int(aws.Int64Value(weightedTarget.Weight)),
				}

				vWeightedTargets = append(vWeightedTargets, mWeightedTarget)
			}

			mHttpRoute["action"] = []interface{}{
				map[string]interface{}{
					"weighted_target": vWeightedTargets,
				},
			}
		}
	}

	if httpRouteMatch := httpRoute.Match; httpRouteMatch != nil {
		vHttpRouteHeaders := []interface{}{}

		for _, httpRouteHeader := range httpRouteMatch.Headers {
			mHttpRouteHeader := map[string]interface{}{
				"invert": aws.BoolValue(httpRouteHeader.Invert),
				"name":   aws.StringValue(httpRouteHeader.Name),
			}

			if match := httpRouteHeader.Match; match != nil {
				mMatch := map[string]interface{}{
					"exact":  aws.StringValue(match.Exact),
					"prefix": aws.StringValue(match.Prefix),
					"regex":  aws.StringValue(match.Regex),
					"suffix": aws.StringValue(match.Suffix),
				}

				if r := match.Range; r != nil {
					mRange := map[string]interface{}{
						"end":   int(aws.Int64Value(r.End)),
						"start": int(aws.Int64Value(r.Start)),
					}

					mMatch["range"] = []interface{}{mRange}
				}

				mHttpRouteHeader["match"] = []interface{}{mMatch}
			}

			vHttpRouteHeaders = append(vHttpRouteHeaders, mHttpRouteHeader)
		}

		mHttpRoute["match"] = []interface{}{
			map[string]interface{}{
				"header": vHttpRouteHeaders,
				"method": aws.StringValue(httpRouteMatch.Method),
				"prefix": aws.StringValue(httpRouteMatch.Prefix),
				"scheme": aws.StringValue(httpRouteMatch.Scheme),
			},
		}
	}

	if httpRetryPolicy := httpRoute.RetryPolicy; httpRetryPolicy != nil {
		mHttpRetryPolicy := map[string]interface{}{
			"http_retry_events": flattenStringSet(httpRetryPolicy.HttpRetryEvents),
			"max_retries":       int(aws.Int64Value(httpRetryPolicy.MaxRetries)),
			"per_retry_timeout": flattenAppmeshDuration(httpRetryPolicy.PerRetryTimeout),
			"tcp_retry_events":  flattenStringSet(httpRetryPolicy.TcpRetryEvents),
		}

		mHttpRoute["retry_policy"] = []interface{}{mHttpRetryPolicy}
	}

	mHttpRoute["timeout"] = flattenAppmeshHttpTimeout(httpRoute.Timeout)

	return []interface{}{mHttpRoute}
}

func flattenAppmeshHttpTimeout(httpTimeout *appmesh.HttpTimeout) []interface{} {
	if httpTimeout == nil {
		return []interface{}{}
	}

	mHttpTimeout := map[string]interface{}{
		"idle":        flattenAppmeshDuration(httpTimeout.Idle),
		"per_request": flattenAppmeshDuration(httpTimeout.PerRequest),
	}

	return []interface{}{mHttpTimeout}
}

func flattenAppmeshTcpRoute(tcpRoute *appmesh.TcpRoute) []interface{} {
	if tcpRoute == nil {
		return []interface{}{}
	}

	mTcpRoute := map[string]interface{}{}

	if action := tcpRoute.Action; action != nil {
		if weightedTargets := action.WeightedTargets; weightedTargets != nil {
			vWeightedTargets := []interface{}{}

			for _, weightedTarget := range weightedTargets {
				mWeightedTarget := map[string]interface{}{
					"virtual_node": aws.StringValue(weightedTarget.VirtualNode),
					"weight":       int(aws.Int64Value(weightedTarget.Weight)),
				}

				vWeightedTargets = append(vWeightedTargets, mWeightedTarget)
			}

			mTcpRoute["action"] = []interface{}{
				map[string]interface{}{
					"weighted_target": vWeightedTargets,
				},
			}
		}
	}

	mTcpRoute["timeout"] = flattenAppmeshTcpTimeout(tcpRoute.Timeout)

	return []interface{}{mTcpRoute}
}

func flattenAppmeshTcpTimeout(tcpTimeout *appmesh.TcpTimeout) []interface{} {
	if tcpTimeout == nil {
		return []interface{}{}
	}

	mTcpTimeout := map[string]interface{}{
		"idle": flattenAppmeshDuration(tcpTimeout.Idle),
	}

	return []interface{}{mTcpTimeout}
}

func flattenAppmeshDuration(duration *appmesh.Duration) []interface{} {
	if duration == nil {
		return []interface{}{}
	}

	mDuration := map[string]interface{}{
		"unit":  aws.StringValue(duration.Unit),
		"value": int(aws.Int64Value(duration.Value)),
	}

	return []interface{}{mDuration}
}

func expandRoute53ResolverEndpointIpAddresses(vIpAddresses *schema.Set) []*route53resolver.IpAddressRequest {
	ipAddressRequests := []*route53resolver.IpAddressRequest{}

	for _, vIpAddress := range vIpAddresses.List() {
		ipAddressRequest := &route53resolver.IpAddressRequest{}

		mIpAddress := vIpAddress.(map[string]interface{})

		if vSubnetId, ok := mIpAddress["subnet_id"].(string); ok && vSubnetId != "" {
			ipAddressRequest.SubnetId = aws.String(vSubnetId)
		}
		if vIp, ok := mIpAddress["ip"].(string); ok && vIp != "" {
			ipAddressRequest.Ip = aws.String(vIp)
		}

		ipAddressRequests = append(ipAddressRequests, ipAddressRequest)
	}

	return ipAddressRequests
}

func flattenRoute53ResolverEndpointIpAddresses(ipAddresses []*route53resolver.IpAddressResponse) []interface{} {
	if ipAddresses == nil {
		return []interface{}{}
	}

	vIpAddresses := []interface{}{}

	for _, ipAddress := range ipAddresses {
		mIpAddress := map[string]interface{}{
			"subnet_id": aws.StringValue(ipAddress.SubnetId),
			"ip":        aws.StringValue(ipAddress.Ip),
			"ip_id":     aws.StringValue(ipAddress.IpId),
		}

		vIpAddresses = append(vIpAddresses, mIpAddress)
	}

	return vIpAddresses
}

func expandRoute53ResolverEndpointIpAddressUpdate(vIpAddress interface{}) *route53resolver.IpAddressUpdate {
	ipAddressUpdate := &route53resolver.IpAddressUpdate{}

	mIpAddress := vIpAddress.(map[string]interface{})

	if vSubnetId, ok := mIpAddress["subnet_id"].(string); ok && vSubnetId != "" {
		ipAddressUpdate.SubnetId = aws.String(vSubnetId)
	}
	if vIp, ok := mIpAddress["ip"].(string); ok && vIp != "" {
		ipAddressUpdate.Ip = aws.String(vIp)
	}
	if vIpId, ok := mIpAddress["ip_id"].(string); ok && vIpId != "" {
		ipAddressUpdate.IpId = aws.String(vIpId)
	}

	return ipAddressUpdate
}

func expandRoute53ResolverRuleTargetIps(vTargetIps *schema.Set) []*route53resolver.TargetAddress {
	targetAddresses := []*route53resolver.TargetAddress{}

	for _, vTargetIp := range vTargetIps.List() {
		targetAddress := &route53resolver.TargetAddress{}

		mTargetIp := vTargetIp.(map[string]interface{})

		if vIp, ok := mTargetIp["ip"].(string); ok && vIp != "" {
			targetAddress.Ip = aws.String(vIp)
		}
		if vPort, ok := mTargetIp["port"].(int); ok {
			targetAddress.Port = aws.Int64(int64(vPort))
		}

		targetAddresses = append(targetAddresses, targetAddress)
	}

	return targetAddresses
}

func flattenRoute53ResolverRuleTargetIps(targetAddresses []*route53resolver.TargetAddress) []interface{} {
	if targetAddresses == nil {
		return []interface{}{}
	}

	vTargetIps := []interface{}{}

	for _, targetAddress := range targetAddresses {
		mTargetIp := map[string]interface{}{
			"ip":   aws.StringValue(targetAddress.Ip),
			"port": int(aws.Int64Value(targetAddress.Port)),
		}

		vTargetIps = append(vTargetIps, mTargetIp)
	}

	return vTargetIps
}
