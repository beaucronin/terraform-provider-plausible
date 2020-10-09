package plausible

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	awsCredentials "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
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
	"github.com/hashicorp/go-cleanhttp"
	"github.com/mitchellh/go-homedir"
)

type Config struct {
	AppName                     string
	AccessKey                   string
	SecretKey                   string
	CredsFilename               string
	Profile                     string
	Token                       string
	Partition                   string
	AccountId                   string
	Region                      string
	MaxRetries                  int
	AssumeRoleARN               string
	AssumeRoleDurationSeconds   int
	AssumeRoleExternalID        string
	AssumeRolePolicy            string
	AssumeRolePolicyARNs        []string
	AssumeRoleSessionName       string
	AssumeRoleTags              map[string]string
	AssumeRoleTransitiveTagKeys []string
	AllowedAccountIds           []string
	ForbiddenAccountIds         []string
	Endpoints                   map[string]string
	Insecure                    bool
	SkipCredsValidation         bool
	SkipGetEC2Platforms         bool
	SkipRegionValidation        bool
	SkipRequestingAccountId     bool
	SkipMetadataApiCheck        bool
	S3ForcePathStyle            bool
	terraformVersion            string
	CallerDocumentationURL      string
	CallerName                  string
	DebugLogging                bool
	IamEndpoint                 string
	StsEndpoint                 string
}

type AWSClient struct {
	appname                   string
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

	sess, accountID, partition, err := awsbase.GetSessionWithAccountIDAndPartition(c)
	if err != nil {
		return nil, fmt.Errorf("error configuring Terraform AWS Provider: %w", err)
	}

	dnsSuffix := "amazonaws.com"
	if p, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), c.Region); ok {
		dnsSuffix = p.DNSSuffix()
	}

	client := &AWSClient{
		appname:                c.AppName,
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

	// Services that require multiple client configurations
	s3Config := &aws.Config{
		Endpoint:         aws.String(c.Endpoints["s3"]),
		S3ForcePathStyle: aws.Bool(c.S3ForcePathStyle),
	}
	client.s3conn = s3.New(sess.Copy(s3Config))
	s3Config.DisableRestProtocolURICleaning = aws.Bool(true)
	client.s3connUriCleaningDisabled = s3.New(sess.Copy(s3Config))

	return client, nil
}

// GetSession attempts to return valid AWS Go SDK session.
func GetSession(c *Config) (*session.Session, error) {
	options, err := GetSessionOptions(c)

	if err != nil {
		return nil, err
	}

	sess, err := session.NewSessionWithOptions(*options)
	if err != nil {
		return nil, fmt.Errorf("Error creating AWS session: %w", err)
	}

	if c.MaxRetries > 0 {
		sess = sess.Copy(&aws.Config{MaxRetries: aws.Int(c.MaxRetries)})
	}

	if !c.SkipCredsValidation {
		if _, _, err := GetAccountIDAndPartitionFromSTSGetCallerIdentity(sts.New(sess)); err != nil {
			return nil, fmt.Errorf("error validating provider credentials: %w", err)
		}
	}

	return sess, nil
}

