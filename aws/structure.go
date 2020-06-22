package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/private/protocol/json/jsonutil"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
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
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/worklink"
	"github.com/beevik/etree"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/mitchellh/copystructure"
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

// Takes the result of flatmap. Expand for an array of listeners and
// returns ECS Volume compatible objects
func expandEcsVolumes(configured []interface{}) []*ecs.Volume {
	volumes := make([]*ecs.Volume, 0, len(configured))

	// Loop over our configured volumes and create
	// an array of aws-sdk-go compatible objects
	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})

		l := &ecs.Volume{
			Name: aws.String(data["name"].(string)),
		}

		hostPath := data["host_path"].(string)
		if hostPath != "" {
			l.Host = &ecs.HostVolumeProperties{
				SourcePath: aws.String(hostPath),
			}
		}

		configList, ok := data["docker_volume_configuration"].([]interface{})
		if ok && len(configList) > 0 {
			config := configList[0].(map[string]interface{})
			l.DockerVolumeConfiguration = &ecs.DockerVolumeConfiguration{}

			if v, ok := config["scope"].(string); ok && v != "" {
				l.DockerVolumeConfiguration.Scope = aws.String(v)
			}

			if v, ok := config["autoprovision"]; ok && v != "" {
				scope := l.DockerVolumeConfiguration.Scope
				if scope == nil || *scope != ecs.ScopeTask || v.(bool) {
					l.DockerVolumeConfiguration.Autoprovision = aws.Bool(v.(bool))
				}
			}

			if v, ok := config["driver"].(string); ok && v != "" {
				l.DockerVolumeConfiguration.Driver = aws.String(v)
			}

			if v, ok := config["driver_opts"].(map[string]interface{}); ok && len(v) > 0 {
				l.DockerVolumeConfiguration.DriverOpts = stringMapToPointers(v)
			}

			if v, ok := config["labels"].(map[string]interface{}); ok && len(v) > 0 {
				l.DockerVolumeConfiguration.Labels = stringMapToPointers(v)
			}
		}

		efsConfig, ok := data["efs_volume_configuration"].([]interface{})
		if ok && len(efsConfig) > 0 {
			config := efsConfig[0].(map[string]interface{})
			l.EfsVolumeConfiguration = &ecs.EFSVolumeConfiguration{}

			if v, ok := config["file_system_id"].(string); ok && v != "" {
				l.EfsVolumeConfiguration.FileSystemId = aws.String(v)
			}

			if v, ok := config["root_directory"].(string); ok && v != "" {
				l.EfsVolumeConfiguration.RootDirectory = aws.String(v)
			}
		}

		volumes = append(volumes, l)
	}

	return volumes
}

