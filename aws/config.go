package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/autoscaling"
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
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/aws/aws-sdk-go/service/codedeploy"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/aws/aws-sdk-go/service/costandusagereportservice"
	"github.com/aws/aws-sdk-go/service/databasemigrationservice"
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
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/aws/aws-sdk-go/service/glacier"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/inspector"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/aws/aws-sdk-go/service/kinesisanalyticsv2"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/aws/aws-sdk-go/service/licensemanager"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/aws/aws-sdk-go/service/macie"
	"github.com/aws/aws-sdk-go/service/mediaconnect"
	"github.com/aws/aws-sdk-go/service/mediaconvert"
	"github.com/aws/aws-sdk-go/service/medialive"
	"github.com/aws/aws-sdk-go/service/mediapackage"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/aws/aws-sdk-go/service/mediastoredata"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/aws/aws-sdk-go/service/opsworkscm"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/aws/aws-sdk-go/service/resourcegroups"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/aws/aws-sdk-go/service/serverlessapplicationrepository"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
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
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/aws/aws-sdk-go/service/worklink"
	"github.com/aws/aws-sdk-go/service/workspaces"
	awsbase "github.com/hashicorp/aws-sdk-go-base"
	"github.com/hashicorp/terraform/helper/logging"
	"github.com/hashicorp/terraform/terraform"
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

	AcmEndpoint              string
	ApigatewayEndpoint       string
	CloudFormationEndpoint   string
	CloudWatchEndpoint       string
	CloudWatchEventsEndpoint string
	CloudWatchLogsEndpoint   string
	DynamoDBEndpoint         string
	DeviceFarmEndpoint       string
	Ec2Endpoint              string
	EcsEndpoint              string
	AutoscalingEndpoint      string
	EcrEndpoint              string
	EfsEndpoint              string
	EsEndpoint               string
	ElbEndpoint              string
	FirehoseEndpoint         string
	IamEndpoint              string
	KinesisEndpoint          string
	KinesisAnalyticsEndpoint string
	KmsEndpoint              string
	LambdaEndpoint           string
	RdsEndpoint              string
	RedshiftEndpoint         string
	R53Endpoint              string
	S3Endpoint               string
	S3ControlEndpoint        string
	SesEndpoint              string
	SnsEndpoint              string
	SqsEndpoint              string
	StsEndpoint              string
	SsmEndpoint              string
	Insecure                 bool

	SkipCredsValidation     bool
	SkipGetEC2Platforms     bool
	SkipRegionValidation    bool
	SkipRequestingAccountId bool
	SkipMetadataApiCheck    bool
	S3ForcePathStyle        bool
}

