package plausible

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigateway"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mitchellh/go-homedir"
)

func resourceFunction() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFunctionCreate,
		ReadContext:   resourceFunctionRead,
		UpdateContext: resourceFunctionUpdate,
		DeleteContext: resourceFunctionDelete,
		Schema: map[string]*schema.Schema{
			"source": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"function_name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_code_hash": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"last_updated": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"account_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"handler": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "function.handler",
			},
			"memory_size": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  128,
			},
			"runtime": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "python3.7",
			},
			"timeout": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  10,
			},
			"publish": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"api_route_trigger": &schema.Schema{
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_id": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"route": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"method": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"get", "post", "put", "delete", "patch",
							}, true),
						},
						"content_type": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default:  "application/json",
						},
					},
				},
			},

			"schedule_trigger": &schema.Schema{
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cron": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"schedule_trigger_enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"schedule_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"schedule_name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"schedule_target_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"subscription_trigger": &schema.Schema{
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"publisher_id": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"subscription_trigger_enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"subscription_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"datastore_trigger": &schema.Schema{
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"datastore_id": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"datastore_trigger_enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},

			"environment": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"variables": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func resourceFunctionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	conn := m.(*AWSClient).lambdaconn
	accountId := m.(*AWSClient).accountid
	d.Set("account_id", accountId)

	var functionName string
	if v, ok := d.GetOk("function_name"); ok {
		functionName = v.(string)
	} else {
		functionName = resource.UniqueId()
	}
	source, _ := d.GetOk("source")
	zipFilename, err := doTheZip(source.(string))
	zipFilename = fmt.Sprintf("%s/lambda.zip", source)

	var functionCode *lambda.FunctionCode
	file, err := loadFileContent(zipFilename)
	if err != nil {
		return diag.Errorf("Unable to load %q: %s", zipFilename, err)
	}
	functionCode = &lambda.FunctionCode{
		ZipFile: file,
	}

	roleName := fmt.Sprintf("arn:aws:iam::%s:role/PlausibleLambdaRole", accountId)
	params := &lambda.CreateFunctionInput{
		Code:         functionCode,
		FunctionName: aws.String(functionName),
		Handler:      aws.String(d.Get("handler").(string)),
		MemorySize:   aws.Int64(int64(d.Get("memory_size").(int))),
		Runtime:      aws.String(d.Get("runtime").(string)),
		Timeout:      aws.Int64(int64(d.Get("timeout").(int))),
		Publish:      aws.Bool(d.Get("publish").(bool)),
		Role:         aws.String(roleName),
	}

	lambdaOut, err := conn.CreateFunction(params)
	if err != nil {
		return diag.Errorf("Error creating function: %s", err)
	}

	functionArn := lambdaOut.FunctionArn
	d.SetId(*functionArn)
	d.Set("arn", *functionArn)
	d.Set("function_name", functionName)

	if v, ok := d.GetOk("schedule_trigger"); ok {
		// A schedule trigger requires a CloudWatch rule that contains the schedule,
		// a Lambda permission that allows this rule to invoke the Lambda, and
		// an Event Target that tells the rule to invoke the Lambda

		// Create CloudWatch event rule
		cwconn := m.(*AWSClient).cloudwatcheventsconn
		ruleName := resource.UniqueId()
		ruleInput := events.PutRuleInput{
			Name: aws.String(ruleName),
		}
		if c, ok := v.([]*schema.ResourceData)[0].GetOk("cron"); ok {
			ruleInput.ScheduleExpression = aws.String(c.(string))
		}

		ruleInput.State = aws.String("true")
		ruleOut, err := cwconn.PutRule(&ruleInput)
		d.Set("schedule_trigger_enabled", true)
		d.Set("schedule_id", ruleOut.RuleArn)
		d.Set("schedule_name", ruleName)

		// Create Lambda permission
		input := lambda.AddPermissionInput{
			Action:       aws.String("lambda:InvokeFunction"),
			FunctionName: aws.String(functionName),
			Principal:    aws.String("events.amazonaws.com"),
			StatementId:  aws.String(resource.UniqueId()),
			SourceArn:    aws.String(d.Get("schedule_arn").(string)),
		}

		_, err = conn.AddPermission(&input)
		if err != nil {
			return diag.Errorf("Error adding lambda permission %+v", err)
		}

		// Create Cloudwatch event target
		targetId := resource.UniqueId()
		d.Set("schedule_target_id", targetId)
		targetInput := &events.PutTargetsInput{
			Rule: aws.String(ruleName),
			Targets: []*events.Target{
				&events.Target{
					Arn: aws.String(d.Id()),
				},
			},
		}
		_, err = cwconn.PutTargets(targetInput)
		if err != nil {
			return diag.Errorf("Creating CloudWatch Event Target failed: %s", err)
		}
	} else {
		d.Set("schedule_trigger_enabled", false)
	}

	if v, ok := d.GetOk("api_route_trigger"); ok {
		apiconn := m.(*AWSClient).apigatewayconn
		routeTriggerInfo := v.([]interface{})[0].(map[string]interface{}) //v.([]*schema.ResourceData)[0]
		var api_id = routeTriggerInfo["api_id"].(string)
		var method = routeTriggerInfo["method"].(string)
		var route = routeTriggerInfo["route"].(string)
		region := m.(*AWSClient).region

		// Create Lambda permission
		sourceArn := fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/*", region, accountId, api_id)
		input := lambda.AddPermissionInput{
			Action:       aws.String("lambda:InvokeFunction"),
			FunctionName: aws.String(functionName),
			Principal:    aws.String("apigateway.amazonaws.com"),
			StatementId:  aws.String(resource.UniqueId()),
			SourceArn:    aws.String(sourceArn),
		}

		_, err = conn.AddPermission(&input)
		if err != nil {
			return diag.Errorf("Error adding lambda permission %+v", err)
		}

		// Create API method integration
		r, err := apiconn.GetResources(&apigateway.GetResourcesInput{
			Limit:     aws.Int64(int64(500)),
			RestApiId: aws.String(api_id),
		})
		if err != nil {
			return diag.Errorf("Error getting resources for api %+v", err)
		}
		var resourceId *string
		for _, r := range r.Items {
			// rsc := r.(*apigateway.Resource)
			if *r.Path == route {
				resourceId = r.Id
				break
			}
		}
		if resourceId == nil {
			return diag.Errorf("No resource found in api for route %s", route)
		}

		// https://docs.aws.amazon.com/apigateway/latest/developerguide/set-up-lambda-custom-integrations.html
		// https://docs.aws.amazon.com/apigateway/api-reference/link-relation/integration-put/
		uri := fmt.Sprintf("arn:aws:apigateway:%s:lambda:path/2015-03-31/functions/arn:aws:lambda:%s:%s:function:%s/invocations", region, region, accountId, functionName, route)
		_, err = apiconn.PutIntegration(&apigateway.PutIntegrationInput{
			HttpMethod:            aws.String(method),
			ResourceId:            aws.String(*resourceId),
			RestApiId:             aws.String(api_id),
			Type:                  aws.String("AWS"),
			IntegrationHttpMethod: aws.String("POST"),
			Uri:                   aws.String(uri),
			Credentials:           aws.String(fmt.Sprintf("arn:aws:iam::%s:role/PlausibleLambdaRole", accountId)),
		})
		if err != nil {
			return diag.Errorf("Error creating API Gateway Integration: %+v", err)
		}

		// d.SetId(fmt.Sprintf("agi-%s-%s-%s", d.Get("rest_api_id").(string), d.Get("resource_id").(string), d.Get("http_method").(string)))

	}

	if v, ok := d.GetOk("subscription_trigger"); ok {
		// Create topic subscription
		snsconn := m.(*AWSClient).snsconn

		req := &sns.SubscribeInput{
			Protocol: aws.String("lambda"),
			Endpoint: aws.String(d.Id()),
			TopicArn: aws.String(v.([]*schema.ResourceData)[0].Get("publisher_id").(string)),
		}
		output, err := snsconn.Subscribe(req)
		if err != nil {
			return diag.Errorf("Creating SNS subscription failed: %s", err)
		}

		d.Set("subscription_trigger_enabled", true)
		d.Set("subscription_id", output.SubscriptionArn)
	} else {
		d.Set("subscription_trigger_enabled", false)
	}

	if v, ok := d.GetOk("datastore_trigger"); ok {
		// Resource creation depends on the type of datastore - key/value, object
		datastoreId := v.([]*schema.ResourceData)[0].Get("publisher_id").(string)
		datastoreArn, err := arn.Parse(datastoreId)
		if strings.Contains(strings.ToLower(datastoreId), ":dynamodb") {
			// Enable dynamodb stream
			ddbconn := m.(*AWSClient).dynamodbconn
			tableName := strings.Split(datastoreArn.Resource, "/")[1]
			input := &dynamodb.UpdateTableInput{
				TableName: aws.String(tableName),
				StreamSpecification: &dynamodb.StreamSpecification{
					StreamEnabled:  aws.Bool(true),
					StreamViewType: aws.String(dynamodb.StreamViewTypeNewImage),
				},
			}
			updateOutput, _ := ddbconn.UpdateTable(input)

			// Create Lambda event source mapping
			params := &lambda.CreateEventSourceMappingInput{
				EventSourceArn: updateOutput.TableDescription.LatestStreamArn,
				FunctionName:   aws.String(functionName),
				Enabled:        aws.Bool(true),
			}
			// eventSourceMappingConfiguration, err := conn.CreateEventSourceMapping(params)
			_, _ = conn.CreateEventSourceMapping(params)
		} else if strings.Contains(strings.ToLower(datastoreId), ":s3") {
			// Configure S3 object-level events
			// s3conn := m.(*AWSClient).s3conn

			// Create Lambda permission
			input := lambda.AddPermissionInput{
				Action:       aws.String("lambda:InvokeFunction"),
				FunctionName: aws.String(functionName),
				Principal:    aws.String("apigateway.amazonaws.com"),
				StatementId:  aws.String(resource.UniqueId()),
				SourceArn:    aws.String("arn:aws:execute-api:*:*:*"),
			}

			_, err = conn.AddPermission(&input)
			if err != nil {
				return diag.Errorf("Error adding lambda permission %s", err)
			}

			notificationConfiguration := &s3.NotificationConfiguration{}
			lc := &s3.LambdaFunctionConfiguration{}
			lc.Id = aws.String(resource.UniqueId())
			lc.LambdaFunctionArn = aws.String(*lambdaOut.FunctionArn)
			lc.Events = []*string{aws.String("s3:ObjectCreated:*")}
			notificationConfiguration.LambdaFunctionConfigurations = []*s3.LambdaFunctionConfiguration{lc}
			var bucket = "my ucket"
			_ = &s3.PutBucketNotificationConfigurationInput{
				Bucket:                    aws.String(bucket),
				NotificationConfiguration: notificationConfiguration,
			}
		}
		d.Set("datastore_trigger_enabled", true)
	} else {
		d.Set("datastore_trigger_enabled", false)
	}

	return nil
}

func resourceFunctionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := m.(*AWSClient).lambdaconn

	params := &lambda.GetFunctionInput{
		FunctionName: aws.String(d.Get("function_name").(string)),
	}
	getFunctionOutput, err := conn.GetFunction(params)
	if err != nil {
		return diag.Errorf("Error getting FunctionOutput %s", err)
	}
	function := getFunctionOutput.Configuration
	d.Set("arn", function.FunctionArn)
	d.Set("function_name", function.FunctionName)
	d.Set("handler", function.Handler)
	d.Set("memory_size", function.MemorySize)
	d.Set("last_modified", function.LastModified)
	d.Set("role", function.Role)
	d.Set("runtime", function.Runtime)
	d.Set("timeout", function.Timeout)
	// d.Set("kms_key_arn", function.KMSKeyArn)
	d.Set("source_code_hash", function.CodeSha256)
	d.Set("source_code_size", function.CodeSize)

	// invokeArn := lambdaFunctionInvokeArn(*function.FunctionArn, meta)
	// d.Set("invoke_arn", invokeArn)

	return diags
}

func resourceFunctionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func resourceFunctionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := m.(*AWSClient).lambdaconn
	cwconn := m.(*AWSClient).cloudwatcheventsconn

	if v := d.Get("schedule_trigger_enabled").(bool); v {
		// Delete CloudWatch event target
		targetInput := &events.RemoveTargetsInput{
			Ids:  []*string{aws.String(d.Get("schedule_target_id").(string))},
			Rule: aws.String(d.Get("schedule_name").(string)),
		}
		_, err := cwconn.RemoveTargets(targetInput)
		if err != nil {
			return diag.Errorf("Error removing scheduling target: %s", err)
		}

		// Delete Lambda permission
		permInput := lambda.RemovePermissionInput{
			FunctionName: aws.String(d.Get("function_name").(string)),
			StatementId:  aws.String(d.Id()),
		}
		_, err = conn.RemovePermission(&permInput)
		if err != nil {
			return diag.Errorf("Error removing Lambda permission: %s", err)
		}

		// Delete CloudWatch event rule
		ruleInput := &events.DeleteRuleInput{
			Name: aws.String(d.Id()),
		}
		_, err = cwconn.DeleteRule(ruleInput)
		if err != nil {
			return diag.Errorf("Error removing scheduling rule: %s", err)
		}
	}

	if _, ok := d.GetOk("subscription_trigger_enabled"); ok {
		// Delete topic subscription
	}

	if _, ok := d.GetOk("datastore_trigger_enabled"); ok {
		// delete
	}

	if _, ok := d.GetOk("api_route_trigger_enabled"); ok {
	}

	// Delete the lambda function
	return diags
}

func loadFileContent(v string) ([]byte, error) {
	filename, err := homedir.Expand(v)
	if err != nil {
		return nil, err
	}
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return fileContent, nil
}

func doTheZip(s string) (string, error) {
	return "", nil
}
