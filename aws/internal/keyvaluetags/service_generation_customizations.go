// This file contains code generation customizations for each AWS Go SDK service.

package keyvaluetags

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/quicksight"
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
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/aws/aws-sdk-go/service/qldb"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/aws/aws-sdk-go/service/resourcegroups"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/aws/aws-sdk-go/service/swf"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/aws/aws-sdk-go/service/wafv2"
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
	case "transfer":
		funcType = reflect.TypeOf(transfer.New)
	case "waf":
		funcType = reflect.TypeOf(waf.New)
	case "wafregional":
		funcType = reflect.TypeOf(wafregional.New)
	case "wafv2":
		funcType = reflect.TypeOf(wafv2.New)
	case "workspaces":
		funcType = reflect.TypeOf(workspaces.New)
	default:
		panic(fmt.Sprintf("unrecognized ServiceClientType: %s", serviceName))
	}

	return funcType.Out(0).String()
}

func ServiceTagPackage(serviceName string) string {
	switch serviceName {
	case "wafregional":
		return "waf"
	default:
		return serviceName
	}
}
