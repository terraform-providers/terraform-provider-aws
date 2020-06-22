// This file contains code generation customizations for each AWS Go SDK service.

package keyvaluetags

import (
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go/service/accessanalyzer"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/aws/aws-sdk-go/service/cloud9"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/aws/aws-sdk-go/service/codedeploy"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/aws/aws-sdk-go/service/codestarnotifications"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/aws/aws-sdk-go/service/dataexchange"
	"github.com/aws/aws-sdk-go/service/datapipeline"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/aws/aws-sdk-go/service/dax"
	"github.com/aws/aws-sdk-go/service/devicefarm"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/aws/aws-sdk-go/service/dlm"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/aws/aws-sdk-go/service/glacier"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/aws/aws-sdk-go/service/inspector"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/aws/aws-sdk-go/service/iotanalytics"
	"github.com/aws/aws-sdk-go/service/iotevents"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/aws/aws-sdk-go/service/kinesisanalyticsv2"
	"github.com/aws/aws-sdk-go/service/kinesisvideo"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/licensemanager"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/aws/aws-sdk-go/service/mediaconnect"
	"github.com/aws/aws-sdk-go/service/mediaconvert"
	"github.com/aws/aws-sdk-go/service/medialive"
	"github.com/aws/aws-sdk-go/service/mediapackage"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/aws/aws-sdk-go/service/qldb"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/aws/aws-sdk-go/service/resourcegroups"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/aws/aws-sdk-go/service/swf"
	"github.com/aws/aws-sdk-go/service/synthetics"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/aws/aws-sdk-go/service/worklink"
	"github.com/aws/aws-sdk-go/service/workspaces"
)