func GetSessionOptions(c *Config) (*session.Options, error) {
	options := &session.Options{
		Config: aws.Config{
			EndpointResolver: c.EndpointResolver(),
			HTTPClient:       cleanhttp.DefaultClient(),
			MaxRetries:       aws.Int(0),
			Region:           aws.String(c.Region),
		},
		Profile:           c.Profile,
		SharedConfigState: session.SharedConfigEnable,
	}

	// get and validate credentials
	creds, err := GetCredentials(c)
	if err != nil {
		return nil, err
	}

	// add the validated credentials to the session options
	options.Config.Credentials = creds

	if c.Insecure {
		transport := options.Config.HTTPClient.Transport.(*http.Transport)
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	return options, nil
}

func GetCredentials(c *Config) (*awsCredentials.Credentials, error) {
	sharedCredentialsFilename, err := homedir.Expand(c.CredsFilename)

	if err != nil {
		return nil, fmt.Errorf("error expanding shared credentials filename: %w", err)
	}

	// build a chain provider, lazy-evaluated by aws-sdk
	providers := []awsCredentials.Provider{
		&awsCredentials.StaticProvider{Value: awsCredentials.Value{
			AccessKeyID:     c.AccessKey,
			SecretAccessKey: c.SecretKey,
			SessionToken:    c.Token,
		}},
		&awsCredentials.EnvProvider{},
		&awsCredentials.SharedCredentialsProvider{
			Filename: sharedCredentialsFilename,
			Profile:  c.Profile,
		},
	}

	// Validate the credentials before returning them
	creds := awsCredentials.NewChainCredentials(providers)
	cp, err := creds.Get()
	if err != nil {
		return nil, fmt.Errorf("Error loading credentials for AWS Provider: %w", err)
	} else {
		log.Printf("[INFO] AWS Auth provider used: %q", cp.ProviderName)
	}

	// This is the "normal" flow (i.e. not assuming a role)
	if c.AssumeRoleARN == "" {
		return creds, nil
	}

	// Otherwise we need to construct an STS client with the main credentials, and verify
	// that we can assume the defined role.
	log.Printf("[INFO] Attempting to AssumeRole %s (SessionName: %q, ExternalId: %q)",
		c.AssumeRoleARN, c.AssumeRoleSessionName, c.AssumeRoleExternalID)

	awsConfig := &aws.Config{
		Credentials:      creds,
		EndpointResolver: c.EndpointResolver(),
		Region:           aws.String(c.Region),
		MaxRetries:       aws.Int(c.MaxRetries),
		HTTPClient:       cleanhttp.DefaultClient(),
	}

	assumeRoleSession, err := session.NewSession(awsConfig)

	if err != nil {
		return nil, fmt.Errorf("error creating assume role session: %w", err)
	}

	stsclient := sts.New(assumeRoleSession)
	assumeRoleProvider := &stscreds.AssumeRoleProvider{
		Client:  stsclient,
		RoleARN: c.AssumeRoleARN,
	}

	if c.AssumeRoleDurationSeconds > 0 {
		assumeRoleProvider.Duration = time.Duration(c.AssumeRoleDurationSeconds) * time.Second
	}

	if c.AssumeRoleExternalID != "" {
		assumeRoleProvider.ExternalID = aws.String(c.AssumeRoleExternalID)
	}

	if c.AssumeRolePolicy != "" {
		assumeRoleProvider.Policy = aws.String(c.AssumeRolePolicy)
	}

	if len(c.AssumeRolePolicyARNs) > 0 {
		var policyDescriptorTypes []*sts.PolicyDescriptorType

		for _, policyARN := range c.AssumeRolePolicyARNs {
			policyDescriptorType := &sts.PolicyDescriptorType{
				Arn: aws.String(policyARN),
			}
			policyDescriptorTypes = append(policyDescriptorTypes, policyDescriptorType)
		}

		assumeRoleProvider.PolicyArns = policyDescriptorTypes
	}

	if c.AssumeRoleSessionName != "" {
		assumeRoleProvider.RoleSessionName = c.AssumeRoleSessionName
	}

	if len(c.AssumeRoleTags) > 0 {
		var tags []*sts.Tag

		for k, v := range c.AssumeRoleTags {
			tag := &sts.Tag{
				Key:   aws.String(k),
				Value: aws.String(v),
			}
			tags = append(tags, tag)
		}

		assumeRoleProvider.Tags = tags
	}

	if len(c.AssumeRoleTransitiveTagKeys) > 0 {
		assumeRoleProvider.TransitiveTagKeys = aws.StringSlice(c.AssumeRoleTransitiveTagKeys)
	}

	providers = []awsCredentials.Provider{assumeRoleProvider}

	assumeRoleCreds := awsCredentials.NewChainCredentials(providers)
	_, err = assumeRoleCreds.Get()

	return assumeRoleCreds, nil
}
