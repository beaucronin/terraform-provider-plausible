package plausible

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type RegistryItem struct {
	Id        string
	Type      string
	CreatedAt string
	Triggers  []*map[string]string
}

// Add or update a registry item
func registryPut(appName string, id string, _type string, triggers []*map[string]string) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := dynamodb.New(sess)

	t := time.Now().UTC()
	createdAt := t.Format("20060102150405")

	item := RegistryItem{
		Id:        id,
		Type:      _type,
		CreatedAt: createdAt,
	}

	if triggers != nil {
		item.Triggers = triggers
	}

	av, _ := dynamodbattribute.MarshalMap(item)

	tableName := TableName(appName)
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}
	_, err := svc.PutItem(input)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func registryGet(appName string, id string) (*RegistryItem, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := dynamodb.New(sess)

	tableName := TableName(appName)

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {S: aws.String(id)},
		},
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	if result.Item == nil {
		msg := "Could not find '" + id + "'"
		return nil, errors.New(msg)
	}

	item := RegistryItem{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
	}
	return &item, nil
}

func registryDelete(appName string, id string) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := dynamodb.New(sess)

	tableName := TableName(appName)

	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {S: aws.String(id)},
		},
	}

	_, err := svc.DeleteItem(input)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func TableName(appName string) string {
	return fmt.Sprintf("PlausibleRegistry%s", appName)
}
