package plausible

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	awsCredentials "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/mitchellh/go-homedir"
)

type AWSConfig struct {
	AppName          string
	AccessKey        string
	SecretKey        string
	CredsFilename    string
	Profile          string
	Token            string
	Partition        string
	AccountId        string
	Region           string
	terraformVersion string
	CallerName       string
}

type AWSClient struct {
	apigatewayconn       *apigateway.APIGateway
	cloudwatcheventsconn *cloudwatchevents.CloudWatchEvents
	dynamodbconn         *dynamodb.DynamoDB
	firehoseconn         *firehose.Firehose
	kinesisanalyticsconn *kinesisanalytics.KinesisAnalytics
	kinesisconn          *kinesis.Kinesis
	lambdaconn           *lambda.Lambda
	s3conn               *s3.S3
	snsconn              *sns.SNS
	sqsconn              *sqs.SQS
	AppName              string
}

func (conf *AWSConfig) Client() (interface{}, error) {
	sess, _ := GetSession(conf)
	client := &AWSClient{
		apigatewayconn:       apigateway.New(sess.Copy()),
		cloudwatcheventsconn: cloudwatchevents.New(sess.Copy()),
		dynamodbconn:         dynamodb.New(sess.Copy()),
		firehoseconn:         firehose.New(sess.Copy()),
		kinesisanalyticsconn: kinesisanalytics.New(sess.Copy()),
		kinesisconn:          kinesis.New(sess.Copy()),
		lambdaconn:           lambda.New(sess.Copy()),
		s3conn:               s3.New(sess.Copy()),
		snsconn:              sns.New(sess.Copy()),
		sqsconn:              sqs.New(sess.Copy()),
		AppName:              conf.AppName,
	}

	return client, nil
}

func GetSession(conf *AWSConfig) (*session.Session, error) {
	creds, _ := GetCredentials(conf)
	options := &session.Options{
		Config: aws.Config{
			Credentials: creds,
			Region:      aws.String(conf.Region),
		},
		Profile: conf.Profile,
	}
	sess, err := session.NewSession()
	return sess, err
}

func GetCredentials(c *AWSConfig) (*awsCredentials.Credentials, error) {
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
			Profile: c.Profile,
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
	return creds, nil
}