// ServiceClientType determines the service client Go type.
// The AWS Go SDK does not provide a constant or reproducible inference methodology
// to get the correct type name of each service, so we resort to reflection for now.
func ServiceClientType(serviceName string) string {
	var funcType reflect.Type

	switch serviceName {
	case "accessanalyzer":
		funcType = reflect.TypeOf(accessanalyzer.New)
	case "acm":
		funcType = reflect.TypeOf(acm.New)
	case "acmpca":
		funcType = reflect.TypeOf(acmpca.New)
	case "amplify":
		funcType = reflect.TypeOf(amplify.New)
	case "apigateway":
		funcType = reflect.TypeOf(apigateway.New)
	case "apigatewayv2":
		funcType = reflect.TypeOf(apigatewayv2.New)
	case "appmesh":
		funcType = reflect.TypeOf(appmesh.New)
	case "appstream":
		funcType = reflect.TypeOf(appstream.New)
	case "appsync":
		funcType = reflect.TypeOf(appsync.New)
	case "athena":
		funcType = reflect.TypeOf(athena.New)
	case "backup":
		funcType = reflect.TypeOf(backup.New)
	case "cloud9":
		funcType = reflect.TypeOf(cloud9.New)
	case "cloudfront":
		funcType = reflect.TypeOf(cloudfront.New)
	case "cloudhsmv2":
		funcType = reflect.TypeOf(cloudhsmv2.New)
	case "cloudtrail":
		funcType = reflect.TypeOf(cloudtrail.New)
	case "cloudwatch":
		funcType = reflect.TypeOf(cloudwatch.New)
	case "cloudwatchevents":
		funcType = reflect.TypeOf(cloudwatchevents.New)
	case "cloudwatchlogs":
		funcType = reflect.TypeOf(cloudwatchlogs.New)
	case "codecommit":
		funcType = reflect.TypeOf(codecommit.New)
	case "codedeploy":
		funcType = reflect.TypeOf(codedeploy.New)
	case "codepipeline":
		funcType = reflect.TypeOf(codepipeline.New)
	case "codestarnotifications":
		funcType = reflect.TypeOf(codestarnotifications.New)
	case "cognitoidentity":
		funcType = reflect.TypeOf(cognitoidentity.New)
	case "cognitoidentityprovider":
		funcType = reflect.TypeOf(cognitoidentityprovider.New)
	case "configservice":
		funcType = reflect.TypeOf(configservice.New)
	case "databasemigrationservice":
		funcType = reflect.TypeOf(databasemigrationservice.New)
	case "dataexchange":
		funcType = reflect.TypeOf(dataexchange.New)
	case "datapipeline":
		funcType = reflect.TypeOf(datapipeline.New)
	case "datasync":
		funcType = reflect.TypeOf(datasync.New)
	case "dax":
		funcType = reflect.TypeOf(dax.New)
	case "devicefarm":
		funcType = reflect.TypeOf(devicefarm.New)
	case "directconnect":
		funcType = reflect.TypeOf(directconnect.New)
	case "directoryservice":
		funcType = reflect.TypeOf(directoryservice.New)
	case "dlm":
		funcType = reflect.TypeOf(dlm.New)
	case "docdb":
		funcType = reflect.TypeOf(docdb.New)
	case "dynamodb":
		funcType = reflect.TypeOf(dynamodb.New)
	case "ec2":
		funcType = reflect.TypeOf(ec2.New)
	case "ecr":
		funcType = reflect.TypeOf(ecr.New)
	case "ecs":
		funcType = reflect.TypeOf(ecs.New)
	case "efs":
		funcType = reflect.TypeOf(efs.New)
	case "eks":
		funcType = reflect.TypeOf(eks.New)
	case "elasticache":
		funcType = reflect.TypeOf(elasticache.New)
	case "elasticbeanstalk":
		funcType = reflect.TypeOf(elasticbeanstalk.New)
	case "elasticsearchservice":
		funcType = reflect.TypeOf(elasticsearchservice.New)
	case "elb":
		funcType = reflect.TypeOf(elb.New)
	case "elbv2":
		funcType = reflect.TypeOf(elbv2.New)
	case "emr":
		funcType = reflect.TypeOf(emr.New)
	case "firehose":
		funcType = reflect.TypeOf(firehose.New)
	case "fsx":
		funcType = reflect.TypeOf(fsx.New)
	case "gamelift":
		funcType = reflect.TypeOf(gamelift.New)
	case "glacier":
		funcType = reflect.TypeOf(glacier.New)
	case "globalaccelerator":
		funcType = reflect.TypeOf(globalaccelerator.New)
	case "glue":
		funcType = reflect.TypeOf(glue.New)
	case "guardduty":
		funcType = reflect.TypeOf(guardduty.New)
	case "greengrass":
		funcType = reflect.TypeOf(greengrass.New)
	case "imagebuilder":
		funcType = reflect.TypeOf(imagebuilder.New)
	case "inspector":
		funcType = reflect.TypeOf(inspector.New)
	case "iot":
		funcType = reflect.TypeOf(iot.New)
	case "iotanalytics":
		funcType = reflect.TypeOf(iotanalytics.New)
	case "iotevents":
		funcType = reflect.TypeOf(iotevents.New)
	case "kafka":
		funcType = reflect.TypeOf(kafka.New)
	case "kinesis":
		funcType = reflect.TypeOf(kinesis.New)
	case "kinesisanalytics":
		funcType = reflect.TypeOf(kinesisanalytics.New)
	case "kinesisanalyticsv2":
		funcType = reflect.TypeOf(kinesisanalyticsv2.New)
	case "kinesisvideo":
		funcType = reflect.TypeOf(kinesisvideo.New)
	case "kms":
		funcType = reflect.TypeOf(kms.New)
	case "lambda":
		funcType = reflect.TypeOf(lambda.New)
	case "licensemanager":
		funcType = reflect.TypeOf(licensemanager.New)
	case "lightsail":
		funcType = reflect.TypeOf(lightsail.New)
	case "mediaconnect":
		funcType = reflect.TypeOf(mediaconnect.New)
	case "mediaconvert":
		funcType = reflect.TypeOf(mediaconvert.New)
	case "medialive":
		funcType = reflect.TypeOf(medialive.New)
	case "mediapackage":
		funcType = reflect.TypeOf(mediapackage.New)
	case "mediastore":
		funcType = reflect.TypeOf(mediastore.New)
	case "mq":
		funcType = reflect.TypeOf(mq.New)
	case "neptune":
		funcType = reflect.TypeOf(neptune.New)
	case "networkmanager":
		funcType = reflect.TypeOf(networkmanager.New)
	case "opsworks":
		funcType = reflect.TypeOf(opsworks.New)
	case "organizations":
		funcType = reflect.TypeOf(organizations.New)
	case "pinpoint":
		funcType = reflect.TypeOf(pinpoint.New)
	case "qldb":
		funcType = reflect.TypeOf(qldb.New)
	case "quicksight":
		funcType = reflect.TypeOf(quicksight.New)
	case "ram":
		funcType = reflect.TypeOf(ram.New)
	case "rds":
		funcType = reflect.TypeOf(rds.New)
	case "redshift":
		funcType = reflect.TypeOf(redshift.New)
	case "resourcegroups":
		funcType = reflect.TypeOf(resourcegroups.New)
	case "route53":
		funcType = reflect.TypeOf(route53.New)
	case "route53resolver":
		funcType = reflect.TypeOf(route53resolver.New)
	case "sagemaker":
		funcType = reflect.TypeOf(sagemaker.New)
	case "secretsmanager":
		funcType = reflect.TypeOf(secretsmanager.New)
	case "securityhub":
		funcType = reflect.TypeOf(securityhub.New)
	case "servicediscovery":
		funcType = reflect.TypeOf(servicediscovery.New)
	case "sfn":
		funcType = reflect.TypeOf(sfn.New)
	case "sns":
		funcType = reflect.TypeOf(sns.New)
	case "sqs":
		funcType = reflect.TypeOf(sqs.New)
	case "ssm":
		funcType = reflect.TypeOf(ssm.New)
	case "storagegateway":
		funcType = reflect.TypeOf(storagegateway.New)
	case "swf":
		funcType = reflect.TypeOf(swf.New)
	case "synthetics":
		funcType = reflect.TypeOf(synthetics.New)
	case "transfer":
		funcType = reflect.TypeOf(transfer.New)
	case "waf":
		funcType = reflect.TypeOf(waf.New)
	case "wafregional":
		funcType = reflect.TypeOf(wafregional.New)
	case "wafv2":
		funcType = reflect.TypeOf(wafv2.New)
	case "worklink":
		funcType = reflect.TypeOf(worklink.New)
	case "workspaces":
		funcType = reflect.TypeOf(workspaces.New)
	default:
		panic(fmt.Sprintf("unrecognized ServiceClientType: %s", serviceName))
	}

	return funcType.Out(0).String()
}

