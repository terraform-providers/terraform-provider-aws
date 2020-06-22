package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/accessanalyzer"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go/service/applicationinsights"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscalingplans"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/aws/aws-sdk-go/service/cloud9"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	"github.com/aws/aws-sdk-go/service/cloudsearch"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/aws/aws-sdk-go/service/codedeploy"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/aws/aws-sdk-go/service/codestarnotifications"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/aws/aws-sdk-go/service/costandusagereportservice"
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
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/aws/aws-sdk-go/service/elastictranscoder"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/fms"
	"github.com/aws/aws-sdk-go/service/forecastservice"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/aws/aws-sdk-go/service/glacier"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/aws/aws-sdk-go/service/iam"
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
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/aws/aws-sdk-go/service/licensemanager"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/aws/aws-sdk-go/service/macie"
	"github.com/aws/aws-sdk-go/service/managedblockchain"
	"github.com/aws/aws-sdk-go/service/marketplacecatalog"
	"github.com/aws/aws-sdk-go/service/mediaconnect"
	"github.com/aws/aws-sdk-go/service/mediaconvert"
	"github.com/aws/aws-sdk-go/service/medialive"
	"github.com/aws/aws-sdk-go/service/mediapackage"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/aws/aws-sdk-go/service/mediastoredata"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/aws/aws-sdk-go/service/personalize"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/aws/aws-sdk-go/service/qldb"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/aws/aws-sdk-go/service/resourcegroups"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53domains"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/aws/aws-sdk-go/service/serverlessapplicationrepository"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/aws/aws-sdk-go/service/servicequotas"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/swf"
	"github.com/aws/aws-sdk-go/service/synthetics"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/aws/aws-sdk-go/service/worklink"
	"github.com/aws/aws-sdk-go/service/workmail"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/aws/aws-sdk-go/service/xray"
	awsbase "github.com/hashicorp/aws-sdk-go-base"
	"github.com/hashicorp/terraform-plugin-sdk/helper/logging"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

type Config struct {
	AccessKey     string
	SecretKey     string
	CredsFilename string
	Profile       string
	Token         string
	Region        string
	MaxRetries    int

	AssumeRoleARN         string
	AssumeRoleExternalID  string
	AssumeRoleSessionName string
	AssumeRolePolicy      string

	AllowedAccountIds   []string
	ForbiddenAccountIds []string

	Endpoints        map[string]string
	IgnoreTagsConfig *keyvaluetags.IgnoreConfig
	Insecure         bool

	SkipCredsValidation     bool
	SkipGetEC2Platforms     bool
	SkipRegionValidation    bool
	SkipRequestingAccountId bool
	SkipMetadataApiCheck    bool
	S3ForcePathStyle        bool

	terraformVersion string
}

