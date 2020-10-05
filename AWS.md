# AWS Resource Mapping

Unlike most Terraform providers, which aim to map directly to the APIs of their underlying services, the Plausible provider may create multiple AWS resources to fulfill the capabilities of a single, higher-level resource. This document describes this mapping.

## **Function**
* ➜ Lambda Function
* **Schedule Trigger**
    * ➜ X Cloudwatch Rule
    * ➜ X Lambda Permission
    * ➜ X Cloudwatch Event Target
* **API Route Trigger**
    * *existing API Gateway Method* ⤇
    * ➜ X Lambda Permission
    * ➜ X API Method Integration
* **Subscription Trigger**
    * *existing SNS Topic* ⤇
    * ➜ X SQS Queue
    * ➜ X SNS Subscription
    * ➜ Lambda EventSource Mapping
* **Datastore Trigger** (KeyValue)
    * *existing DynamoDB Table* ⤇
    * ➜ DynamoDB Stream
    * ➜ Lambda EventSource Mapping
* **Datastore Trigger** (Object)
    * *existing S3 Bucket* ⤇
    * ➜ Lambda Permission
    * ➜ S3 Bucket Notification
* **Output - KeyValue Store**
    * *built-in facility*
* **Output - Object Store**
    * *built-in facility* OR 
    * ➜ Kinesis Firehose Delivery
* **Output - Publisher**
    * *Direct*
* **Output - Function**
    * ➜ SQS Queue
    * ➜ [Role?]

## ObjectStore
* ➜ S3 

## KeyValue Store
* ➜ DynamoDB Table & Global Secondary Indexes

## Publisher
* ➜ SNS Topic

## HTTP API
* ➜ API Gateway REST API
* ➜ API Stage
* ➜ Deployment
* ➜ Resources
* ➜ Methods