// ServiceListTagsFunction determines the service list tagging function.
func ServiceListTagsFunction(serviceName string) string {
	switch serviceName {
	case "acm":
		return "ListTagsForCertificate"
	case "acmpca":
		return "ListTags"
	case "apigatewayv2":
		return "GetTags"
	case "backup":
		return "ListTags"
	case "cloudhsmv2":
		return "ListTags"
	case "cloudtrail":
		return "ListTags"
	case "cloudwatchlogs":
		return "ListTagsLogGroup"
	case "dax":
		return "ListTags"
	case "directconnect":
		return "DescribeTags"
	case "dynamodb":
		return "ListTagsOfResource"
	case "ec2":
		return "DescribeTags"
	case "efs":
		return "DescribeTags"
	case "elasticsearchservice":
		return "ListTags"
	case "elb":
		return "DescribeTags"
	case "elbv2":
		return "DescribeTags"
	case "firehose":
		return "ListTagsForDeliveryStream"
	case "glacier":
		return "ListTagsForVault"
	case "glue":
		return "GetTags"
	case "kinesis":
		return "ListTagsForStream"
	case "kinesisvideo":
		return "ListTagsForStream"
	case "kms":
		return "ListResourceTags"
	case "lambda":
		return "ListTags"
	case "mq":
		return "ListTags"
	case "opsworks":
		return "ListTags"
	case "redshift":
		return "DescribeTags"
	case "resourcegroups":
		return "GetTags"
	case "sagemaker":
		return "ListTags"
	case "sqs":
		return "ListQueueTags"
	case "workspaces":
		return "DescribeTags"
	default:
		return "ListTagsForResource"
	}
}