type AWSClient struct {
	accountid                           string
	acmconn                             *acm.ACM
	acmpcaconn                          *acmpca.ACMPCA
	apigateway                          *apigateway.APIGateway
	apigatewayv2conn                    *apigatewayv2.ApiGatewayV2
	appautoscalingconn                  *applicationautoscaling.ApplicationAutoScaling
	appmeshconn                         *appmesh.AppMesh
	appsyncconn                         *appsync.AppSync
	athenaconn                          *athena.Athena
	autoscalingconn                     *autoscaling.AutoScaling
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
	codebuildconn                       *codebuild.CodeBuild
	codecommitconn                      *codecommit.CodeCommit
	codedeployconn                      *codedeploy.CodeDeploy
	codepipelineconn                    *codepipeline.CodePipeline
	cognitoconn                         *cognitoidentity.CognitoIdentity
	cognitoidpconn                      *cognitoidentityprovider.CognitoIdentityProvider
	configconn                          *configservice.ConfigService
	costandusagereportconn              *costandusagereportservice.CostandUsageReportService
	datapipelineconn                    *datapipeline.DataPipeline
	datasyncconn                        *datasync.DataSync
	daxconn                             *dax.DAX
	devicefarmconn                      *devicefarm.DeviceFarm
	dlmconn                             *dlm.DLM
	dmsconn                             *databasemigrationservice.DatabaseMigrationService
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
	fsxconn                             *fsx.FSx
	gameliftconn                        *gamelift.GameLift
	glacierconn                         *glacier.Glacier
	globalacceleratorconn               *globalaccelerator.GlobalAccelerator
	glueconn                            *glue.Glue
	guarddutyconn                       *guardduty.GuardDuty
	iamconn                             *iam.IAM
	inspectorconn                       *inspector.Inspector
	iotconn                             *iot.IoT
	kafkaconn                           *kafka.Kafka
	kinesisanalyticsconn                *kinesisanalytics.KinesisAnalytics
	kinesisanalyticsv2conn              *kinesisanalyticsv2.KinesisAnalyticsV2
	kinesisconn                         *kinesis.Kinesis
	kmsconn                             *kms.KMS
	lambdaconn                          *lambda.Lambda
	lexmodelconn                        *lexmodelbuildingservice.LexModelBuildingService
	licensemanagerconn                  *licensemanager.LicenseManager
	lightsailconn                       *lightsail.Lightsail
	macieconn                           *macie.Macie
	mediaconnectconn                    *mediaconnect.MediaConnect
	mediaconvertconn                    *mediaconvert.MediaConvert
	medialiveconn                       *medialive.MediaLive
	mediapackageconn                    *mediapackage.MediaPackage
	mediastoreconn                      *mediastore.MediaStore
	mediastoredataconn                  *mediastoredata.MediaStoreData
	mqconn                              *mq.MQ
	neptuneconn                         *neptune.Neptune
	opsworksconn                        *opsworks.OpsWorks
	opsworkscmconn                      *opsworkscm.OpsWorksCM
	organizationsconn                   *organizations.Organizations
	partition                           string
	pinpointconn                        *pinpoint.Pinpoint
	pricingconn                         *pricing.Pricing
	r53conn                             *route53.Route53
	ramconn                             *ram.RAM
	rdsconn                             *rds.RDS
	redshiftconn                        *redshift.Redshift
	region                              string
	resourcegroupsconn                  *resourcegroups.ResourceGroups
	route53resolverconn                 *route53resolver.Route53Resolver
	s3conn                              *s3.S3
	s3controlconn                       *s3control.S3Control
	sagemakerconn                       *sagemaker.SageMaker
	scconn                              *servicecatalog.ServiceCatalog
	sdconn                              *servicediscovery.ServiceDiscovery
	secretsmanagerconn                  *secretsmanager.SecretsManager
	securityhubconn                     *securityhub.SecurityHub
	serverlessapplicationrepositoryconn *serverlessapplicationrepository.ServerlessApplicationRepository
	sesConn                             *ses.SES
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
	transferconn                        *transfer.Transfer
	wafconn                             *waf.WAF
	wafregionalconn                     *wafregional.WAFRegional
	worklinkconn                        *worklink.WorkLink
	workspacesconn                      *workspaces.WorkSpaces
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
		IamEndpoint:             c.IamEndpoint,
		Insecure:                c.Insecure,
		MaxRetries:              c.MaxRetries,
		Profile:                 c.Profile,
		Region:                  c.Region,
		SecretKey:               c.SecretKey,
		SkipCredsValidation:     c.SkipCredsValidation,
		SkipMetadataApiCheck:    c.SkipMetadataApiCheck,
		SkipRequestingAccountId: c.SkipRequestingAccountId,
		StsEndpoint:             c.StsEndpoint,
		Token:                   c.Token,
		UserAgentProducts: []*awsbase.UserAgentProduct{
			{Name: "APN", Version: "1.0"},
			{Name: "HashiCorp", Version: "1.0"},
			{Name: "Terraform", Version: terraform.VersionString()},
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

	client := &AWSClient{
		accountid:                           accountID,
		acmconn:                             acm.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.AcmEndpoint)})),
		acmpcaconn:                          acmpca.New(sess),
		apigateway:                          apigateway.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.ApigatewayEndpoint)})),
		apigatewayv2conn:                    apigatewayv2.New(sess),
		appautoscalingconn:                  applicationautoscaling.New(sess),
		appmeshconn:                         appmesh.New(sess),
		appsyncconn:                         appsync.New(sess),
		athenaconn:                          athena.New(sess),
		autoscalingconn:                     autoscaling.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.AutoscalingEndpoint)})),
		backupconn:                          backup.New(sess),
		batchconn:                           batch.New(sess),
		budgetconn:                          budgets.New(sess),
		cfconn:                              cloudformation.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.CloudFormationEndpoint)})),
		cloud9conn:                          cloud9.New(sess),
		cloudfrontconn:                      cloudfront.New(sess),
		cloudhsmv2conn:                      cloudhsmv2.New(sess),
		cloudsearchconn:                     cloudsearch.New(sess),
		cloudtrailconn:                      cloudtrail.New(sess),
		cloudwatchconn:                      cloudwatch.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.CloudWatchEndpoint)})),
		cloudwatcheventsconn:                cloudwatchevents.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.CloudWatchEventsEndpoint)})),
		cloudwatchlogsconn:                  cloudwatchlogs.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.CloudWatchLogsEndpoint)})),
		codebuildconn:                       codebuild.New(sess),
		codecommitconn:                      codecommit.New(sess),
		codedeployconn:                      codedeploy.New(sess),
		codepipelineconn:                    codepipeline.New(sess),
		cognitoconn:                         cognitoidentity.New(sess),
		cognitoidpconn:                      cognitoidentityprovider.New(sess),
		configconn:                          configservice.New(sess),
		costandusagereportconn:              costandusagereportservice.New(sess),
		datapipelineconn:                    datapipeline.New(sess),
		datasyncconn:                        datasync.New(sess),
		daxconn:                             dax.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.DynamoDBEndpoint)})),
		devicefarmconn:                      devicefarm.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.DeviceFarmEndpoint)})),
		dlmconn:                             dlm.New(sess),
		dmsconn:                             databasemigrationservice.New(sess),
		docdbconn:                           docdb.New(sess),
		dsconn:                              directoryservice.New(sess),
		dxconn:                              directconnect.New(sess),
		dynamodbconn:                        dynamodb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.DynamoDBEndpoint)})),
		ec2conn:                             ec2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Ec2Endpoint)})),
		ecrconn:                             ecr.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.EcrEndpoint)})),
		ecsconn:                             ecs.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.EcsEndpoint)})),
		efsconn:                             efs.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.EfsEndpoint)})),
		eksconn:                             eks.New(sess),
		elasticacheconn:                     elasticache.New(sess),
		elasticbeanstalkconn:                elasticbeanstalk.New(sess),
		elastictranscoderconn:               elastictranscoder.New(sess),
		elbconn:                             elb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.ElbEndpoint)})),
		elbv2conn:                           elbv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.ElbEndpoint)})),
		emrconn:                             emr.New(sess),
		esconn:                              elasticsearch.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.EsEndpoint)})),
		firehoseconn:                        firehose.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.FirehoseEndpoint)})),
		fmsconn:                             fms.New(sess),
		fsxconn:                             fsx.New(sess),
		gameliftconn:                        gamelift.New(sess),
		glacierconn:                         glacier.New(sess),
		globalacceleratorconn:               globalaccelerator.New(sess),
		glueconn:                            glue.New(sess),
		guarddutyconn:                       guardduty.New(sess),
		iamconn:                             iam.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.IamEndpoint)})),
		inspectorconn:                       inspector.New(sess),
		iotconn:                             iot.New(sess),
		kafkaconn:                           kafka.New(sess),
		kinesisanalyticsconn:                kinesisanalytics.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.KinesisAnalyticsEndpoint)})),
		kinesisanalyticsv2conn:              kinesisanalyticsv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.KinesisAnalyticsEndpoint)})),
		kinesisconn:                         kinesis.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.KinesisEndpoint)})),
		kmsconn:                             kms.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.KmsEndpoint)})),
		lambdaconn:                          lambda.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.LambdaEndpoint)})),
		lexmodelconn:                        lexmodelbuildingservice.New(sess),
		licensemanagerconn:                  licensemanager.New(sess),
		lightsailconn:                       lightsail.New(sess),
		macieconn:                           macie.New(sess),
		mediaconnectconn:                    mediaconnect.New(sess),
		mediaconvertconn:                    mediaconvert.New(sess),
		medialiveconn:                       medialive.New(sess),
		mediapackageconn:                    mediapackage.New(sess),
		mediastoreconn:                      mediastore.New(sess),
		mediastoredataconn:                  mediastoredata.New(sess),
		mqconn:                              mq.New(sess),
		neptuneconn:                         neptune.New(sess),
		opsworksconn:                        opsworks.New(sess),
		opsworkscmconn:                      opsworkscm.New(sess),
		organizationsconn:                   organizations.New(sess),
		partition:                           partition,
		pinpointconn:                        pinpoint.New(sess),
		pricingconn:                         pricing.New(sess),
		r53conn:                             route53.New(sess.Copy(&aws.Config{Region: aws.String("us-east-1"), Endpoint: aws.String(c.R53Endpoint)})),
		ramconn:                             ram.New(sess),
		rdsconn:                             rds.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.RdsEndpoint)})),
		redshiftconn:                        redshift.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.RedshiftEndpoint)})),
		region:                              c.Region,
		resourcegroupsconn:                  resourcegroups.New(sess),
		route53resolverconn:                 route53resolver.New(sess),
		s3conn:                              s3.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.S3Endpoint), S3ForcePathStyle: aws.Bool(c.S3ForcePathStyle)})),
		s3controlconn:                       s3control.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.S3ControlEndpoint)})),
		sagemakerconn:                       sagemaker.New(sess),
		scconn:                              servicecatalog.New(sess),
		sdconn:                              servicediscovery.New(sess),
		secretsmanagerconn:                  secretsmanager.New(sess),
		securityhubconn:                     securityhub.New(sess),
		serverlessapplicationrepositoryconn: serverlessapplicationrepository.New(sess),
		sesConn:                             ses.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.SesEndpoint)})),
		sfnconn:                             sfn.New(sess),
		shieldconn:                          shield.New(sess),
		simpledbconn:                        simpledb.New(sess),
		snsconn:                             sns.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.SnsEndpoint)})),
		sqsconn:                             sqs.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.SqsEndpoint)})),
		ssmconn:                             ssm.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.SsmEndpoint)})),
		storagegatewayconn:                  storagegateway.New(sess),
		stsconn:                             sts.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.StsEndpoint)})),
		swfconn:                             swf.New(sess),
		transferconn:                        transfer.New(sess),
		wafconn:                             waf.New(sess),
		wafregionalconn:                     wafregional.New(sess),
		worklinkconn:                        worklink.New(sess),
		workspacesconn:                      workspaces.New(sess),
	}

	// Workaround for https://github.com/aws/aws-sdk-go/issues/1376
	client.kinesisconn.Handlers.Retry.PushBack(func(r *request.Request) {
		if !strings.HasPrefix(r.Operation.Name, "Describe") && !strings.HasPrefix(r.Operation.Name, "List") {
			return
		}
		err, ok := r.Error.(awserr.Error)
		if !ok || err == nil {
			return
		}
		if err.Code() == kinesis.ErrCodeLimitExceededException {
			r.Retryable = aws.Bool(true)
		}
	})

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

	// See https://github.com/aws/aws-sdk-go/pull/1276
	client.dynamodbconn.Handlers.Retry.PushBack(func(r *request.Request) {
		if r.Operation.Name != "PutItem" && r.Operation.Name != "UpdateItem" && r.Operation.Name != "DeleteItem" {
			return
		}
		if isAWSErr(r.Error, dynamodb.ErrCodeLimitExceededException, "Subscriber limit exceeded:") {
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

	client.storagegatewayconn.Handlers.Retry.PushBack(func(r *request.Request) {
		// InvalidGatewayRequestException: The specified gateway proxy network connection is busy.
		if isAWSErr(r.Error, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified gateway proxy network connection is busy") {
			r.Retryable = aws.Bool(true)
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