type AWSClient struct {
	accessanalyzerconn                  *accessanalyzer.AccessAnalyzer
	accountid                           string
	acmconn                             *acm.ACM
	acmpcaconn                          *acmpca.ACMPCA
	amplifyconn                         *amplify.Amplify
	apigatewayconn                      *apigateway.APIGateway
	apigatewayv2conn                    *apigatewayv2.ApiGatewayV2
	appautoscalingconn                  *applicationautoscaling.ApplicationAutoScaling
	applicationinsightsconn             *applicationinsights.ApplicationInsights
	appmeshconn                         *appmesh.AppMesh
	appstreamconn                       *appstream.AppStream
	appsyncconn                         *appsync.AppSync
	athenaconn                          *athena.Athena
	autoscalingconn                     *autoscaling.AutoScaling
	autoscalingplansconn                *autoscalingplans.AutoScalingPlans
	backupconn                          *backup.Backup
	batchconn                           *batch.Batch
	budgetconn                          *budgets.Budgets
	cfconn                              *cloudformation.CloudFormation
	cloud9conn                          *cloud9.Cloud9
	cloudfrontconn                      *cloudfront.CloudFront
	cloudhsmv2conn                      *cloudhsmv2.CloudHSMV2
	cloudsearchconn                     *cloudsearch.CloudSearch
	cloudtrailconn                      *cloudtrail.CloudTrail
	cloudwatchconn                      *cloudwatch.CloudWatch
	cloudwatcheventsconn                *cloudwatchevents.CloudWatchEvents
	cloudwatchlogsconn                  *cloudwatchlogs.CloudWatchLogs
	codeartifactconn                    *codeartifact.CodeArtifact
	codebuildconn                       *codebuild.CodeBuild
	codecommitconn                      *codecommit.CodeCommit
	codedeployconn                      *codedeploy.CodeDeploy
	codepipelineconn                    *codepipeline.CodePipeline
	codestarnotificationsconn           *codestarnotifications.CodeStarNotifications
	cognitoconn                         *cognitoidentity.CognitoIdentity
	cognitoidpconn                      *cognitoidentityprovider.CognitoIdentityProvider
	configconn                          *configservice.ConfigService
	costandusagereportconn              *costandusagereportservice.CostandUsageReportService
	dataexchangeconn                    *dataexchange.DataExchange
	datapipelineconn                    *datapipeline.DataPipeline
	datasyncconn                        *datasync.DataSync
	daxconn                             *dax.DAX
	devicefarmconn                      *devicefarm.DeviceFarm
	dlmconn                             *dlm.DLM
	dmsconn                             *databasemigrationservice.DatabaseMigrationService
	dnsSuffix                           string
	docdbconn                           *docdb.DocDB
	dsconn                              *directoryservice.DirectoryService
	dxconn                              *directconnect.DirectConnect
	dynamodbconn                        *dynamodb.DynamoDB
	ec2conn                             *ec2.EC2
	ecrconn                             *ecr.ECR
	ecsconn                             *ecs.ECS
	efsconn                             *efs.EFS
	eksconn                             *eks.EKS
	elasticacheconn                     *elasticache.ElastiCache
	elasticbeanstalkconn                *elasticbeanstalk.ElasticBeanstalk
	elastictranscoderconn               *elastictranscoder.ElasticTranscoder
	elbconn                             *elb.ELB
	elbv2conn                           *elbv2.ELBV2
	emrconn                             *emr.EMR
	esconn                              *elasticsearch.ElasticsearchService
	firehoseconn                        *firehose.Firehose
	fmsconn                             *fms.FMS
	forecastconn                        *forecastservice.ForecastService
	fsxconn                             *fsx.FSx
	gameliftconn                        *gamelift.GameLift
	glacierconn                         *glacier.Glacier
	globalacceleratorconn               *globalaccelerator.GlobalAccelerator
	glueconn                            *glue.Glue
	guarddutyconn                       *guardduty.GuardDuty
	greengrassconn                      *greengrass.Greengrass
	iamconn                             *iam.IAM
	IgnoreTagsConfig                    *keyvaluetags.IgnoreConfig
	imagebuilderconn                    *imagebuilder.Imagebuilder
	inspectorconn                       *inspector.Inspector
	iotconn                             *iot.IoT
	iotanalyticsconn                    *iotanalytics.IoTAnalytics
	ioteventsconn                       *iotevents.IoTEvents
	kafkaconn                           *kafka.Kafka
	kinesisanalyticsconn                *kinesisanalytics.KinesisAnalytics
	kinesisanalyticsv2conn              *kinesisanalyticsv2.KinesisAnalyticsV2
	kinesisconn                         *kinesis.Kinesis
	kinesisvideoconn                    *kinesisvideo.KinesisVideo
	kmsconn                             *kms.KMS
	lakeformationconn                   *lakeformation.LakeFormation
	lambdaconn                          *lambda.Lambda
	lexmodelconn                        *lexmodelbuildingservice.LexModelBuildingService
	licensemanagerconn                  *licensemanager.LicenseManager
	lightsailconn                       *lightsail.Lightsail
	macieconn                           *macie.Macie
	managedblockchainconn               *managedblockchain.ManagedBlockchain
	marketplacecatalogconn              *marketplacecatalog.MarketplaceCatalog
	mediaconnectconn                    *mediaconnect.MediaConnect
	mediaconvertconn                    *mediaconvert.MediaConvert
	mediaconvertaccountconn             *mediaconvert.MediaConvert
	medialiveconn                       *medialive.MediaLive
	mediapackageconn                    *mediapackage.MediaPackage
	mediastoreconn                      *mediastore.MediaStore
	mediastoredataconn                  *mediastoredata.MediaStoreData
	mqconn                              *mq.MQ
	neptuneconn                         *neptune.Neptune
	networkmanagerconn                  *networkmanager.NetworkManager
	opsworksconn                        *opsworks.OpsWorks
	organizationsconn                   *organizations.Organizations
	outpostsconn                        *outposts.Outposts
	partition                           string
	personalizeconn                     *personalize.Personalize
	pinpointconn                        *pinpoint.Pinpoint
	pricingconn                         *pricing.Pricing
	qldbconn                            *qldb.QLDB
	quicksightconn                      *quicksight.QuickSight
	r53conn                             *route53.Route53
	ramconn                             *ram.RAM
	rdsconn                             *rds.RDS
	redshiftconn                        *redshift.Redshift
	region                              string
	resourcegroupsconn                  *resourcegroups.ResourceGroups
	route53domainsconn                  *route53domains.Route53Domains
	route53resolverconn                 *route53resolver.Route53Resolver
	s3conn                              *s3.S3
	s3connUriCleaningDisabled           *s3.S3
	s3controlconn                       *s3control.S3Control
	sagemakerconn                       *sagemaker.SageMaker
	scconn                              *servicecatalog.ServiceCatalog
	sdconn                              *servicediscovery.ServiceDiscovery
	secretsmanagerconn                  *secretsmanager.SecretsManager
	securityhubconn                     *securityhub.SecurityHub
	serverlessapplicationrepositoryconn *serverlessapplicationrepository.ServerlessApplicationRepository
	servicequotasconn                   *servicequotas.ServiceQuotas
	sesconn                             *ses.SES
	sfnconn                             *sfn.SFN
	shieldconn                          *shield.Shield
	simpledbconn                        *simpledb.SimpleDB
	snsconn                             *sns.SNS
	sqsconn                             *sqs.SQS
	ssmconn                             *ssm.SSM
	storagegatewayconn                  *storagegateway.StorageGateway
	stsconn                             *sts.STS
	supportedplatforms                  []string
	swfconn                             *swf.SWF
	syntheticsconn                      *synthetics.Synthetics
	terraformVersion                    string
	transferconn                        *transfer.Transfer
	wafconn                             *waf.WAF
	wafregionalconn                     *wafregional.WAFRegional
	wafv2conn                           *wafv2.WAFV2
	worklinkconn                        *worklink.WorkLink
	workmailconn                        *workmail.WorkMail
	workspacesconn                      *workspaces.WorkSpaces
	xrayconn                            *xray.XRay
}