// ServiceListTagsInputFilterIdentifierName determines the service list tag filter identifier field.
// This causes the implementation to use the Filters field with the Input struct.
func ServiceListTagsInputFilterIdentifierName(serviceName string) string {
	switch serviceName {
	case "ec2":
		return "resource-id"
	default:
		return ""
	}
}

// ServiceListTagsInputIdentifierField determines the service list tag identifier field.
func ServiceListTagsInputIdentifierField(serviceName string) string {
	switch serviceName {
	case "cloudtrail":
		return "ResourceIdList"
	case "directconnect":
		return "ResourceArns"
	case "efs":
		return "FileSystemId"
	case "workspaces":
		return "ResourceId"
	default:
		return ServiceTagInputIdentifierField(serviceName)
	}
}

// ServiceListTagInputIdentifierRequiresSlice determines if the service list tagging resource field requires a slice.
func ServiceListTagsInputIdentifierRequiresSlice(serviceName string) string {
	switch serviceName {
	case "cloudtrail":
		return "yes"
	case "directconnect":
		return "yes"
	case "elb":
		return "yes"
	case "elbv2":
		return "yes"
	default:
		return ""
	}
}

// ServiceListTagsInputResourceTypeField determines the service list tagging resource type field.
func ServiceListTagsInputResourceTypeField(serviceName string) string {
	switch serviceName {
	case "route53":
		return "ResourceType"
	case "ssm":
		return "ResourceType"
	default:
		return ""
	}
}

// ServiceListTagsOutputTagsField determines the service list tag field.
func ServiceListTagsOutputTagsField(serviceName string) string {
	switch serviceName {
	case "cloudfront":
		return "Tags.Items"
	case "cloudhsmv2":
		return "TagList"
	case "cloudtrail":
		return "ResourceTagList[0].TagsList"
	case "databasemigrationservice":
		return "TagList"
	case "directconnect":
		return "ResourceTags[0].Tags"
	case "docdb":
		return "TagList"
	case "elasticache":
		return "TagList"
	case "elasticbeanstalk":
		return "ResourceTags"
	case "elasticsearchservice":
		return "TagList"
	case "elb":
		return "TagDescriptions[0].Tags"
	case "elbv2":
		return "TagDescriptions[0].Tags"
	case "mediaconvert":
		return "ResourceTags.Tags"
	case "neptune":
		return "TagList"
	case "networkmanager":
		return "TagList"
	case "pinpoint":
		return "TagsModel.Tags"
	case "rds":
		return "TagList"
	case "route53":
		return "ResourceTagSet.Tags"
	case "ssm":
		return "TagList"
	case "waf":
		return "TagInfoForResource.TagList"
	case "wafregional":
		return "TagInfoForResource.TagList"
	case "wafv2":
		return "TagInfoForResource.TagList"
	case "workspaces":
		return "TagList"
	default:
		return "Tags"
	}
}

// ServiceResourceNotFoundErrorCode determines the error code of tagable resources when not found
func ServiceResourceNotFoundErrorCode(serviceName string) string {
	switch serviceName {
	default:
		return "ResourceNotFoundException"
	}
}

// ServiceResourceNotFoundErrorCode determines the common substring of error codes of tagable resources when not found
// This value takes precedence over ServiceResourceNotFoundErrorCode when defined for a service.
func ServiceResourceNotFoundErrorCodeContains(serviceName string) string {
	switch serviceName {
	case "ec2":
		return ".NotFound"
	default:
		return ""
	}
}

// ServiceRetryCreationOnResourceNotFound determines if tag creation should be retried when the tagable resource is not found
// This should only be used for services with eventual consistency considerations.
func ServiceRetryCreationOnResourceNotFound(serviceName string) string {
	switch serviceName {
	case "ec2":
		return "yes"
	default:
		return ""
	}
}

