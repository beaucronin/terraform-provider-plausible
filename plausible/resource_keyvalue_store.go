package plausible

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceKeyValueStore() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeyValueStoreCreate,
		ReadContext:   resourceKeyValueStoreRead,
		UpdateContext: resourceKeyValueStoreUpdate,
		DeleteContext: resourceKeyValueStoreDelete,
		Schema: map[string]*schema.Schema{
			"collection_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"primary_index": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"partition_key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"row_key": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"secondary_index": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"partition_key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"row_key": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceKeyValueStoreCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	conn := m.(*AWSClient).dynamodbconn

	// Create KeySchema for primary index
	piList := d.Get("primary_index").([]interface{})
	pi := piList[0].(map[string]interface{})

	keySchema := []*dynamodb.KeySchemaElement{}
	keySchema = append(keySchema, &dynamodb.KeySchemaElement{
		AttributeName: aws.String(pi["partition_key"].(string)),
		KeyType:       aws.String(dynamodb.KeyTypeHash),
	})

	if v, ok := pi["row_key"]; ok && v != nil && v != "" {
		keySchema = append(keySchema, &dynamodb.KeySchemaElement{
			AttributeName: aws.String(pi["row_key"].(string)),
			KeyType:       aws.String(dynamodb.KeyTypeRange),
		})
	}

	req := &dynamodb.CreateTableInput{
		TableName:   aws.String(d.Get("collection_name").(string)),
		BillingMode: aws.String("PAY_PER_REQUEST"),
		KeySchema:   keySchema,
	}

	if v, ok := d.GetOk("secondary_index"); ok {
		secondaryIndexes := []*dynamodb.GlobalSecondaryIndex{}
		gsiSet := v.(*schema.Set)

		for _, gsiObject := range gsiSet.List() {
			gsi := gsiObject.(map[string]interface{})
			keySchema := []*dynamodb.KeySchemaElement{}
			keySchema = append(keySchema, &dynamodb.KeySchemaElement{
				AttributeName: aws.String(gsi["partition_key"].(string)),
				KeyType:       aws.String(dynamodb.KeyTypeHash),
			})

			if v, ok := gsi["row_key"]; ok && v != nil && v != "" {
				keySchema = append(keySchema, &dynamodb.KeySchemaElement{
					AttributeName: aws.String(pi["row_key"].(string)),
					KeyType:       aws.String(dynamodb.KeyTypeRange),
				})
			}

			gsiDescription := &dynamodb.GlobalSecondaryIndex{
				IndexName:             aws.String(gsi["name"].(string)),
				KeySchema:             keySchema,
				Projection:            &dynamodb.Projection{ProjectionType: aws.String("ALL")},
				ProvisionedThroughput: nil,
			}
			secondaryIndexes = append(secondaryIndexes, gsiDescription)
		}
		req.GlobalSecondaryIndexes = secondaryIndexes
	}

	output, err := conn.CreateTable(req)

	if err != nil {
		return diag.Errorf("error creating DynamoDB Table: %s", err)
	}

	d.SetId(aws.StringValue(output.TableDescription.TableArn))

	return resourceKeyValueStoreRead(ctx, d, m)
}

func resourceKeyValueStoreRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	conn := m.(*AWSClient).dynamodbconn
	result, err := conn.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(d.Id()),
	})

	if err != nil {
		return diag.Errorf("Error getting DynamoDB Table")
	}

	table := result.Table

	// Need to record the pk as a one-member list, because that is how the resource is specified in TF
	piList := make([]map[string]interface{}, 0, 1)
	pi := map[string]interface{}{}
	for _, attr := range table.KeySchema {
		if *attr.KeyType == dynamodb.KeyTypeHash {
			pi["partition_key"] = *attr.AttributeName
		}

		if *attr.KeyType == dynamodb.KeyTypeRange {
			pi["row_key"] = *attr.AttributeName
		}
	}
	piList = append(piList, pi)

	// Collect the GSI descriptions and create a List of them for the state
	gsiList := make([]map[string]interface{}, 0, len(table.GlobalSecondaryIndexes))
	for _, gsiObject := range table.GlobalSecondaryIndexes {
		gsi := map[string]interface{}{
			"name": *gsiObject.IndexName,
		}

		for _, attribute := range gsiObject.KeySchema {
			if *attribute.KeyType == dynamodb.KeyTypeHash {
				gsi["partition_key"] = *attribute.AttributeName
			}

			if *attribute.KeyType == dynamodb.KeyTypeRange {
				gsi["row_key"] = *attribute.AttributeName
			}
		}
		gsiList = append(gsiList, gsi)
	}
	err = d.Set("secondary_index", gsiList)
	if err != nil {
		return diag.Errorf("Error setting secondary indexes %s", err)
	}

	return nil
}

func resourceKeyValueStoreUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceKeyValueStoreDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func expandDynamoDbKeySchema(data map[string]interface{}) []*dynamodb.KeySchemaElement {
	keySchema := []*dynamodb.KeySchemaElement{}

	if v, ok := data["partition_key"]; ok && v != nil && v != "" {
		keySchema = append(keySchema, &dynamodb.KeySchemaElement{
			AttributeName: aws.String(v.(string)),
			KeyType:       aws.String(dynamodb.KeyTypeHash),
		})
	}

	if v, ok := data["row_key"]; ok && v != nil && v != "" {
		keySchema = append(keySchema, &dynamodb.KeySchemaElement{
			AttributeName: aws.String(v.(string)),
			KeyType:       aws.String(dynamodb.KeyTypeRange),
		})
	}

	return keySchema
}