// Takes JSON in a string. Decodes JSON into
// an array of ecs.ContainerDefinition compatible objects
func expandEcsContainerDefinitions(rawDefinitions string) ([]*ecs.ContainerDefinition, error) {
	var definitions []*ecs.ContainerDefinition

	err := json.Unmarshal([]byte(rawDefinitions), &definitions)
	if err != nil {
		return nil, fmt.Errorf("Error decoding JSON: %s", err)
	}

	return definitions, nil
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
			memberships := expandStringList(raw.(*schema.Set).List())
			if len(memberships) > 0 {
				o.DBSecurityGroupMemberships = memberships
			}
		}

		if raw, ok := data["vpc_security_group_memberships"]; ok {
			memberships := expandStringList(raw.(*schema.Set).List())
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
func expandElastiCacheParameters(configured []interface{}) []*elasticache.ParameterNameValue {
	parameters := make([]*elasticache.ParameterNameValue, 0, len(configured))

	// Loop over our configured parameters and create
	// an array of aws-sdk-go compatible objects
	for _, pRaw := range configured {
		data := pRaw.(map[string]interface{})

		p := &elasticache.ParameterNameValue{
			ParameterName:  aws.String(data["name"].(string)),
			ParameterValue: aws.String(data["value"].(string)),
		}

		parameters = append(parameters, p)
	}

	return parameters
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

// Flattens an array of Volumes into a []map[string]interface{}
func flattenEcsVolumes(list []*ecs.Volume) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, volume := range list {
		l := map[string]interface{}{
			"name": *volume.Name,
		}

		if volume.Host != nil && volume.Host.SourcePath != nil {
			l["host_path"] = *volume.Host.SourcePath
		}

		if volume.DockerVolumeConfiguration != nil {
			l["docker_volume_configuration"] = flattenDockerVolumeConfiguration(volume.DockerVolumeConfiguration)
		}

		if volume.EfsVolumeConfiguration != nil {
			l["efs_volume_configuration"] = flattenEFSVolumeConfiguration(volume.EfsVolumeConfiguration)
		}

		result = append(result, l)
	}
	return result
}

func flattenDockerVolumeConfiguration(config *ecs.DockerVolumeConfiguration) []interface{} {
	var items []interface{}
	m := make(map[string]interface{})

	if v := config.Scope; v != nil {
		m["scope"] = aws.StringValue(v)
	}

	if v := config.Autoprovision; v != nil {
		m["autoprovision"] = aws.BoolValue(v)
	}

	if v := config.Driver; v != nil {
		m["driver"] = aws.StringValue(v)
	}

	if config.DriverOpts != nil {
		m["driver_opts"] = pointersMapToStringList(config.DriverOpts)
	}

	if v := config.Labels; v != nil {
		m["labels"] = pointersMapToStringList(v)
	}

	items = append(items, m)
	return items
}

func flattenEFSVolumeConfiguration(config *ecs.EFSVolumeConfiguration) []interface{} {
	var items []interface{}
	m := make(map[string]interface{})
	if config != nil {
		if v := config.FileSystemId; v != nil {
			m["file_system_id"] = aws.StringValue(v)
		}

		if v := config.RootDirectory; v != nil {
			m["root_directory"] = aws.StringValue(v)
		}
	}

	items = append(items, m)
	return items
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

// Encodes an array of ecs.ContainerDefinitions into a JSON string
func flattenEcsContainerDefinitions(definitions []*ecs.ContainerDefinition) (string, error) {
	b, err := jsonutil.BuildJSON(definitions)
	if err != nil {
		return "", err
	}

	return string(b), nil
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
func flattenElastiCacheParameters(list []*elasticache.Parameter) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, i := range list {
		if i.ParameterValue != nil {
			result = append(result, map[string]interface{}{
				"name":  strings.ToLower(*i.ParameterName),
				"value": *i.ParameterValue,
			})
		}
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
func flattenDocDBParameters(list []*docdb.Parameter) []map[string]interface{} {
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

// Takes the result of schema.Set of strings and returns a []*string
func expandStringSet(configured *schema.Set) []*string {
	return expandStringList(configured.List())
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
	return schema.NewSet(schema.HashString, flattenStringList(list))
}

// hashStringCaseInsensitive hashes strings in a case insensitive manner.
// If you want a Set of strings and are case inensitive, this is the SchemaSetFunc you want.
func hashStringCaseInsensitive(v interface{}) int {
	return hashcode.String(strings.ToLower(v.(string)))
}

func flattenCaseInsensitiveStringSet(list []*string) *schema.Set {
	return schema.NewSet(hashStringCaseInsensitive, flattenStringList(list))
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

//Flattens network interface attachment into a map[string]interface
func flattenAttachment(a *ec2.NetworkInterfaceAttachment) map[string]interface{} {
	att := make(map[string]interface{})
	if a.InstanceId != nil {
		att["instance"] = *a.InstanceId
	}
	att["device_index"] = *a.DeviceIndex
	att["attachment_id"] = *a.AttachmentId
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
	att := make(map[string]interface{})
	if a.AllocationId != nil {
		att["allocation_id"] = *a.AllocationId
	}
	if a.AssociationId != nil {
		att["association_id"] = *a.AssociationId
	}
	if a.IpOwnerId != nil {
		att["ip_owner_id"] = *a.IpOwnerId
	}
	if a.PublicDnsName != nil {
		att["public_dns_name"] = *a.PublicDnsName
	}
	if a.PublicIp != nil {
		att["public_ip"] = *a.PublicIp
	}
	return []interface{}{att}
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

	return domainEndpointOptions
}

func flattenESDomainEndpointOptions(domainEndpointOptions *elasticsearch.DomainEndpointOptions) []interface{} {
	if domainEndpointOptions == nil {
		return nil
	}

	m := map[string]interface{}{
		"enforce_https":       aws.BoolValue(domainEndpointOptions.EnforceHTTPS),
		"tls_security_policy": aws.StringValue(domainEndpointOptions.TLSSecurityPolicy),
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
		m["availability_zones"] = schema.NewSet(schema.HashString, flattenStringList(o.AvailabilityZones))
	}
	if o.SecurityGroupIds != nil {
		m["security_group_ids"] = schema.NewSet(schema.HashString, flattenStringList(o.SecurityGroupIds))
	}
	if o.SubnetIds != nil {
		m["subnet_ids"] = schema.NewSet(schema.HashString, flattenStringList(o.SubnetIds))
	}
	if o.VPCId != nil {
		m["vpc_id"] = *o.VPCId
	}

	return []map[string]interface{}{m}
}

func expandESVPCOptions(m map[string]interface{}) *elasticsearch.VPCOptions {
	options := elasticsearch.VPCOptions{}

	if v, ok := m["security_group_ids"]; ok {
		options.SecurityGroupIds = expandStringList(v.(*schema.Set).List())
	}
	if v, ok := m["subnet_ids"]; ok {
		options.SubnetIds = expandStringList(v.(*schema.Set).List())
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
		recordingGroup.ResourceTypes = expandStringList(v.(*schema.Set).List())
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
		m["resource_types"] = schema.NewSet(schema.HashString, flattenStringList(g.ResourceTypes))
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

func stringMapToPointers(m map[string]interface{}) map[string]*string {
	list := make(map[string]*string, len(m))
	for i, v := range m {
		list[i] = aws.String(v.(string))
	}
	return list
}

// diffStringMaps returns the set of keys and values that must be created,
// and the set of keys and values that must be destroyed.
// Equivalent to 'diffTagsGeneric'.
func diffStringMaps(oldMap, newMap map[string]interface{}) (map[string]*string, map[string]*string) {
	// First, we're creating everything we have
	create := map[string]*string{}
	for k, v := range newMap {
		create[k] = aws.String(v.(string))
	}

	// Build the map of what to remove
	remove := map[string]*string{}
	for k, v := range oldMap {
		old, ok := create[k]
		if !ok || aws.StringValue(old) != v.(string) {
			// Delete it!
			remove[k] = aws.String(v.(string))
		} else if ok {
			// already present so remove from new
			delete(create, k)
		}
	}

	return create, remove
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

func flattenLambdaEnvironment(lambdaEnv *lambda.EnvironmentResponse) []interface{} {
	envs := make(map[string]interface{})
	en := make(map[string]string)

	if lambdaEnv == nil {
		return nil
	}

	for k, v := range lambdaEnv.Variables {
		en[k] = *v
	}
	if len(en) > 0 {
		envs["variables"] = en
	}

	return []interface{}{envs}
}

func expandLambdaEventSourceMappingDestinationConfig(vDest []interface{}) *lambda.DestinationConfig {
	if len(vDest) == 0 {
		return nil
	}

	dest := &lambda.DestinationConfig{}
	onFailure := &lambda.OnFailure{}

	if len(vDest) > 0 {
		if config, ok := vDest[0].(map[string]interface{}); ok {
			if vOnFailure, ok := config["on_failure"].([]interface{}); ok && len(vOnFailure) > 0 && vOnFailure[0] != nil {
				mOnFailure := vOnFailure[0].(map[string]interface{})
				onFailure.SetDestination(mOnFailure["destination_arn"].(string))
			}
		}
	}
	dest.SetOnFailure(onFailure)
	return dest
}

func flattenLambdaEventSourceMappingDestinationConfig(dest *lambda.DestinationConfig) []interface{} {
	mDest := map[string]interface{}{}
	mOnFailure := map[string]interface{}{}
	if dest != nil {
		if dest.OnFailure != nil {
			if dest.OnFailure.Destination != nil {
				mOnFailure["destination_arn"] = *dest.OnFailure.Destination
				mDest["on_failure"] = []interface{}{mOnFailure}
			}
		}
	}

	if len(mDest) == 0 {
		return nil
	}

	return []interface{}{mDest}
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

	settings["subnet_ids"] = schema.NewSet(schema.HashString, flattenStringList(s.SubnetIds))
	settings["security_group_ids"] = schema.NewSet(schema.HashString, flattenStringList(s.SecurityGroupIds))
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

func deprecatedExpandApiGatewayMethodParametersJSONOperations(d *schema.ResourceData, key string, prefix string) ([]*apigateway.PatchOperation, error) {
	operations := make([]*apigateway.PatchOperation, 0)
	oldParameters, newParameters := d.GetChange(key)
	oldParametersMap := make(map[string]interface{})
	newParametersMap := make(map[string]interface{})

	if err := json.Unmarshal([]byte(oldParameters.(string)), &oldParametersMap); err != nil {
		err := fmt.Errorf("Error unmarshaling old %s: %s", key, err)
		return operations, err
	}

	if err := json.Unmarshal([]byte(newParameters.(string)), &newParametersMap); err != nil {
		err := fmt.Errorf("Error unmarshaling new %s: %s", key, err)
		return operations, err
	}

	for k := range oldParametersMap {
		operation := apigateway.PatchOperation{
			Op:   aws.String("remove"),
			Path: aws.String(fmt.Sprintf("/%s/%s", prefix, k)),
		}

		for nK, nV := range newParametersMap {
			if nK == k {
				operation.Op = aws.String("replace")
				operation.Value = aws.String(strconv.FormatBool(nV.(bool)))
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
			operation := apigateway.PatchOperation{
				Op:    aws.String("add"),
				Path:  aws.String(fmt.Sprintf("/%s/%s", prefix, nK)),
				Value: aws.String(strconv.FormatBool(nV.(bool))),
			}
			operations = append(operations, &operation)
		}
	}

	return operations, nil
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

func expandCloudWatchLogMetricTransformations(m map[string]interface{}) []*cloudwatchlogs.MetricTransformation {
	transformation := cloudwatchlogs.MetricTransformation{
		MetricName:      aws.String(m["name"].(string)),
		MetricNamespace: aws.String(m["namespace"].(string)),
		MetricValue:     aws.String(m["value"].(string)),
	}

	if m["default_value"].(string) != "" {
		value, _ := strconv.ParseFloat(m["default_value"].(string), 64)
		transformation.DefaultValue = aws.Float64(value)
	}

	return []*cloudwatchlogs.MetricTransformation{&transformation}
}

func flattenCloudWatchLogMetricTransformations(ts []*cloudwatchlogs.MetricTransformation) []interface{} {
	mts := make([]interface{}, 0)
	m := make(map[string]interface{})

	m["name"] = aws.StringValue(ts[0].MetricName)
	m["namespace"] = aws.StringValue(ts[0].MetricNamespace)
	m["value"] = aws.StringValue(ts[0].MetricValue)

	if ts[0].DefaultValue == nil {
		m["default_value"] = ""
	} else {
		m["default_value"] = strconv.FormatFloat(aws.Float64Value(ts[0].DefaultValue), 'f', -1, 64)
	}

	mts = append(mts, m)

	return mts
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
		m["compliance_resource_types"] = schema.NewSet(schema.HashString, flattenStringList(scope.ComplianceResourceTypes))
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
			scope.ComplianceResourceTypes = expandStringList(l.List())
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

func normalizeCloudFormationTemplate(templateString interface{}) (string, error) {
	if looksLikeJsonString(templateString) {
		return structure.NormalizeJsonString(templateString.(string))
	}

	return checkYamlString(templateString)
}

func flattenApiGatewayUsageApiStages(s []*apigateway.ApiStage) []map[string]interface{} {
	stages := make([]map[string]interface{}, 0)

	for _, bd := range s {
		if bd.ApiId != nil && bd.Stage != nil {
			stage := make(map[string]interface{})
			stage["api_id"] = *bd.ApiId
			stage["stage"] = *bd.Stage

			stages = append(stages, stage)
		}
	}

	if len(stages) > 0 {
		return stages
	}

	return nil
}

func flattenApiGatewayUsagePlanThrottling(s *apigateway.ThrottleSettings) []map[string]interface{} {
	settings := make(map[string]interface{})

	if s == nil {
		return nil
	}

	if s.BurstLimit != nil {
		settings["burst_limit"] = *s.BurstLimit
	}

	if s.RateLimit != nil {
		settings["rate_limit"] = *s.RateLimit
	}

	return []map[string]interface{}{settings}
}

func flattenApiGatewayUsagePlanQuota(s *apigateway.QuotaSettings) []map[string]interface{} {
	settings := make(map[string]interface{})

	if s == nil {
		return nil
	}

	if s.Limit != nil {
		settings["limit"] = *s.Limit
	}

	if s.Offset != nil {
		settings["offset"] = *s.Offset
	}

	if s.Period != nil {
		settings["period"] = *s.Period
	}

	return []map[string]interface{}{settings}
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

func flattenCognitoUserPoolEmailConfiguration(s *cognitoidentityprovider.EmailConfigurationType) []map[string]interface{} {
	m := make(map[string]interface{})

	if s == nil {
		return nil
	}

	if s.ReplyToEmailAddress != nil {
		m["reply_to_email_address"] = *s.ReplyToEmailAddress
	}

	if s.From != nil {
		m["from_email_address"] = *s.From
	}

	if s.SourceArn != nil {
		m["source_arn"] = *s.SourceArn
	}

	if s.EmailSendingAccount != nil {
		m["email_sending_account"] = *s.EmailSendingAccount
	}

	if len(m) > 0 {
		return []map[string]interface{}{m}
	}

	return []map[string]interface{}{}
}

func expandCognitoUserPoolAdminCreateUserConfig(config map[string]interface{}) *cognitoidentityprovider.AdminCreateUserConfigType {
	configs := &cognitoidentityprovider.AdminCreateUserConfigType{}

	if v, ok := config["allow_admin_create_user_only"]; ok {
		configs.AllowAdminCreateUserOnly = aws.Bool(v.(bool))
	}

	if v, ok := config["invite_message_template"]; ok {
		data := v.([]interface{})

		if len(data) > 0 {
			m, ok := data[0].(map[string]interface{})

			if ok {
				imt := &cognitoidentityprovider.MessageTemplateType{}

				if v, ok := m["email_message"]; ok {
					imt.EmailMessage = aws.String(v.(string))
				}

				if v, ok := m["email_subject"]; ok {
					imt.EmailSubject = aws.String(v.(string))
				}

				if v, ok := m["sms_message"]; ok {
					imt.SMSMessage = aws.String(v.(string))
				}

				configs.InviteMessageTemplate = imt
			}
		}
	}

	configs.UnusedAccountValidityDays = aws.Int64(int64(config["unused_account_validity_days"].(int)))

	return configs
}

func flattenCognitoUserPoolAdminCreateUserConfig(s *cognitoidentityprovider.AdminCreateUserConfigType) []map[string]interface{} {
	config := map[string]interface{}{}

	if s == nil {
		return nil
	}

	if s.AllowAdminCreateUserOnly != nil {
		config["allow_admin_create_user_only"] = *s.AllowAdminCreateUserOnly
	}

	if s.InviteMessageTemplate != nil {
		subconfig := map[string]interface{}{}

		if s.InviteMessageTemplate.EmailMessage != nil {
			subconfig["email_message"] = *s.InviteMessageTemplate.EmailMessage
		}

		if s.InviteMessageTemplate.EmailSubject != nil {
			subconfig["email_subject"] = *s.InviteMessageTemplate.EmailSubject
		}

		if s.InviteMessageTemplate.SMSMessage != nil {
			subconfig["sms_message"] = *s.InviteMessageTemplate.SMSMessage
		}

		if len(subconfig) > 0 {
			config["invite_message_template"] = []map[string]interface{}{subconfig}
		}
	}

	config["unused_account_validity_days"] = *s.UnusedAccountValidityDays

	return []map[string]interface{}{config}
}

func expandCognitoUserPoolDeviceConfiguration(config map[string]interface{}) *cognitoidentityprovider.DeviceConfigurationType {
	configs := &cognitoidentityprovider.DeviceConfigurationType{}

	if v, ok := config["challenge_required_on_new_device"]; ok {
		configs.ChallengeRequiredOnNewDevice = aws.Bool(v.(bool))
	}

	if v, ok := config["device_only_remembered_on_user_prompt"]; ok {
		configs.DeviceOnlyRememberedOnUserPrompt = aws.Bool(v.(bool))
	}

	return configs
}

func flattenCognitoUserPoolDeviceConfiguration(s *cognitoidentityprovider.DeviceConfigurationType) []map[string]interface{} {
	config := map[string]interface{}{}

	if s == nil {
		return nil
	}

	if s.ChallengeRequiredOnNewDevice != nil {
		config["challenge_required_on_new_device"] = *s.ChallengeRequiredOnNewDevice
	}

	if s.DeviceOnlyRememberedOnUserPrompt != nil {
		config["device_only_remembered_on_user_prompt"] = *s.DeviceOnlyRememberedOnUserPrompt
	}

	return []map[string]interface{}{config}
}

func expandCognitoUserPoolLambdaConfig(config map[string]interface{}) *cognitoidentityprovider.LambdaConfigType {
	configs := &cognitoidentityprovider.LambdaConfigType{}

	if v, ok := config["create_auth_challenge"]; ok && v.(string) != "" {
		configs.CreateAuthChallenge = aws.String(v.(string))
	}

	if v, ok := config["custom_message"]; ok && v.(string) != "" {
		configs.CustomMessage = aws.String(v.(string))
	}

	if v, ok := config["define_auth_challenge"]; ok && v.(string) != "" {
		configs.DefineAuthChallenge = aws.String(v.(string))
	}

	if v, ok := config["post_authentication"]; ok && v.(string) != "" {
		configs.PostAuthentication = aws.String(v.(string))
	}

	if v, ok := config["post_confirmation"]; ok && v.(string) != "" {
		configs.PostConfirmation = aws.String(v.(string))
	}

	if v, ok := config["pre_authentication"]; ok && v.(string) != "" {
		configs.PreAuthentication = aws.String(v.(string))
	}

	if v, ok := config["pre_sign_up"]; ok && v.(string) != "" {
		configs.PreSignUp = aws.String(v.(string))
	}

	if v, ok := config["pre_token_generation"]; ok && v.(string) != "" {
		configs.PreTokenGeneration = aws.String(v.(string))
	}

	if v, ok := config["user_migration"]; ok && v.(string) != "" {
		configs.UserMigration = aws.String(v.(string))
	}

	if v, ok := config["verify_auth_challenge_response"]; ok && v.(string) != "" {
		configs.VerifyAuthChallengeResponse = aws.String(v.(string))
	}

	return configs
}

func flattenCognitoUserPoolLambdaConfig(s *cognitoidentityprovider.LambdaConfigType) []map[string]interface{} {
	m := map[string]interface{}{}

	if s == nil {
		return nil
	}

	if s.CreateAuthChallenge != nil {
		m["create_auth_challenge"] = *s.CreateAuthChallenge
	}

	if s.CustomMessage != nil {
		m["custom_message"] = *s.CustomMessage
	}

	if s.DefineAuthChallenge != nil {
		m["define_auth_challenge"] = *s.DefineAuthChallenge
	}

	if s.PostAuthentication != nil {
		m["post_authentication"] = *s.PostAuthentication
	}

	if s.PostConfirmation != nil {
		m["post_confirmation"] = *s.PostConfirmation
	}

	if s.PreAuthentication != nil {
		m["pre_authentication"] = *s.PreAuthentication
	}

	if s.PreSignUp != nil {
		m["pre_sign_up"] = *s.PreSignUp
	}

	if s.PreTokenGeneration != nil {
		m["pre_token_generation"] = *s.PreTokenGeneration
	}

	if s.UserMigration != nil {
		m["user_migration"] = *s.UserMigration
	}

	if s.VerifyAuthChallengeResponse != nil {
		m["verify_auth_challenge_response"] = *s.VerifyAuthChallengeResponse
	}

	if len(m) > 0 {
		return []map[string]interface{}{m}
	}

	return []map[string]interface{}{}
}

func expandCognitoUserPoolPasswordPolicy(config map[string]interface{}) *cognitoidentityprovider.PasswordPolicyType {
	configs := &cognitoidentityprovider.PasswordPolicyType{}

	if v, ok := config["minimum_length"]; ok {
		configs.MinimumLength = aws.Int64(int64(v.(int)))
	}

	if v, ok := config["require_lowercase"]; ok {
		configs.RequireLowercase = aws.Bool(v.(bool))
	}

	if v, ok := config["require_numbers"]; ok {
		configs.RequireNumbers = aws.Bool(v.(bool))
	}

	if v, ok := config["require_symbols"]; ok {
		configs.RequireSymbols = aws.Bool(v.(bool))
	}

	if v, ok := config["require_uppercase"]; ok {
		configs.RequireUppercase = aws.Bool(v.(bool))
	}

	if v, ok := config["temporary_password_validity_days"]; ok {
		configs.TemporaryPasswordValidityDays = aws.Int64(int64(v.(int)))
	}

	return configs
}

func flattenCognitoUserPoolUserPoolAddOns(s *cognitoidentityprovider.UserPoolAddOnsType) []map[string]interface{} {
	config := make(map[string]interface{})

	if s == nil {
		return []map[string]interface{}{}
	}

	if s.AdvancedSecurityMode != nil {
		config["advanced_security_mode"] = *s.AdvancedSecurityMode
	}

	return []map[string]interface{}{config}
}

func flattenCognitoUserPoolPasswordPolicy(s *cognitoidentityprovider.PasswordPolicyType) []map[string]interface{} {
	m := map[string]interface{}{}

	if s == nil {
		return nil
	}

	if s.MinimumLength != nil {
		m["minimum_length"] = *s.MinimumLength
	}

	if s.RequireLowercase != nil {
		m["require_lowercase"] = *s.RequireLowercase
	}

	if s.RequireNumbers != nil {
		m["require_numbers"] = *s.RequireNumbers
	}

	if s.RequireSymbols != nil {
		m["require_symbols"] = *s.RequireSymbols
	}

	if s.RequireUppercase != nil {
		m["require_uppercase"] = *s.RequireUppercase
	}

	if s.TemporaryPasswordValidityDays != nil {
		m["temporary_password_validity_days"] = *s.TemporaryPasswordValidityDays
	}

	if len(m) > 0 {
		return []map[string]interface{}{m}
	}

	return []map[string]interface{}{}
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

func expandCognitoUserPoolSchema(inputs []interface{}) []*cognitoidentityprovider.SchemaAttributeType {
	configs := make([]*cognitoidentityprovider.SchemaAttributeType, len(inputs))

	for i, input := range inputs {
		param := input.(map[string]interface{})
		config := &cognitoidentityprovider.SchemaAttributeType{}

		if v, ok := param["attribute_data_type"]; ok {
			config.AttributeDataType = aws.String(v.(string))
		}

		if v, ok := param["developer_only_attribute"]; ok {
			config.DeveloperOnlyAttribute = aws.Bool(v.(bool))
		}

		if v, ok := param["mutable"]; ok {
			config.Mutable = aws.Bool(v.(bool))
		}

		if v, ok := param["name"]; ok {
			config.Name = aws.String(v.(string))
		}

		if v, ok := param["required"]; ok {
			config.Required = aws.Bool(v.(bool))
		}

		if v, ok := param["number_attribute_constraints"]; ok {
			data := v.([]interface{})

			if len(data) > 0 {
				m, ok := data[0].(map[string]interface{})
				if ok {
					numberAttributeConstraintsType := &cognitoidentityprovider.NumberAttributeConstraintsType{}

					if v, ok := m["min_value"]; ok && v.(string) != "" {
						numberAttributeConstraintsType.MinValue = aws.String(v.(string))
					}

					if v, ok := m["max_value"]; ok && v.(string) != "" {
						numberAttributeConstraintsType.MaxValue = aws.String(v.(string))
					}

					config.NumberAttributeConstraints = numberAttributeConstraintsType
				}
			}
		}

		if v, ok := param["string_attribute_constraints"]; ok {
			data := v.([]interface{})

			if len(data) > 0 {
				m, _ := data[0].(map[string]interface{})
				if ok {
					stringAttributeConstraintsType := &cognitoidentityprovider.StringAttributeConstraintsType{}

					if l, ok := m["min_length"]; ok && l.(string) != "" {
						stringAttributeConstraintsType.MinLength = aws.String(l.(string))
					}

					if l, ok := m["max_length"]; ok && l.(string) != "" {
						stringAttributeConstraintsType.MaxLength = aws.String(l.(string))
					}

					config.StringAttributeConstraints = stringAttributeConstraintsType
				}
			}
		}

		configs[i] = config
	}

	return configs
}

func cognitoUserPoolSchemaAttributeMatchesStandardAttribute(input *cognitoidentityprovider.SchemaAttributeType) bool {
	if input == nil {
		return false
	}

	// All standard attributes always returned by API
	// https://docs.aws.amazon.com/cognito/latest/developerguide/user-pool-settings-attributes.html#cognito-user-pools-standard-attributes
	var standardAttributes = []cognitoidentityprovider.SchemaAttributeType{
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("address"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("birthdate"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("10"),
				MinLength: aws.String("10"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("email"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeBoolean),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("email_verified"),
			Required:               aws.Bool(false),
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("gender"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("given_name"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("family_name"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("locale"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("middle_name"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("name"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("nickname"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("phone_number"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeBoolean),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("phone_number_verified"),
			Required:               aws.Bool(false),
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("picture"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("preferred_username"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("profile"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(false),
			Name:                   aws.String("sub"),
			Required:               aws.Bool(true),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("1"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeNumber),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("updated_at"),
			NumberAttributeConstraints: &cognitoidentityprovider.NumberAttributeConstraintsType{
				MinValue: aws.String("0"),
			},
			Required: aws.Bool(false),
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("website"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
		{
			AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
			DeveloperOnlyAttribute: aws.Bool(false),
			Mutable:                aws.Bool(true),
			Name:                   aws.String("zoneinfo"),
			Required:               aws.Bool(false),
			StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
				MaxLength: aws.String("2048"),
				MinLength: aws.String("0"),
			},
		},
	}
	for _, standardAttribute := range standardAttributes {
		if reflect.DeepEqual(*input, standardAttribute) {
			return true
		}
	}
	return false
}

func flattenCognitoUserPoolSchema(configuredAttributes, inputs []*cognitoidentityprovider.SchemaAttributeType) []map[string]interface{} {
	values := make([]map[string]interface{}, 0)

	for _, input := range inputs {
		if input == nil {
			continue
		}

		// The API returns all standard attributes
		// https://docs.aws.amazon.com/cognito/latest/developerguide/user-pool-settings-attributes.html#cognito-user-pools-standard-attributes
		// Ignore setting them in state if they are unconfigured to prevent a huge and unexpected diff
		configured := false

		for _, configuredAttribute := range configuredAttributes {
			if reflect.DeepEqual(input, configuredAttribute) {
				configured = true
			}
		}

		if !configured {
			if cognitoUserPoolSchemaAttributeMatchesStandardAttribute(input) {
				continue
			}
			// When adding a Cognito Identity Provider, the API will automatically add an "identities" attribute
			identitiesAttribute := cognitoidentityprovider.SchemaAttributeType{
				AttributeDataType:          aws.String(cognitoidentityprovider.AttributeDataTypeString),
				DeveloperOnlyAttribute:     aws.Bool(false),
				Mutable:                    aws.Bool(true),
				Name:                       aws.String("identities"),
				Required:                   aws.Bool(false),
				StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{},
			}
			if reflect.DeepEqual(*input, identitiesAttribute) {
				continue
			}
		}

		var value = map[string]interface{}{
			"attribute_data_type":      aws.StringValue(input.AttributeDataType),
			"developer_only_attribute": aws.BoolValue(input.DeveloperOnlyAttribute),
			"mutable":                  aws.BoolValue(input.Mutable),
			"name":                     strings.TrimPrefix(strings.TrimPrefix(aws.StringValue(input.Name), "dev:"), "custom:"),
			"required":                 aws.BoolValue(input.Required),
		}

		if input.NumberAttributeConstraints != nil {
			subvalue := make(map[string]interface{})

			if input.NumberAttributeConstraints.MinValue != nil {
				subvalue["min_value"] = aws.StringValue(input.NumberAttributeConstraints.MinValue)
			}

			if input.NumberAttributeConstraints.MaxValue != nil {
				subvalue["max_value"] = aws.StringValue(input.NumberAttributeConstraints.MaxValue)
			}

			value["number_attribute_constraints"] = []map[string]interface{}{subvalue}
		}

		if input.StringAttributeConstraints != nil {
			subvalue := make(map[string]interface{})

			if input.StringAttributeConstraints.MinLength != nil {
				subvalue["min_length"] = aws.StringValue(input.StringAttributeConstraints.MinLength)
			}

			if input.StringAttributeConstraints.MaxLength != nil {
				subvalue["max_length"] = aws.StringValue(input.StringAttributeConstraints.MaxLength)
			}

			value["string_attribute_constraints"] = []map[string]interface{}{subvalue}
		}

		values = append(values, value)
	}

	return values
}

func expandCognitoUserPoolUsernameConfiguration(config map[string]interface{}) *cognitoidentityprovider.UsernameConfigurationType {
	usernameConfigurationType := &cognitoidentityprovider.UsernameConfigurationType{
		CaseSensitive: aws.Bool(config["case_sensitive"].(bool)),
	}

	return usernameConfigurationType
}

func flattenCognitoUserPoolUsernameConfiguration(u *cognitoidentityprovider.UsernameConfigurationType) []map[string]interface{} {
	m := map[string]interface{}{}

	if u == nil {
		return nil
	}

	m["case_sensitive"] = *u.CaseSensitive

	return []map[string]interface{}{m}
}

func expandCognitoUserPoolVerificationMessageTemplate(config map[string]interface{}) *cognitoidentityprovider.VerificationMessageTemplateType {
	verificationMessageTemplateType := &cognitoidentityprovider.VerificationMessageTemplateType{}

	if v, ok := config["default_email_option"]; ok && v.(string) != "" {
		verificationMessageTemplateType.DefaultEmailOption = aws.String(v.(string))
	}

	if v, ok := config["email_message"]; ok && v.(string) != "" {
		verificationMessageTemplateType.EmailMessage = aws.String(v.(string))
	}

	if v, ok := config["email_message_by_link"]; ok && v.(string) != "" {
		verificationMessageTemplateType.EmailMessageByLink = aws.String(v.(string))
	}

	if v, ok := config["email_subject"]; ok && v.(string) != "" {
		verificationMessageTemplateType.EmailSubject = aws.String(v.(string))
	}

	if v, ok := config["email_subject_by_link"]; ok && v.(string) != "" {
		verificationMessageTemplateType.EmailSubjectByLink = aws.String(v.(string))
	}

	if v, ok := config["sms_message"]; ok && v.(string) != "" {
		verificationMessageTemplateType.SmsMessage = aws.String(v.(string))
	}

	return verificationMessageTemplateType
}

func flattenCognitoUserPoolVerificationMessageTemplate(s *cognitoidentityprovider.VerificationMessageTemplateType) []map[string]interface{} {
	m := map[string]interface{}{}

	if s == nil {
		return nil
	}

	if s.DefaultEmailOption != nil {
		m["default_email_option"] = *s.DefaultEmailOption
	}

	if s.EmailMessage != nil {
		m["email_message"] = *s.EmailMessage
	}

	if s.EmailMessageByLink != nil {
		m["email_message_by_link"] = *s.EmailMessageByLink
	}

	if s.EmailSubject != nil {
		m["email_subject"] = *s.EmailSubject
	}

	if s.EmailSubjectByLink != nil {
		m["email_subject_by_link"] = *s.EmailSubjectByLink
	}

	if s.SmsMessage != nil {
		m["sms_message"] = *s.SmsMessage
	}

	if len(m) > 0 {
		return []map[string]interface{}{m}
	}

	return []map[string]interface{}{}
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

	config["subnet_ids"] = schema.NewSet(schema.HashString, flattenStringList(c.SubnetIds))
	config["security_group_ids"] = schema.NewSet(schema.HashString, flattenStringList(c.SecurityGroupIds))
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

func expandMqUsers(cfg []interface{}) []*mq.User {
	users := make([]*mq.User, len(cfg))
	for i, m := range cfg {
		u := m.(map[string]interface{})
		user := mq.User{
			Username: aws.String(u["username"].(string)),
			Password: aws.String(u["password"].(string)),
		}
		if v, ok := u["console_access"]; ok {
			user.ConsoleAccess = aws.Bool(v.(bool))
		}
		if v, ok := u["groups"]; ok {
			user.Groups = expandStringList(v.(*schema.Set).List())
		}
		users[i] = &user
	}
	return users
}

// We use cfgdUsers to get & set the password
func flattenMqUsers(users []*mq.User, cfgUsers []interface{}) *schema.Set {
	existingPairs := make(map[string]string)
	for _, u := range cfgUsers {
		user := u.(map[string]interface{})
		username := user["username"].(string)
		existingPairs[username] = user["password"].(string)
	}

	out := make([]interface{}, 0)
	for _, u := range users {
		m := map[string]interface{}{
			"username": *u.Username,
		}
		password := ""
		if p, ok := existingPairs[*u.Username]; ok {
			password = p
		}
		if password != "" {
			m["password"] = password
		}
		if u.ConsoleAccess != nil {
			m["console_access"] = *u.ConsoleAccess
		}
		if len(u.Groups) > 0 {
			m["groups"] = schema.NewSet(schema.HashString, flattenStringList(u.Groups))
		}
		out = append(out, m)
	}
	return schema.NewSet(resourceAwsMqUserHash, out)
}

func expandMqWeeklyStartTime(cfg []interface{}) *mq.WeeklyStartTime {
	if len(cfg) < 1 {
		return nil
	}

	m := cfg[0].(map[string]interface{})
	return &mq.WeeklyStartTime{
		DayOfWeek: aws.String(m["day_of_week"].(string)),
		TimeOfDay: aws.String(m["time_of_day"].(string)),
		TimeZone:  aws.String(m["time_zone"].(string)),
	}
}

func flattenMqWeeklyStartTime(wst *mq.WeeklyStartTime) []interface{} {
	if wst == nil {
		return []interface{}{}
	}
	m := make(map[string]interface{})
	if wst.DayOfWeek != nil {
		m["day_of_week"] = *wst.DayOfWeek
	}
	if wst.TimeOfDay != nil {
		m["time_of_day"] = *wst.TimeOfDay
	}
	if wst.TimeZone != nil {
		m["time_zone"] = *wst.TimeZone
	}
	return []interface{}{m}
}

func expandMqConfigurationId(cfg []interface{}) *mq.ConfigurationId {
	if len(cfg) < 1 {
		return nil
	}

	m := cfg[0].(map[string]interface{})
	out := mq.ConfigurationId{
		Id: aws.String(m["id"].(string)),
	}
	if v, ok := m["revision"].(int); ok && v > 0 {
		out.Revision = aws.Int64(int64(v))
	}

	return &out
}

func flattenMqConfigurationId(cid *mq.ConfigurationId) []interface{} {
	if cid == nil {
		return []interface{}{}
	}
	m := make(map[string]interface{})
	if cid.Id != nil {
		m["id"] = *cid.Id
	}
	if cid.Revision != nil {
		m["revision"] = *cid.Revision
	}
	return []interface{}{m}
}

func flattenMqBrokerInstances(instances []*mq.BrokerInstance) []interface{} {
	if len(instances) == 0 {
		return []interface{}{}
	}
	l := make([]interface{}, len(instances))
	for i, instance := range instances {
		m := make(map[string]interface{})
		if instance.ConsoleURL != nil {
			m["console_url"] = *instance.ConsoleURL
		}
		if len(instance.Endpoints) > 0 {
			m["endpoints"] = aws.StringValueSlice(instance.Endpoints)
		}
		if instance.IpAddress != nil {
			m["ip_address"] = *instance.IpAddress
		}
		l[i] = m
	}

	return l
}

func flattenMqLogs(logs *mq.LogsSummary) []interface{} {
	if logs == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"general": aws.BoolValue(logs.General),
		"audit":   aws.BoolValue(logs.Audit),
	}

	return []interface{}{m}
}

func expandMqLogs(l []interface{}) *mq.Logs {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	logs := &mq.Logs{
		Audit:   aws.Bool(m["audit"].(bool)),
		General: aws.Bool(m["general"].(bool)),
	}

	return logs
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

func diffDynamoDbGSI(oldGsi, newGsi []interface{}, billingMode string) (ops []*dynamodb.GlobalSecondaryIndexUpdate, e error) {
	// Transform slices into maps
	oldGsis := make(map[string]interface{})
	for _, gsidata := range oldGsi {
		m := gsidata.(map[string]interface{})
		oldGsis[m["name"].(string)] = m
	}
	newGsis := make(map[string]interface{})
	for _, gsidata := range newGsi {
		m := gsidata.(map[string]interface{})
		// validate throughput input early, to avoid unnecessary processing
		if e = validateDynamoDbProvisionedThroughput(m, billingMode); e != nil {
			return
		}
		newGsis[m["name"].(string)] = m
	}

	for _, data := range newGsi {
		newMap := data.(map[string]interface{})
		newName := newMap["name"].(string)

		if _, exists := oldGsis[newName]; !exists {
			m := data.(map[string]interface{})
			idxName := m["name"].(string)

			ops = append(ops, &dynamodb.GlobalSecondaryIndexUpdate{
				Create: &dynamodb.CreateGlobalSecondaryIndexAction{
					IndexName:             aws.String(idxName),
					KeySchema:             expandDynamoDbKeySchema(m),
					ProvisionedThroughput: expandDynamoDbProvisionedThroughput(m, billingMode),
					Projection:            expandDynamoDbProjection(m),
				},
			})
		}
	}

	for _, data := range oldGsi {
		oldMap := data.(map[string]interface{})
		oldName := oldMap["name"].(string)

		newData, exists := newGsis[oldName]
		if exists {
			newMap := newData.(map[string]interface{})
			idxName := newMap["name"].(string)

			oldWriteCapacity, oldReadCapacity := oldMap["write_capacity"].(int), oldMap["read_capacity"].(int)
			newWriteCapacity, newReadCapacity := newMap["write_capacity"].(int), newMap["read_capacity"].(int)
			capacityChanged := (oldWriteCapacity != newWriteCapacity || oldReadCapacity != newReadCapacity)

			oldAttributes, err := stripCapacityAttributes(oldMap)
			if err != nil {
				e = err
				return
			}
			newAttributes, err := stripCapacityAttributes(newMap)
			if err != nil {
				e = err
				return
			}
			otherAttributesChanged := !reflect.DeepEqual(oldAttributes, newAttributes)

			if capacityChanged && !otherAttributesChanged {
				update := &dynamodb.GlobalSecondaryIndexUpdate{
					Update: &dynamodb.UpdateGlobalSecondaryIndexAction{
						IndexName:             aws.String(idxName),
						ProvisionedThroughput: expandDynamoDbProvisionedThroughput(newMap, billingMode),
					},
				}
				ops = append(ops, update)
			} else if otherAttributesChanged {
				// Other attributes cannot be updated
				ops = append(ops, &dynamodb.GlobalSecondaryIndexUpdate{
					Delete: &dynamodb.DeleteGlobalSecondaryIndexAction{
						IndexName: aws.String(idxName),
					},
				})

				ops = append(ops, &dynamodb.GlobalSecondaryIndexUpdate{
					Create: &dynamodb.CreateGlobalSecondaryIndexAction{
						IndexName:             aws.String(idxName),
						KeySchema:             expandDynamoDbKeySchema(newMap),
						ProvisionedThroughput: expandDynamoDbProvisionedThroughput(newMap, billingMode),
						Projection:            expandDynamoDbProjection(newMap),
					},
				})
			}
		} else {
			idxName := oldName
			ops = append(ops, &dynamodb.GlobalSecondaryIndexUpdate{
				Delete: &dynamodb.DeleteGlobalSecondaryIndexAction{
					IndexName: aws.String(idxName),
				},
			})
		}
	}
	return
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

func flattenDynamoDbTtl(ttlOutput *dynamodb.DescribeTimeToLiveOutput) []interface{} {
	m := map[string]interface{}{
		"enabled": false,
	}

	if ttlOutput == nil || ttlOutput.TimeToLiveDescription == nil {
		return []interface{}{m}
	}

	ttlDesc := ttlOutput.TimeToLiveDescription

	m["attribute_name"] = aws.StringValue(ttlDesc.AttributeName)
	m["enabled"] = (aws.StringValue(ttlDesc.TimeToLiveStatus) == dynamodb.TimeToLiveStatusEnabled)

	return []interface{}{m}
}

func flattenDynamoDbPitr(pitrDesc *dynamodb.DescribeContinuousBackupsOutput) []interface{} {
	m := map[string]interface{}{
		"enabled": false,
	}

	if pitrDesc == nil {
		return []interface{}{m}
	}

	if pitrDesc.ContinuousBackupsDescription != nil {
		pitr := pitrDesc.ContinuousBackupsDescription.PointInTimeRecoveryDescription
		if pitr != nil {
			m["enabled"] = (*pitr.PointInTimeRecoveryStatus == dynamodb.PointInTimeRecoveryStatusEnabled)
		}
	}

	return []interface{}{m}
}

func flattenAwsDynamoDbReplicaDescriptions(apiObjects []*dynamodb.ReplicaDescription) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"region_name": aws.StringValue(apiObject.RegionName),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenAwsDynamoDbTableResource(d *schema.ResourceData, table *dynamodb.TableDescription) error {
	d.Set("billing_mode", dynamodb.BillingModeProvisioned)
	if table.BillingModeSummary != nil {
		d.Set("billing_mode", table.BillingModeSummary.BillingMode)
	}

	d.Set("write_capacity", table.ProvisionedThroughput.WriteCapacityUnits)
	d.Set("read_capacity", table.ProvisionedThroughput.ReadCapacityUnits)

	attributes := []interface{}{}
	for _, attrdef := range table.AttributeDefinitions {
		attribute := map[string]string{
			"name": *attrdef.AttributeName,
			"type": *attrdef.AttributeType,
		}
		attributes = append(attributes, attribute)
	}

	d.Set("attribute", attributes)
	d.Set("name", table.TableName)

	for _, attribute := range table.KeySchema {
		if *attribute.KeyType == dynamodb.KeyTypeHash {
			d.Set("hash_key", attribute.AttributeName)
		}

		if *attribute.KeyType == dynamodb.KeyTypeRange {
			d.Set("range_key", attribute.AttributeName)
		}
	}

	lsiList := make([]map[string]interface{}, 0, len(table.LocalSecondaryIndexes))
	for _, lsiObject := range table.LocalSecondaryIndexes {
		lsi := map[string]interface{}{
			"name":            *lsiObject.IndexName,
			"projection_type": *lsiObject.Projection.ProjectionType,
		}

		for _, attribute := range lsiObject.KeySchema {

			if *attribute.KeyType == dynamodb.KeyTypeRange {
				lsi["range_key"] = *attribute.AttributeName
			}
		}
		nkaList := make([]string, len(lsiObject.Projection.NonKeyAttributes))
		for _, nka := range lsiObject.Projection.NonKeyAttributes {
			nkaList = append(nkaList, *nka)
		}
		lsi["non_key_attributes"] = nkaList

		lsiList = append(lsiList, lsi)
	}

	err := d.Set("local_secondary_index", lsiList)
	if err != nil {
		return err
	}

	gsiList := make([]map[string]interface{}, 0, len(table.GlobalSecondaryIndexes))
	for _, gsiObject := range table.GlobalSecondaryIndexes {
		gsi := map[string]interface{}{
			"write_capacity": *gsiObject.ProvisionedThroughput.WriteCapacityUnits,
			"read_capacity":  *gsiObject.ProvisionedThroughput.ReadCapacityUnits,
			"name":           *gsiObject.IndexName,
		}

		for _, attribute := range gsiObject.KeySchema {
			if *attribute.KeyType == dynamodb.KeyTypeHash {
				gsi["hash_key"] = *attribute.AttributeName
			}

			if *attribute.KeyType == dynamodb.KeyTypeRange {
				gsi["range_key"] = *attribute.AttributeName
			}
		}

		gsi["projection_type"] = *(gsiObject.Projection.ProjectionType)

		nonKeyAttrs := make([]string, 0, len(gsiObject.Projection.NonKeyAttributes))
		for _, nonKeyAttr := range gsiObject.Projection.NonKeyAttributes {
			nonKeyAttrs = append(nonKeyAttrs, *nonKeyAttr)
		}
		gsi["non_key_attributes"] = nonKeyAttrs

		gsiList = append(gsiList, gsi)
	}

	if table.StreamSpecification != nil {
		d.Set("stream_view_type", table.StreamSpecification.StreamViewType)
		d.Set("stream_enabled", table.StreamSpecification.StreamEnabled)
	} else {
		d.Set("stream_view_type", "")
		d.Set("stream_enabled", false)
	}

	d.Set("stream_arn", table.LatestStreamArn)
	d.Set("stream_label", table.LatestStreamLabel)

	err = d.Set("global_secondary_index", gsiList)
	if err != nil {
		return err
	}

	sseOptions := []map[string]interface{}{}
	if sseDescription := table.SSEDescription; sseDescription != nil {
		sseOptions = []map[string]interface{}{{
			"enabled":     aws.StringValue(sseDescription.Status) == dynamodb.SSEStatusEnabled,
			"kms_key_arn": aws.StringValue(sseDescription.KMSMasterKeyArn),
		}}
	}
	err = d.Set("server_side_encryption", sseOptions)
	if err != nil {
		return err
	}

	err = d.Set("replica", flattenAwsDynamoDbReplicaDescriptions(table.Replicas))
	if err != nil {
		return err
	}

	d.Set("arn", table.TableArn)

	return nil
}

func expandDynamoDbAttributes(cfg []interface{}) []*dynamodb.AttributeDefinition {
	attributes := make([]*dynamodb.AttributeDefinition, len(cfg))
	for i, attribute := range cfg {
		attr := attribute.(map[string]interface{})
		attributes[i] = &dynamodb.AttributeDefinition{
			AttributeName: aws.String(attr["name"].(string)),
			AttributeType: aws.String(attr["type"].(string)),
		}
	}
	return attributes
}

// TODO: Get rid of keySchemaM - the user should just explicitly define
// this in the config, we shouldn't magically be setting it like this.
// Removal will however require config change, hence BC. :/
func expandDynamoDbLocalSecondaryIndexes(cfg []interface{}, keySchemaM map[string]interface{}) []*dynamodb.LocalSecondaryIndex {
	indexes := make([]*dynamodb.LocalSecondaryIndex, len(cfg))
	for i, lsi := range cfg {
		m := lsi.(map[string]interface{})
		idxName := m["name"].(string)

		// TODO: See https://github.com/terraform-providers/terraform-provider-aws/issues/3176
		if _, ok := m["hash_key"]; !ok {
			m["hash_key"] = keySchemaM["hash_key"]
		}

		indexes[i] = &dynamodb.LocalSecondaryIndex{
			IndexName:  aws.String(idxName),
			KeySchema:  expandDynamoDbKeySchema(m),
			Projection: expandDynamoDbProjection(m),
		}
	}
	return indexes
}

func expandDynamoDbGlobalSecondaryIndex(data map[string]interface{}, billingMode string) *dynamodb.GlobalSecondaryIndex {
	return &dynamodb.GlobalSecondaryIndex{
		IndexName:             aws.String(data["name"].(string)),
		KeySchema:             expandDynamoDbKeySchema(data),
		Projection:            expandDynamoDbProjection(data),
		ProvisionedThroughput: expandDynamoDbProvisionedThroughput(data, billingMode),
	}
}

func validateDynamoDbProvisionedThroughput(data map[string]interface{}, billingMode string) error {
	// if billing mode is PAY_PER_REQUEST, don't need to validate the throughput settings
	if billingMode == dynamodb.BillingModePayPerRequest {
		return nil
	}

	writeCapacity, writeCapacitySet := data["write_capacity"].(int)
	readCapacity, readCapacitySet := data["read_capacity"].(int)

	if !writeCapacitySet || !readCapacitySet {
		return fmt.Errorf("Read and Write capacity should be set when billing mode is %s", dynamodb.BillingModeProvisioned)
	}

	if writeCapacity < 1 {
		return fmt.Errorf("Write capacity must be > 0 when billing mode is %s", dynamodb.BillingModeProvisioned)
	}

	if readCapacity < 1 {
		return fmt.Errorf("Read capacity must be > 0 when billing mode is %s", dynamodb.BillingModeProvisioned)
	}

	return nil
}

func expandDynamoDbProvisionedThroughput(data map[string]interface{}, billingMode string) *dynamodb.ProvisionedThroughput {

	if billingMode == dynamodb.BillingModePayPerRequest {
		return nil
	}

	return &dynamodb.ProvisionedThroughput{
		WriteCapacityUnits: aws.Int64(int64(data["write_capacity"].(int))),
		ReadCapacityUnits:  aws.Int64(int64(data["read_capacity"].(int))),
	}
}

func expandDynamoDbProjection(data map[string]interface{}) *dynamodb.Projection {
	projection := &dynamodb.Projection{
		ProjectionType: aws.String(data["projection_type"].(string)),
	}

	if v, ok := data["non_key_attributes"].([]interface{}); ok && len(v) > 0 {
		projection.NonKeyAttributes = expandStringList(v)
	}

	return projection
}

func expandDynamoDbKeySchema(data map[string]interface{}) []*dynamodb.KeySchemaElement {
	keySchema := []*dynamodb.KeySchemaElement{}

	if v, ok := data["hash_key"]; ok && v != nil && v != "" {
		keySchema = append(keySchema, &dynamodb.KeySchemaElement{
			AttributeName: aws.String(v.(string)),
			KeyType:       aws.String(dynamodb.KeyTypeHash),
		})
	}

	if v, ok := data["range_key"]; ok && v != nil && v != "" {
		keySchema = append(keySchema, &dynamodb.KeySchemaElement{
			AttributeName: aws.String(v.(string)),
			KeyType:       aws.String(dynamodb.KeyTypeRange),
		})
	}

	return keySchema
}

func expandDynamoDbEncryptAtRestOptions(vOptions []interface{}) *dynamodb.SSESpecification {
	options := &dynamodb.SSESpecification{}

	enabled := false
	if len(vOptions) > 0 {
		mOptions := vOptions[0].(map[string]interface{})

		enabled = mOptions["enabled"].(bool)
		if enabled {
			if vKmsKeyArn, ok := mOptions["kms_key_arn"].(string); ok && vKmsKeyArn != "" {
				options.KMSMasterKeyId = aws.String(vKmsKeyArn)
				options.SSEType = aws.String(dynamodb.SSETypeKms)
			}
		}
	}
	options.Enabled = aws.Bool(enabled)

	return options
}

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
		SearchableAttributes: expandStringList(config["searchable_attributes"].(*schema.Set).List()),
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

func expandRdsClusterScalingConfiguration(l []interface{}, engineMode string) *rds.ScalingConfiguration {
	if len(l) == 0 || l[0] == nil {
		// Our default value for MinCapacity is different from AWS's.
		// We need to override it here to avoid a non-empty plan with an empty ScalingConfiguration.
		if engineMode == "serverless" {
			return &rds.ScalingConfiguration{
				MinCapacity: aws.Int64(int64(rdsClusterScalingConfiguration_DefaultMinCapacity)),
			}
		} else {
			return nil
		}
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
				mVirtualService := vVirtualService[0].(map[string]interface{})

				backend.VirtualService = &appmesh.VirtualServiceBackend{}

				if vVirtualServiceName, ok := mVirtualService["virtual_service_name"].(string); ok {
					backend.VirtualService.VirtualServiceName = aws.String(vVirtualServiceName)
				}
			}
			backends = append(backends, backend)
		}

		spec.Backends = backends
	}

	if vListeners, ok := mSpec["listener"].([]interface{}); ok && len(vListeners) > 0 && vListeners[0] != nil {
		listeners := []*appmesh.Listener{}

		for _, vListener := range vListeners {
			listener := &appmesh.Listener{}

			mListener := vListener.(map[string]interface{})

			if vHealthCheck, ok := mListener["health_check"].([]interface{}); ok && len(vHealthCheck) > 0 && vHealthCheck[0] != nil {
				mHealthCheck := vHealthCheck[0].(map[string]interface{})

				listener.HealthCheck = &appmesh.HealthCheckPolicy{}

				if vHealthyThreshold, ok := mHealthCheck["healthy_threshold"].(int); ok && vHealthyThreshold > 0 {
					listener.HealthCheck.HealthyThreshold = aws.Int64(int64(vHealthyThreshold))
				}
				if vIntervalMillis, ok := mHealthCheck["interval_millis"].(int); ok && vIntervalMillis > 0 {
					listener.HealthCheck.IntervalMillis = aws.Int64(int64(vIntervalMillis))
				}
				if vPath, ok := mHealthCheck["path"].(string); ok && vPath != "" {
					listener.HealthCheck.Path = aws.String(vPath)
				}
				if vPort, ok := mHealthCheck["port"].(int); ok && vPort > 0 {
					listener.HealthCheck.Port = aws.Int64(int64(vPort))
				}
				if vProtocol, ok := mHealthCheck["protocol"].(string); ok && vProtocol != "" {
					listener.HealthCheck.Protocol = aws.String(vProtocol)
				}
				if vTimeoutMillis, ok := mHealthCheck["timeout_millis"].(int); ok && vTimeoutMillis > 0 {
					listener.HealthCheck.TimeoutMillis = aws.Int64(int64(vTimeoutMillis))
				}
				if vUnhealthyThreshold, ok := mHealthCheck["unhealthy_threshold"].(int); ok && vUnhealthyThreshold > 0 {
					listener.HealthCheck.UnhealthyThreshold = aws.Int64(int64(vUnhealthyThreshold))
				}
			}

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

	if vLogging, ok := mSpec["logging"].([]interface{}); ok && len(vLogging) > 0 && vLogging[0] != nil {
		mLogging := vLogging[0].(map[string]interface{})

		if vAccessLog, ok := mLogging["access_log"].([]interface{}); ok && len(vAccessLog) > 0 && vAccessLog[0] != nil {
			mAccessLog := vAccessLog[0].(map[string]interface{})

			if vFile, ok := mAccessLog["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
				mFile := vFile[0].(map[string]interface{})

				if vPath, ok := mFile["path"].(string); ok && vPath != "" {
					spec.Logging = &appmesh.Logging{
						AccessLog: &appmesh.AccessLog{
							File: &appmesh.FileAccessLog{
								Path: aws.String(vPath),
							},
						},
					}
				}
			}
		}
	}

	if vServiceDiscovery, ok := mSpec["service_discovery"].([]interface{}); ok && len(vServiceDiscovery) > 0 && vServiceDiscovery[0] != nil {
		spec.ServiceDiscovery = &appmesh.ServiceDiscovery{}

		mServiceDiscovery := vServiceDiscovery[0].(map[string]interface{})

		if vAwsCloudMap, ok := mServiceDiscovery["aws_cloud_map"].([]interface{}); ok && len(vAwsCloudMap) > 0 && vAwsCloudMap[0] != nil {
			spec.ServiceDiscovery.AwsCloudMap = &appmesh.AwsCloudMapServiceDiscovery{}

			mAwsCloudMap := vAwsCloudMap[0].(map[string]interface{})

			if vAttributes, ok := mAwsCloudMap["attributes"].(map[string]interface{}); ok && len(vAttributes) > 0 {
				attributes := []*appmesh.AwsCloudMapInstanceAttribute{}

				for k, v := range vAttributes {
					attributes = append(attributes, &appmesh.AwsCloudMapInstanceAttribute{
						Key:   aws.String(k),
						Value: aws.String(v.(string)),
					})
				}

				spec.ServiceDiscovery.AwsCloudMap.Attributes = attributes
			}
			if vNamespaceName, ok := mAwsCloudMap["namespace_name"].(string); ok && vNamespaceName != "" {
				spec.ServiceDiscovery.AwsCloudMap.NamespaceName = aws.String(vNamespaceName)
			}
			if vServiceName, ok := mAwsCloudMap["service_name"].(string); ok && vServiceName != "" {
				spec.ServiceDiscovery.AwsCloudMap.ServiceName = aws.String(vServiceName)
			}
		}

		if vDns, ok := mServiceDiscovery["dns"].([]interface{}); ok && len(vDns) > 0 && vDns[0] != nil {
			mDns := vDns[0].(map[string]interface{})

			if vHostname, ok := mDns["hostname"].(string); ok && vHostname != "" {
				spec.ServiceDiscovery.Dns = &appmesh.DnsServiceDiscovery{
					Hostname: aws.String(vHostname),
				}
			}
		}
	}

	return spec
}

func flattenAppmeshVirtualNodeSpec(spec *appmesh.VirtualNodeSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	mSpec := map[string]interface{}{}

	if spec.Backends != nil {
		vBackends := []interface{}{}

		for _, backend := range spec.Backends {
			mBackend := map[string]interface{}{}

			if backend.VirtualService != nil {
				mVirtualService := map[string]interface{}{
					"virtual_service_name": aws.StringValue(backend.VirtualService.VirtualServiceName),
				}
				mBackend["virtual_service"] = []interface{}{mVirtualService}
			}

			vBackends = append(vBackends, mBackend)
		}

		mSpec["backend"] = schema.NewSet(appmeshVirtualNodeBackendHash, vBackends)
	}

	if spec.Listeners != nil && spec.Listeners[0] != nil {
		// Per schema definition, set at most 1 Listener
		listener := spec.Listeners[0]
		mListener := map[string]interface{}{}

		if listener.HealthCheck != nil {
			mHealthCheck := map[string]interface{}{
				"healthy_threshold":   int(aws.Int64Value(listener.HealthCheck.HealthyThreshold)),
				"interval_millis":     int(aws.Int64Value(listener.HealthCheck.IntervalMillis)),
				"path":                aws.StringValue(listener.HealthCheck.Path),
				"port":                int(aws.Int64Value(listener.HealthCheck.Port)),
				"protocol":            aws.StringValue(listener.HealthCheck.Protocol),
				"timeout_millis":      int(aws.Int64Value(listener.HealthCheck.TimeoutMillis)),
				"unhealthy_threshold": int(aws.Int64Value(listener.HealthCheck.UnhealthyThreshold)),
			}
			mListener["health_check"] = []interface{}{mHealthCheck}
		}

		if listener.PortMapping != nil {
			mPortMapping := map[string]interface{}{
				"port":     int(aws.Int64Value(listener.PortMapping.Port)),
				"protocol": aws.StringValue(listener.PortMapping.Protocol),
			}
			mListener["port_mapping"] = []interface{}{mPortMapping}
		}

		mSpec["listener"] = []interface{}{mListener}
	}

	if spec.Logging != nil && spec.Logging.AccessLog != nil && spec.Logging.AccessLog.File != nil {
		mSpec["logging"] = []interface{}{
			map[string]interface{}{
				"access_log": []interface{}{
					map[string]interface{}{
						"file": []interface{}{
							map[string]interface{}{
								"path": aws.StringValue(spec.Logging.AccessLog.File.Path),
							},
						},
					},
				},
			},
		}
	}

	if spec.ServiceDiscovery != nil {
		mServiceDiscovery := map[string]interface{}{}

		if spec.ServiceDiscovery.AwsCloudMap != nil {
			vAttributes := map[string]interface{}{}

			for _, attribute := range spec.ServiceDiscovery.AwsCloudMap.Attributes {
				vAttributes[aws.StringValue(attribute.Key)] = aws.StringValue(attribute.Value)
			}

			mServiceDiscovery["aws_cloud_map"] = []interface{}{
				map[string]interface{}{
					"attributes":     vAttributes,
					"namespace_name": aws.StringValue(spec.ServiceDiscovery.AwsCloudMap.NamespaceName),
					"service_name":   aws.StringValue(spec.ServiceDiscovery.AwsCloudMap.ServiceName),
				},
			}
		}

		if spec.ServiceDiscovery.Dns != nil {
			mServiceDiscovery["dns"] = []interface{}{
				map[string]interface{}{
					"hostname": aws.StringValue(spec.ServiceDiscovery.Dns.Hostname),
				},
			}
		}

		mSpec["service_discovery"] = []interface{}{mServiceDiscovery}
	}

	return []interface{}{mSpec}
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

	if vPriority, ok := mSpec["priority"].(int); ok && vPriority > 0 {
		spec.Priority = aws.Int64(int64(vPriority))
	}

	if vHttpRoute, ok := mSpec["http_route"].([]interface{}); ok && len(vHttpRoute) > 0 && vHttpRoute[0] != nil {
		mHttpRoute := vHttpRoute[0].(map[string]interface{})

		spec.HttpRoute = &appmesh.HttpRoute{}

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

				spec.HttpRoute.Action = &appmesh.HttpRouteAction{
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

			spec.HttpRoute.Match = httpRouteMatch
		}
	}

	if vTcpRoute, ok := mSpec["tcp_route"].([]interface{}); ok && len(vTcpRoute) > 0 && vTcpRoute[0] != nil {
		mTcpRoute := vTcpRoute[0].(map[string]interface{})

		spec.TcpRoute = &appmesh.TcpRoute{}

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

				spec.TcpRoute.Action = &appmesh.TcpRouteAction{
					WeightedTargets: weightedTargets,
				}
			}
		}
	}

	return spec
}

func flattenAppmeshRouteSpec(spec *appmesh.RouteSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	mSpec := map[string]interface{}{
		"priority": int(aws.Int64Value(spec.Priority)),
	}

	if httpRoute := spec.HttpRoute; httpRoute != nil {
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
						"weighted_target": schema.NewSet(appmeshWeightedTargetHash, vWeightedTargets),
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
					"header": schema.NewSet(appmeshHttpRouteHeaderHash, vHttpRouteHeaders),
					"method": aws.StringValue(httpRouteMatch.Method),
					"prefix": aws.StringValue(httpRouteMatch.Prefix),
					"scheme": aws.StringValue(httpRouteMatch.Scheme),
				},
			}
		}

		mSpec["http_route"] = []interface{}{mHttpRoute}
	}

	if tcpRoute := spec.TcpRoute; tcpRoute != nil {
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
						"weighted_target": schema.NewSet(appmeshWeightedTargetHash, vWeightedTargets),
					},
				}
			}
		}

		mSpec["tcp_route"] = []interface{}{mTcpRoute}
	}

	return []interface{}{mSpec}
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

func isIpv6CidrsEquals(first, second string) bool {
	if first == "" || second == "" {
		return false
	}
	_, firstMask, _ := net.ParseCIDR(first)
	_, secondMask, _ := net.ParseCIDR(second)

	return firstMask.String() == secondMask.String()
}