// PartitionHostname returns a hostname with the provider domain suffix for the partition
// e.g. PREFIX.amazonaws.com
// The prefix should not contain a trailing period.
func (client *AWSClient) PartitionHostname(prefix string) string {
	return fmt.Sprintf("%s.%s", prefix, client.dnsSuffix)
}

// RegionalHostname returns a hostname with the provider domain suffix for the region and partition
// e.g. PREFIX.us-west-2.amazonaws.com
// The prefix should not contain a trailing period.
func (client *AWSClient) RegionalHostname(prefix string) string {
	return fmt.Sprintf("%s.%s.%s", prefix, client.region, client.dnsSuffix)
}

// Client configures and returns a fully initialized AWSClient
func (c *Config) Client() (interface{}, error) {
	// Get the auth and region. This can fail if keys/regions were not
	// specified and we're attempting to use the environment.
	if !c.SkipRegionValidation {
		if err := awsbase.ValidateRegion(c.Region); err != nil {
			return nil, err
		}
	}

	log.Println("[INFO] Building AWS auth structure")
	awsbaseConfig := &awsbase.Config{
		AccessKey:               c.AccessKey,
		AssumeRoleARN:           c.AssumeRoleARN,
		AssumeRoleExternalID:    c.AssumeRoleExternalID,
		AssumeRolePolicy:        c.AssumeRolePolicy,
		AssumeRoleSessionName:   c.AssumeRoleSessionName,
		CredsFilename:           c.CredsFilename,
		DebugLogging:            logging.IsDebugOrHigher(),
		IamEndpoint:             c.Endpoints["iam"],
		Insecure:                c.Insecure,
		MaxRetries:              c.MaxRetries,
		Profile:                 c.Profile,
		Region:                  c.Region,
		SecretKey:               c.SecretKey,
		SkipCredsValidation:     c.SkipCredsValidation,
		SkipMetadataApiCheck:    c.SkipMetadataApiCheck,
		SkipRequestingAccountId: c.SkipRequestingAccountId,
		StsEndpoint:             c.Endpoints["sts"],
		Token:                   c.Token,
		UserAgentProducts: []*awsbase.UserAgentProduct{
			{Name: "APN", Version: "1.0"},
			{Name: "HashiCorp", Version: "1.0"},
			{Name: "Terraform", Version: c.terraformVersion,
				Extra: []string{"+https://www.terraform.io"}},
		},
	}

	sess, accountID, partition, err := awsbase.GetSessionWithAccountIDAndPartition(awsbaseConfig)
	if err != nil {
		return nil, err
	}

	if accountID == "" {
		log.Printf("[WARN] AWS account ID not found for provider. See https://www.terraform.io/docs/providers/aws/index.html#skip_requesting_account_id for implications.")
	}

	if err := awsbase.ValidateAccountID(accountID, c.AllowedAccountIds, c.ForbiddenAccountIds); err != nil {
		return nil, err
	}

	dnsSuffix := "amazonaws.com"
	if p, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), c.Region); ok {
		dnsSuffix = p.DNSSuffix()
	}

	client := &AWSClient{
		accessanalyzerconn:                  accessanalyzer.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["accessanalyzer"])})),
		accountid:                           accountID,
		acmconn:                             acm.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["acm"])})),
		acmpcaconn:                          acmpca.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["acmpca"])})),
		amplifyconn:                         amplify.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["amplify"])})),
		apigatewayconn:                      apigateway.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["apigateway"])})),
		apigatewayv2conn:                    apigatewayv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["apigateway"])})),
		appautoscalingconn:                  applicationautoscaling.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["applicationautoscaling"])})),
		applicationinsightsconn:             applicationinsights.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["applicationinsights"])})),
		appmeshconn:                         appmesh.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["appmesh"])})),
		appstreamconn:                       appstream.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["appstream"])})),
		appsyncconn:                         appsync.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["appsync"])})),
		athenaconn:                          athena.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["athena"])})),
		autoscalingconn:                     autoscaling.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["autoscaling"])})),
		autoscalingplansconn:                autoscalingplans.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["autoscalingplans"])})),
		backupconn:                          backup.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["backup"])})),
		batchconn:                           batch.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["batch"])})),
		budgetconn:                          budgets.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["budgets"])})),
		cfconn:                              cloudformation.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cloudformation"])})),
		cloud9conn:                          cloud9.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cloud9"])})),
		cloudfrontconn:                      cloudfront.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cloudfront"])})),
		cloudhsmv2conn:                      cloudhsmv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cloudhsm"])})),
		cloudsearchconn:                     cloudsearch.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cloudsearch"])})),
		cloudtrailconn:                      cloudtrail.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cloudtrail"])})),
		cloudwatchconn:                      cloudwatch.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cloudwatch"])})),
		cloudwatcheventsconn:                cloudwatchevents.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cloudwatchevents"])})),
		cloudwatchlogsconn:                  cloudwatchlogs.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cloudwatchlogs"])})),
		codeartifactconn:                    codeartifact.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["codeartifact"])})),
		codebuildconn:                       codebuild.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["codebuild"])})),
		codecommitconn:                      codecommit.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["codecommit"])})),
		codedeployconn:                      codedeploy.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["codedeploy"])})),
		codepipelineconn:                    codepipeline.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["codepipeline"])})),
		codestarnotificationsconn:           codestarnotifications.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["codestarnotifications"])})),
		cognitoconn:                         cognitoidentity.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cognitoidentity"])})),
		cognitoidpconn:                      cognitoidentityprovider.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cognitoidp"])})),
		configconn:                          configservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["configservice"])})),
		costandusagereportconn:              costandusagereportservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cur"])})),
		dataexchangeconn:                    dataexchange.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["dataexchange"])})),
		datapipelineconn:                    datapipeline.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["datapipeline"])})),
		datasyncconn:                        datasync.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["datasync"])})),
		daxconn:                             dax.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["dax"])})),
		devicefarmconn:                      devicefarm.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["devicefarm"])})),
		dlmconn:                             dlm.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["dlm"])})),
		dmsconn:                             databasemigrationservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["dms"])})),
		dnsSuffix:                           dnsSuffix,
		docdbconn:                           docdb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["docdb"])})),
		dsconn:                              directoryservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["ds"])})),
		dxconn:                              directconnect.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["directconnect"])})),
		dynamodbconn:                        dynamodb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["dynamodb"])})),
		ec2conn:                             ec2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["ec2"])})),
		ecrconn:                             ecr.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["ecr"])})),
		ecsconn:                             ecs.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["ecs"])})),
		efsconn:                             efs.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["efs"])})),
		eksconn:                             eks.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["eks"])})),
		elasticacheconn:                     elasticache.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["elasticache"])})),
		elasticbeanstalkconn:                elasticbeanstalk.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["elasticbeanstalk"])})),
		elastictranscoderconn:               elastictranscoder.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["elastictranscoder"])})),
		elbconn:                             elb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["elb"])})),
		elbv2conn:                           elbv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["elb"])})),
		emrconn:                             emr.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["emr"])})),
		esconn:                              elasticsearch.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["es"])})),
		firehoseconn:                        firehose.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["firehose"])})),
		fmsconn:                             fms.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["fms"])})),
		forecastconn:                        forecastservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["forecast"])})),
		fsxconn:                             fsx.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["fsx"])})),
		gameliftconn:                        gamelift.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["gamelift"])})),
		glacierconn:                         glacier.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["glacier"])})),
		glueconn:                            glue.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["glue"])})),
		guarddutyconn:                       guardduty.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["guardduty"])})),
		greengrassconn:                      greengrass.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["greengrass"])})),
		iamconn:                             iam.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["iam"])})),
		IgnoreTagsConfig:                    c.IgnoreTagsConfig,
		imagebuilderconn:                    imagebuilder.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["imagebuilder"])})),
		inspectorconn:                       inspector.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["inspector"])})),
		iotconn:                             iot.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["iot"])})),
		iotanalyticsconn:                    iotanalytics.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["iotanalytics"])})),
		ioteventsconn:                       iotevents.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["iotevents"])})),
		kafkaconn:                           kafka.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["kafka"])})),
		kinesisanalyticsconn:                kinesisanalytics.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["kinesisanalytics"])})),
		kinesisanalyticsv2conn:              kinesisanalyticsv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["kinesisanalyticsv2"])})),
		kinesisconn:                         kinesis.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["kinesis"])})),
		kinesisvideoconn:                    kinesisvideo.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["kinesisvideo"])})),
		kmsconn:                             kms.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["kms"])})),
		lakeformationconn:                   lakeformation.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["lakeformation"])})),
		lambdaconn:                          lambda.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["lambda"])})),
		lexmodelconn:                        lexmodelbuildingservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["lexmodels"])})),
		licensemanagerconn:                  licensemanager.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["licensemanager"])})),
		lightsailconn:                       lightsail.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["lightsail"])})),
		macieconn:                           macie.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["macie"])})),
		managedblockchainconn:               managedblockchain.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["managedblockchain"])})),
		marketplacecatalogconn:              marketplacecatalog.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["marketplacecatalog"])})),
		mediaconnectconn:                    mediaconnect.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["mediaconnect"])})),
		mediaconvertconn:                    mediaconvert.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["mediaconvert"])})),
		medialiveconn:                       medialive.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["medialive"])})),
		mediapackageconn:                    mediapackage.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["mediapackage"])})),
		mediastoreconn:                      mediastore.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["mediastore"])})),
		mediastoredataconn:                  mediastoredata.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["mediastoredata"])})),
		mqconn:                              mq.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["mq"])})),
		neptuneconn:                         neptune.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["neptune"])})),
		networkmanagerconn:                  networkmanager.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["networkmanager"])})),
		opsworksconn:                        opsworks.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["opsworks"])})),
		organizationsconn:                   organizations.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["organizations"])})),
		outpostsconn:                        outposts.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["outposts"])})),
		partition:                           partition,
		personalizeconn:                     personalize.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["personalize"])})),
		pinpointconn:                        pinpoint.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["pinpoint"])})),
		pricingconn:                         pricing.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["pricing"])})),
		qldbconn:                            qldb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["qldb"])})),
		quicksightconn:                      quicksight.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["quicksight"])})),
		ramconn:                             ram.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["ram"])})),
		rdsconn:                             rds.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["rds"])})),
		redshiftconn:                        redshift.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["redshift"])})),
		region:                              c.Region,
		resourcegroupsconn:                  resourcegroups.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["resourcegroups"])})),
		route53domainsconn:                  route53domains.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["route53domains"])})),
		route53resolverconn:                 route53resolver.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["route53resolver"])})),
		s3controlconn:                       s3control.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["s3control"])})),
		sagemakerconn:                       sagemaker.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["sagemaker"])})),
		scconn:                              servicecatalog.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["servicecatalog"])})),
		sdconn:                              servicediscovery.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["servicediscovery"])})),
		secretsmanagerconn:                  secretsmanager.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["secretsmanager"])})),
		securityhubconn:                     securityhub.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["securityhub"])})),
		serverlessapplicationrepositoryconn: serverlessapplicationrepository.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["serverlessrepo"])})),
		servicequotasconn:                   servicequotas.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["servicequotas"])})),
		sesconn:                             ses.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["ses"])})),
		sfnconn:                             sfn.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["stepfunctions"])})),
		simpledbconn:                        simpledb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["sdb"])})),
		snsconn:                             sns.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["sns"])})),
		sqsconn:                             sqs.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["sqs"])})),
		ssmconn:                             ssm.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["ssm"])})),
		storagegatewayconn:                  storagegateway.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["storagegateway"])})),
		stsconn:                             sts.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["sts"])})),
		swfconn:                             swf.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["swf"])})),
		syntheticsconn:                      synthetics.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["synthetics"])})),
		terraformVersion:                    c.terraformVersion,
		transferconn:                        transfer.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["transfer"])})),
		wafconn:                             waf.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["waf"])})),
		wafregionalconn:                     wafregional.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["wafregional"])})),
		wafv2conn:                           wafv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["wafv2"])})),
		worklinkconn:                        worklink.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["worklink"])})),
		workmailconn:                        workmail.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["workmail"])})),
		workspacesconn:                      workspaces.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["workspaces"])})),
		xrayconn:                            xray.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["xray"])})),
	}

	// "Global" services that require customizations
	globalAcceleratorConfig := &aws.Config{
		Endpoint: aws.String(c.Endpoints["globalaccelerator"]),
	}
	route53Config := &aws.Config{
		Endpoint: aws.String(c.Endpoints["route53"]),
	}
	shieldConfig := &aws.Config{
		Endpoint: aws.String(c.Endpoints["shield"]),
	}

	// Services that require multiple client configurations
	s3Config := &aws.Config{
		Endpoint:         aws.String(c.Endpoints["s3"]),
		S3ForcePathStyle: aws.Bool(c.S3ForcePathStyle),
	}

	client.s3conn = s3.New(sess.Copy(s3Config))

	s3Config.DisableRestProtocolURICleaning = aws.Bool(true)
	client.s3connUriCleaningDisabled = s3.New(sess.Copy(s3Config))

	// Handle deprecated endpoint configurations
	if c.Endpoints["kinesis_analytics"] != "" {
		client.kinesisanalyticsconn = kinesisanalytics.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["kinesis_analytics"])}))
	}
	if c.Endpoints["r53"] != "" {
		route53Config.Endpoint = aws.String(c.Endpoints["r53"])
	}

	// Force "global" services to correct regions
	switch partition {
	case endpoints.AwsPartitionID:
		globalAcceleratorConfig.Region = aws.String(endpoints.UsWest2RegionID)
		route53Config.Region = aws.String(endpoints.UsEast1RegionID)
		shieldConfig.Region = aws.String(endpoints.UsEast1RegionID)
	case endpoints.AwsCnPartitionID:
		// The AWS Go SDK is missing endpoint information for Route 53 in the AWS China partition.
		// This can likely be removed in the future.
		if aws.StringValue(route53Config.Endpoint) == "" {
			route53Config.Endpoint = aws.String("https://api.route53.cn")
		}
		route53Config.Region = aws.String(endpoints.CnNorthwest1RegionID)
	case endpoints.AwsUsGovPartitionID:
		route53Config.Region = aws.String(endpoints.UsGovWest1RegionID)
	}

	client.globalacceleratorconn = globalaccelerator.New(sess.Copy(globalAcceleratorConfig))
	client.r53conn = route53.New(sess.Copy(route53Config))
	client.shieldconn = shield.New(sess.Copy(shieldConfig))

	// Workaround for https://github.com/aws/aws-sdk-go/issues/1472
	client.appautoscalingconn.Handlers.Retry.PushBack(func(r *request.Request) {
		if !strings.HasPrefix(r.Operation.Name, "Describe") && !strings.HasPrefix(r.Operation.Name, "List") {
			return
		}
		err, ok := r.Error.(awserr.Error)
		if !ok || err == nil {
			return
		}
		if err.Code() == applicationautoscaling.ErrCodeFailedResourceAccessException {
			r.Retryable = aws.Bool(true)
		}
	})

	client.appsyncconn.Handlers.Retry.PushBack(func(r *request.Request) {
		if r.Operation.Name == "CreateGraphqlApi" {
			if isAWSErr(r.Error, appsync.ErrCodeConcurrentModificationException, "a GraphQL API creation is already in progress") {
				r.Retryable = aws.Bool(true)
			}
		}
	})

	client.configconn.Handlers.Retry.PushBack(func(r *request.Request) {
		// When calling Config Organization Rules API actions immediately
		// after Organization creation, the API can randomly return the
		// OrganizationAccessDeniedException error for a few minutes, even
		// after succeeding a few requests.
		switch r.Operation.Name {
		case "DeleteOrganizationConfigRule", "DescribeOrganizationConfigRules", "DescribeOrganizationConfigRuleStatuses", "PutOrganizationConfigRule":
			if !isAWSErr(r.Error, configservice.ErrCodeOrganizationAccessDeniedException, "This action can be only made by AWS Organization's master account.") {
				return
			}

			// We only want to retry briefly as the default max retry count would
			// excessively retry when the error could be legitimate.
			// We currently depend on the DefaultRetryer exponential backoff here.
			// ~10 retries gives a fair backoff of a few seconds.
			if r.RetryCount < 9 {
				r.Retryable = aws.Bool(true)
			} else {
				r.Retryable = aws.Bool(false)
			}
		}
	})

	// See https://github.com/aws/aws-sdk-go/pull/1276
	client.dynamodbconn.Handlers.Retry.PushBack(func(r *request.Request) {
		if r.Operation.Name != "PutItem" && r.Operation.Name != "UpdateItem" && r.Operation.Name != "DeleteItem" {
			return
		}
		if isAWSErr(r.Error, dynamodb.ErrCodeLimitExceededException, "Subscriber limit exceeded:") {
			r.Retryable = aws.Bool(true)
		}
	})

	client.ec2conn.Handlers.Retry.PushBack(func(r *request.Request) {
		if r.Operation.Name == "CreateClientVpnEndpoint" {
			if isAWSErr(r.Error, "OperationNotPermitted", "Endpoint cannot be created while another endpoint is being created") {
				r.Retryable = aws.Bool(true)
			}
		}

		if r.Operation.Name == "CreateVpnConnection" {
			if isAWSErr(r.Error, "VpnConnectionLimitExceeded", "maximum number of mutating objects has been reached") {
				r.Retryable = aws.Bool(true)
			}
		}

		if r.Operation.Name == "CreateVpnGateway" {
			if isAWSErr(r.Error, "VpnGatewayLimitExceeded", "maximum number of mutating objects has been reached") {
				r.Retryable = aws.Bool(true)
			}
		}

		if r.Operation.Name == "AttachVpnGateway" || r.Operation.Name == "DetachVpnGateway" {
			if isAWSErr(r.Error, "InvalidParameterValue", "This call cannot be completed because there are pending VPNs or Virtual Interfaces") {
				r.Retryable = aws.Bool(true)
			}
		}
	})

	client.kafkaconn.Handlers.Retry.PushBack(func(r *request.Request) {
		if isAWSErr(r.Error, kafka.ErrCodeTooManyRequestsException, "Too Many Requests") {
			r.Retryable = aws.Bool(true)
		}
	})

	client.kinesisconn.Handlers.Retry.PushBack(func(r *request.Request) {
		if r.Operation.Name == "CreateStream" {
			if isAWSErr(r.Error, kinesis.ErrCodeLimitExceededException, "simultaneously be in CREATING or DELETING") {
				r.Retryable = aws.Bool(true)
			}
		}
		if r.Operation.Name == "CreateStream" || r.Operation.Name == "DeleteStream" {
			if isAWSErr(r.Error, kinesis.ErrCodeLimitExceededException, "Rate exceeded for stream") {
				r.Retryable = aws.Bool(true)
			}
		}
	})

	client.organizationsconn.Handlers.Retry.PushBack(func(r *request.Request) {
		// Retry on the following error:
		// ConcurrentModificationException: AWS Organizations can't complete your request because it conflicts with another attempt to modify the same entity. Try again later.
		if isAWSErr(r.Error, organizations.ErrCodeConcurrentModificationException, "Try again later") {
			r.Retryable = aws.Bool(true)
		}
	})

	client.storagegatewayconn.Handlers.Retry.PushBack(func(r *request.Request) {
		// InvalidGatewayRequestException: The specified gateway proxy network connection is busy.
		if isAWSErr(r.Error, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified gateway proxy network connection is busy") {
			r.Retryable = aws.Bool(true)
		}
	})

	client.wafv2conn.Handlers.Retry.PushBack(func(r *request.Request) {
		if isAWSErr(r.Error, wafv2.ErrCodeWAFInternalErrorException, "Retry your request") {
			r.Retryable = aws.Bool(true)
		}

		if r.Operation.Name == "CreateIPSet" || r.Operation.Name == "CreateRegexPatternSet" ||
			r.Operation.Name == "CreateRuleGroup" || r.Operation.Name == "CreateWebACL" {
			// WAFv2 supports tag on create which can result in the below error codes according to the documentation
			if isAWSErr(r.Error, wafv2.ErrCodeWAFTagOperationException, "Retry your request") {
				r.Retryable = aws.Bool(true)
			}
			if isAWSErr(err, wafv2.ErrCodeWAFTagOperationInternalErrorException, "Retry your request") {
				r.Retryable = aws.Bool(true)
			}
		}
	})

	if !c.SkipGetEC2Platforms {
		supportedPlatforms, err := GetSupportedEC2Platforms(client.ec2conn)
		if err != nil {
			// We intentionally fail *silently* because there's a chance
			// user just doesn't have ec2:DescribeAccountAttributes permissions
			log.Printf("[WARN] Unable to get supported EC2 platforms: %s", err)
		} else {
			client.supportedplatforms = supportedPlatforms
		}
	}

	return client, nil
}

func hasEc2Classic(platforms []string) bool {
	for _, p := range platforms {
		if p == "EC2" {
			return true
		}
	}
	return false
}

func GetSupportedEC2Platforms(conn *ec2.EC2) ([]string, error) {
	attrName := "supported-platforms"

	input := ec2.DescribeAccountAttributesInput{
		AttributeNames: []*string{aws.String(attrName)},
	}
	attributes, err := conn.DescribeAccountAttributes(&input)
	if err != nil {
		return nil, err
	}

	var platforms []string
	for _, attr := range attributes.AccountAttributes {
		if *attr.AttributeName == attrName {
			for _, v := range attr.AttributeValues {
				platforms = append(platforms, *v.AttributeValue)
			}
			break
		}
	}

	if len(platforms) == 0 {
		return nil, fmt.Errorf("No EC2 platforms detected")
	}

	return platforms, nil
}