// ServiceTagFunction determines the service tagging function.
func ServiceTagFunction(serviceName string) string {
	switch serviceName {
	case "acm":
		return "AddTagsToCertificate"
	case "acmpca":
		return "TagCertificateAuthority"
	case "cloudtrail":
		return "AddTags"
	case "cloudwatchlogs":
		return "TagLogGroup"
	case "databasemigrationservice":
		return "AddTagsToResource"
	case "datapipeline":
		return "AddTags"
	case "directoryservice":
		return "AddTagsToResource"
	case "docdb":
		return "AddTagsToResource"
	case "ec2":
		return "CreateTags"
	case "elasticache":
		return "AddTagsToResource"
	case "elasticbeanstalk":
		return "UpdateTagsForResource"
	case "elasticsearchservice":
		return "AddTags"
	case "elb":
		return "AddTags"
	case "elbv2":
		return "AddTags"
	case "emr":
		return "AddTags"
	case "firehose":
		return "TagDeliveryStream"
	case "glacier":
		return "AddTagsToVault"
	case "kinesis":
		return "AddTagsToStream"
	case "kinesisvideo":
		return "TagStream"
	case "medialive":
		return "CreateTags"
	case "mq":
		return "CreateTags"
	case "neptune":
		return "AddTagsToResource"
	case "rds":
		return "AddTagsToResource"
	case "redshift":
		return "CreateTags"
	case "resourcegroups":
		return "Tag"
	case "route53":
		return "ChangeTagsForResource"
	case "sagemaker":
		return "AddTags"
	case "sqs":
		return "TagQueue"
	case "ssm":
		return "AddTagsToResource"
	case "storagegateway":
		return "AddTagsToResource"
	case "workspaces":
		return "CreateTags"
	default:
		return "TagResource"
	}
}

// ServiceTagFunctionBatchSize determines the batch size (if any) for tagging and untagging.
func ServiceTagFunctionBatchSize(serviceName string) string {
	switch serviceName {
	case "kinesis":
		return "10"
	default:
		return ""
	}
}

// ServiceTagInputIdentifierField determines the service tag identifier field.
func ServiceTagInputIdentifierField(serviceName string) string {
	switch serviceName {
	case "acm":
		return "CertificateArn"
	case "acmpca":
		return "CertificateAuthorityArn"
	case "athena":
		return "ResourceARN"
	case "cloud9":
		return "ResourceARN"
	case "cloudfront":
		return "Resource"
	case "cloudhsmv2":
		return "ResourceId"
	case "cloudtrail":
		return "ResourceId"
	case "cloudwatch":
		return "ResourceARN"
	case "cloudwatchevents":
		return "ResourceARN"
	case "cloudwatchlogs":
		return "LogGroupName"
	case "codestarnotifications":
		return "Arn"
	case "datapipeline":
		return "PipelineId"
	case "dax":
		return "ResourceName"
	case "devicefarm":
		return "ResourceARN"
	case "directoryservice":
		return "ResourceId"
	case "docdb":
		return "ResourceName"
	case "ec2":
		return "Resources"
	case "efs":
		return "ResourceId"
	case "elasticache":
		return "ResourceName"
	case "elasticsearchservice":
		return "ARN"
	case "elb":
		return "LoadBalancerNames"
	case "elbv2":
		return "ResourceArns"
	case "emr":
		return "ResourceId"
	case "firehose":
		return "DeliveryStreamName"
	case "fsx":
		return "ResourceARN"
	case "gamelift":
		return "ResourceARN"
	case "glacier":
		return "VaultName"
	case "kinesis":
		return "StreamName"
	case "kinesisanalytics":
		return "ResourceARN"
	case "kinesisanalyticsv2":
		return "ResourceARN"
	case "kinesisvideo":
		return "StreamARN"
	case "kms":
		return "KeyId"
	case "lambda":
		return "Resource"
	case "lightsail":
		return "ResourceName"
	case "mediaconvert":
		return "Arn"
	case "mediastore":
		return "Resource"
	case "neptune":
		return "ResourceName"
	case "organizations":
		return "ResourceId"
	case "ram":
		return "ResourceShareArn"
	case "rds":
		return "ResourceName"
	case "redshift":
		return "ResourceName"
	case "resourcegroups":
		return "Arn"
	case "route53":
		return "ResourceId"
	case "secretsmanager":
		return "SecretId"
	case "servicediscovery":
		return "ResourceARN"
	case "sqs":
		return "QueueUrl"
	case "ssm":
		return "ResourceId"
	case "storagegateway":
		return "ResourceARN"
	case "transfer":
		return "Arn"
	case "waf":
		return "ResourceARN"
	case "wafregional":
		return "ResourceARN"
	case "wafv2":
		return "ResourceARN"
	case "workspaces":
		return "ResourceId"
	default:
		return "ResourceArn"
	}
}

