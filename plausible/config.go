package plausible

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/cloudsearch"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/aws/aws-sdk-go/service/datapipeline"
	"github.com/aws/aws-sdk-go/service/dax"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/forecastservice"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/aws/aws-sdk-go/service/kinesisanalyticsv2"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53domains"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/xray"
	awsbase "github.com/hashicorp/aws-sdk-go-base"
)

type Config struct {
	AccessKey     string
	SecretKey     string
	CredsFilename string
	Profile       string
	Token         string
	Partition     string
	AccountId     string
	Region        string
	MaxRetries    int

	AssumeRoleARN               string
	AssumeRoleDurationSeconds   int
	AssumeRoleExternalID        string
	AssumeRolePolicy            string
	AssumeRolePolicyARNs        []string
	AssumeRoleSessionName       string
	AssumeRoleTags              map[string]string
	AssumeRoleTransitiveTagKeys []string

	AllowedAccountIds   []string
	ForbiddenAccountIds []string

	Endpoints map[string]string
	Insecure  bool

	SkipCredsValidation     bool
	SkipGetEC2Platforms     bool
	SkipRegionValidation    bool
	SkipRequestingAccountId bool
	SkipMetadataApiCheck    bool
	S3ForcePathStyle        bool

	terraformVersion string
}

type AWSClient struct {
	apigatewayconn            *apigateway.APIGateway
	apigatewayv2conn          *apigatewayv2.ApiGatewayV2
	athenaconn                *athena.Athena
	batchconn                 *batch.Batch
	cloudfrontconn            *cloudfront.CloudFront
	cloudsearchconn           *cloudsearch.CloudSearch
	cloudtrailconn            *cloudtrail.CloudTrail
	cloudwatchconn            *cloudwatch.CloudWatch
	cloudwatcheventsconn      *cloudwatchevents.CloudWatchEvents
	cloudwatchlogsconn        *cloudwatchlogs.CloudWatchLogs
	cognitoconn               *cognitoidentity.CognitoIdentity
	cognitoidpconn            *cognitoidentityprovider.CognitoIdentityProvider
	configconn                *configservice.ConfigService
	datapipelineconn          *datapipeline.DataPipeline
	daxconn                   *dax.DAX
	dnsSuffix                 string
	dynamodbconn              *dynamodb.DynamoDB
	efsconn                   *efs.EFS
	firehoseconn              *firehose.Firehose
	forecastconn              *forecastservice.ForecastService
	glueconn                  *glue.Glue
	iamconn                   *iam.IAM
	kinesisanalyticsconn      *kinesisanalytics.KinesisAnalytics
	kinesisanalyticsv2conn    *kinesisanalyticsv2.KinesisAnalyticsV2
	kinesisconn               *kinesis.Kinesis
	kmsconn                   *kms.KMS
	lambdaconn                *lambda.Lambda
	lexmodelconn              *lexmodelbuildingservice.LexModelBuildingService
	partition                 string
	quicksightconn            *quicksight.QuickSight
	r53conn                   *route53.Route53
	rdsconn                   *rds.RDS
	redshiftconn              *redshift.Redshift
	region                    string
	accountid                 string
	route53domainsconn        *route53domains.Route53Domains
	route53resolverconn       *route53resolver.Route53Resolver
	s3conn                    *s3.S3
	s3connUriCleaningDisabled *s3.S3
	s3controlconn             *s3control.S3Control
	sagemakerconn             *sagemaker.SageMaker
	secretsmanagerconn        *secretsmanager.SecretsManager
	sesconn                   *ses.SES
	sfnconn                   *sfn.SFN
	simpledbconn              *simpledb.SimpleDB
	snsconn                   *sns.SNS
	sqsconn                   *sqs.SQS
	stsconn                   *sts.STS
	xrayconn                  *xray.XRay
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

	awsbaseConfig := &awsbase.Config{
		AccessKey:                   c.AccessKey,
		AssumeRoleARN:               c.AssumeRoleARN,
		AssumeRoleDurationSeconds:   c.AssumeRoleDurationSeconds,
		AssumeRoleExternalID:        c.AssumeRoleExternalID,
		AssumeRolePolicy:            c.AssumeRolePolicy,
		AssumeRolePolicyARNs:        c.AssumeRolePolicyARNs,
		AssumeRoleSessionName:       c.AssumeRoleSessionName,
		AssumeRoleTags:              c.AssumeRoleTags,
		AssumeRoleTransitiveTagKeys: c.AssumeRoleTransitiveTagKeys,
		CallerName:                  "Plausible|AWS Provider",
		CredsFilename:               c.CredsFilename,
		DebugLogging:                true, //logging.IsDebugOrHigher(),
		IamEndpoint:                 c.Endpoints["iam"],
		Insecure:                    c.Insecure,
		MaxRetries:                  c.MaxRetries,
		// Partition:                   c.Partition,
		Profile:                 c.Profile,
		Region:                  c.Region,
		SecretKey:               c.SecretKey,
		SkipCredsValidation:     c.SkipCredsValidation,
		SkipMetadataApiCheck:    c.SkipMetadataApiCheck,
		SkipRequestingAccountId: c.SkipRequestingAccountId,
		StsEndpoint:             c.Endpoints["sts"],
		Token:                   c.Token,
		// UserAgentProducts: []*awsbase.UserAgentProduct{
		// 	{Name: "APN", Version: "1.0"},
		// 	{Name: "HashiCorp", Version: "1.0"},
		// 	{Name: "Terraform", Version: c.terraformVersion,
		// 		Extra: []string{"+https://www.terraform.io"}},
		// },
	}

	sess, accountID, partition, err := awsbase.GetSessionWithAccountIDAndPartition(awsbaseConfig)
	if err != nil {
		return nil, fmt.Errorf("error configuring Terraform AWS Provider: %w", err)
	}

	fmt.Errorf("Account ID %s", accountID)

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
		apigatewayconn:         apigateway.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["apigateway"])})),
		apigatewayv2conn:       apigatewayv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["apigateway"])})),
		athenaconn:             athena.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["athena"])})),
		cloudfrontconn:         cloudfront.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cloudfront"])})),
		cloudwatchconn:         cloudwatch.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cloudwatch"])})),
		cloudwatcheventsconn:   cloudwatchevents.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cloudwatchevents"])})),
		cloudwatchlogsconn:     cloudwatchlogs.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cloudwatchlogs"])})),
		daxconn:                dax.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["dax"])})),
		dnsSuffix:              dnsSuffix,
		dynamodbconn:           dynamodb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["dynamodb"])})),
		firehoseconn:           firehose.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["firehose"])})),
		forecastconn:           forecastservice.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["forecast"])})),
		iamconn:                iam.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["iam"])})),
		kinesisanalyticsconn:   kinesisanalytics.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["kinesisanalytics"])})),
		kinesisanalyticsv2conn: kinesisanalyticsv2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["kinesisanalyticsv2"])})),
		kinesisconn:            kinesis.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["kinesis"])})),
		kmsconn:                kms.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["kms"])})),
		lambdaconn:             lambda.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["lambda"])})),
		partition:              c.Partition,
		rdsconn:                rds.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["rds"])})),
		region:                 c.Region,
		accountid:              accountID,
		route53resolverconn:    route53resolver.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["route53resolver"])})),
		secretsmanagerconn:     secretsmanager.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["secretsmanager"])})),
		sesconn:                ses.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["ses"])})),
		sfnconn:                sfn.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["stepfunctions"])})),
		simpledbconn:           simpledb.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["sdb"])})),
		snsconn:                sns.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["sns"])})),
		sqsconn:                sqs.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["sqs"])})),
		stsconn:                sts.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["sts"])})),
		xrayconn:               xray.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["xray"])})),
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

	client.r53conn = route53.New(sess.Copy(route53Config))

	return client, nil
}
