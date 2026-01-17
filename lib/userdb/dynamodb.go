package userdb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDB struct {
	client *dynamodb.Client

	tableName   string
	handleIndex string
}

var _ DB = (*DynamoDB)(nil)

func NewDynamoDB(
	client *dynamodb.Client,
	tableName string,
	handleIndex string,
) *DynamoDB {
	return &DynamoDB{
		client:      client,
		tableName:   tableName,
		handleIndex: handleIndex,
	}
}

type DynamoDBRecord struct {
	Did      string `dynamodbav:"Did"`
	Handle   string `dynamodbav:"Handle"`
	Password string `dynamodbav:"PW"`
	TimeZone int    `dynamodbav:"TZ"`
}

func dynamoToUser(rec *DynamoDBRecord) *User {
	return &User{
		Did:      rec.Did,
		Handle:   rec.Handle,
		Password: rec.Password,
		TimeZone: rec.TimeZone,
	}
}

func (d *DynamoDB) Get(ctx context.Context, did string) (*User, error) {
	out, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"Did": &types.AttributeValueMemberS{
				Value: did,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	var rec DynamoDBRecord
	if err := attributevalue.UnmarshalMap(out.Item, &rec); err != nil {
		return nil, err
	}

	return dynamoToUser(&rec), nil
}

func (d *DynamoDB) GetByHandle(ctx context.Context, handle string) (*User, error) {
	expr, err := expression.NewBuilder().
		WithKeyCondition(expression.Key("Handle").Equal(expression.Value(handle))).
		Build()
	if err != nil {
		return nil, err
	}

	out, err := d.client.Query(ctx, &dynamodb.QueryInput{
		TableName: aws.String(d.tableName),
		IndexName: aws.String(d.handleIndex),

		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return nil, err
	}

	for _, item := range out.Items {
		var rec DynamoDBRecord
		if err := attributevalue.UnmarshalMap(item, &rec); err != nil {
			return nil, err
		}

		return dynamoToUser(&rec), nil
	}

	return nil, ErrNotExists
}

func (d *DynamoDB) Scan(ctx context.Context, f func(*User) error) error {
	paginator := dynamodb.NewScanPaginator(d.client, &dynamodb.ScanInput{
		TableName: aws.String(d.tableName),
	})
	for paginator.HasMorePages() {
		out, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}

		for _, item := range out.Items {
			var rec DynamoDBRecord
			if err := attributevalue.UnmarshalMap(item, &rec); err != nil {
				return err
			}

			if err := f(dynamoToUser(&rec)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *DynamoDB) Put(ctx context.Context, user *User) error {
	item, err := attributevalue.MarshalMap(&DynamoDBRecord{
		Did:      user.Did,
		Handle:   user.Handle,
		Password: user.Password,
		TimeZone: user.TimeZone,
	})
	if err != nil {
		return err
	}

	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.tableName),
		Item:      item,
	})

	return err
}