// ServiceTagInputIdentifierRequiresSlice determines if the service tagging resource field requires a slice.
func ServiceTagInputIdentifierRequiresSlice(serviceName string) string {
	switch serviceName {
	case "ec2":
		return "yes"
	case "elb":
		return "yes"
	case "elbv2":
		return "yes"
	default:
		return ""
	}
}

// ServiceTagInputTagsField determines the service tagging tags field.
func ServiceTagInputTagsField(serviceName string) string {
	switch serviceName {
	case "cloudhsmv2":
		return "TagList"
	case "cloudtrail":
		return "TagsList"
	case "elasticbeanstalk":
		return "TagsToAdd"
	case "elasticsearchservice":
		return "TagList"
	case "glue":
		return "TagsToAdd"
	case "pinpoint":
		return "TagsModel"
	case "route53":
		return "AddTags"
	default:
		return "Tags"
	}
}

// ServiceTagInputCustomValue determines any custom value for the service tagging tags field.
func ServiceTagInputCustomValue(serviceName string) string {
	switch serviceName {
	case "cloudfront":
		return "&cloudfront.Tags{Items: updatedTags.IgnoreAws().CloudfrontTags()}"
	case "kinesis":
		return "aws.StringMap(updatedTags.IgnoreAws().Map())"
	case "pinpoint":
		return "&pinpoint.TagsModel{Tags: updatedTags.IgnoreAws().PinpointTags()}"
	default:
		return ""
	}
}

// ServiceTagInputResourceTypeField determines the service tagging resource type field.
func ServiceTagInputResourceTypeField(serviceName string) string {
	switch serviceName {
	case "route53":
		return "ResourceType"
	case "ssm":
		return "ResourceType"
	default:
		return ""
	}
}

func ServiceTagPackage(serviceName string) string {
	switch serviceName {
	case "wafregional":
		return "waf"
	default:
		return serviceName
	}
}

// ServiceTagKeyType determines the service tagging tag key type.
func ServiceTagKeyType(serviceName string) string {
	switch serviceName {
	case "elb":
		return "TagKeyOnly"
	default:
		return ""
	}
}

// ServiceTagType determines the service tagging tag type.
func ServiceTagType(serviceName string) string {
	switch serviceName {
	case "appmesh":
		return "TagRef"
	case "datasync":
		return "TagListEntry"
	case "fms":
		return "ResourceTag"
	case "swf":
		return "ResourceTag"
	default:
		return "Tag"
	}
}

// ServiceTagType2 determines if the service tagging has a second tag type.
// The two types must be equivalent.
func ServiceTagType2(serviceName string) string {
	switch serviceName {
	case "ec2":
		return "TagDescription"
	default:
		return ""
	}
}

// ServiceTagTypeKeyField determines the service tagging tag type key field.
func ServiceTagTypeKeyField(serviceName string) string {
	switch serviceName {
	case "kms":
		return "TagKey"
	default:
		return "Key"
	}
}

// ServiceTagTypeValueField determines the service tagging tag type value field.
func ServiceTagTypeValueField(serviceName string) string {
	switch serviceName {
	case "kms":
		return "TagValue"
	default:
		return "Value"
	}
}

// ServiceUntagFunction determines the service untagging function.
func ServiceUntagFunction(serviceName string) string {
	switch serviceName {
	case "acm":
		return "RemoveTagsFromCertificate"
	case "acmpca":
		return "UntagCertificateAuthority"
	case "cloudtrail":
		return "RemoveTags"
	case "cloudwatchlogs":
		return "UntagLogGroup"
	case "databasemigrationservice":
		return "RemoveTagsFromResource"
	case "datapipeline":
		return "RemoveTags"
	case "directoryservice":
		return "RemoveTagsFromResource"
	case "docdb":
		return "RemoveTagsFromResource"
	case "ec2":
		return "DeleteTags"
	case "elasticache":
		return "RemoveTagsFromResource"
	case "elasticbeanstalk":
		return "UpdateTagsForResource"
	case "elasticsearchservice":
		return "RemoveTags"
	case "elb":
		return "RemoveTags"
	case "elbv2":
		return "RemoveTags"
	case "emr":
		return "RemoveTags"
	case "firehose":
		return "UntagDeliveryStream"
	case "glacier":
		return "RemoveTagsFromVault"
	case "kinesis":
		return "RemoveTagsFromStream"
	case "kinesisvideo":
		return "UntagStream"
	case "medialive":
		return "DeleteTags"
	case "mq":
		return "DeleteTags"
	case "neptune":
		return "RemoveTagsFromResource"
	case "rds":
		return "RemoveTagsFromResource"
	case "redshift":
		return "DeleteTags"
	case "resourcegroups":
		return "Untag"
	case "route53":
		return "ChangeTagsForResource"
	case "sagemaker":
		return "DeleteTags"
	case "sqs":
		return "UntagQueue"
	case "ssm":
		return "RemoveTagsFromResource"
	case "storagegateway":
		return "RemoveTagsFromResource"
	case "workspaces":
		return "DeleteTags"
	default:
		return "UntagResource"
	}
}

// ServiceUntagInputRequiresTagType determines if the service untagging requires full Tag type.
func ServiceUntagInputRequiresTagType(serviceName string) string {
	switch serviceName {
	case "acm":
		return "yes"
	case "acmpca":
		return "yes"
	case "cloudtrail":
		return "yes"
	case "ec2":
		return "yes"
	default:
		return ""
	}
}

// ServiceUntagInputRequiresTagKeyType determines if a special type for the untagging function tag key field is needed.
func ServiceUntagInputRequiresTagKeyType(serviceName string) string {
	switch serviceName {
	case "elb":
		return "yes"
	default:
		return ""
	}
}

// ServiceUntagInputTagsField determines the service untagging tags field.
func ServiceUntagInputTagsField(serviceName string) string {
	switch serviceName {
	case "acm":
		return "Tags"
	case "acmpca":
		return "Tags"
	case "backup":
		return "TagKeyList"
	case "cloudhsmv2":
		return "TagKeyList"
	case "cloudtrail":
		return "TagsList"
	case "cloudwatchlogs":
		return "Tags"
	case "datasync":
		return "Keys"
	case "ec2":
		return "Tags"
	case "elasticbeanstalk":
		return "TagsToRemove"
	case "elb":
		return "Tags"
	case "glue":
		return "TagsToRemove"
	case "kinesisvideo":
		return "TagKeyList"
	case "resourcegroups":
		return "Keys"
	case "route53":
		return "RemoveTagKeys"
	default:
		return "TagKeys"
	}
}

// ServiceUntagInputCustomValue determines any custom value for the service untagging tags field.
func ServiceUntagInputCustomValue(serviceName string) string {
	switch serviceName {
	case "cloudfront":
		return "&cloudfront.TagKeys{Items: aws.StringSlice(removedTags.IgnoreAws().Keys())}"
	default:
		return ""
	}
}
