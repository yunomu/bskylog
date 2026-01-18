package crawlerdb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDB struct {
	client    *dynamodb.Client
	tableName string
}

var _ DB = (*DynamoDB)(nil)

func NewDynamoDB(
	client *dynamodb.Client,
	tableName string,
) *DynamoDB {
	return &DynamoDB{
		client:    client,
		tableName: tableName,
	}
}

type DynamoDBRecord struct {
	Did       string `dynamodbav:"Did"`
	LatestCid string `dynamodbav:"Latest"`
	TS        int64  `dynamodbav:"TS"`
}

func dynamoToTimestamp(rec *DynamoDBRecord) *Timestamp {
	return &Timestamp{
		Did:       rec.Did,
		LatestCid: rec.LatestCid,
		Timestamp: rec.TS,
	}
}

func (d *DynamoDB) Get(ctx context.Context, did string) (*Timestamp, error) {
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

	if out.Item == nil {
		return nil, ErrNotExists
	}

	var rec DynamoDBRecord
	if err := attributevalue.UnmarshalMap(out.Item, &rec); err != nil {
		return nil, err
	}

	return dynamoToTimestamp(&rec), nil
}

func (d *DynamoDB) Scan(ctx context.Context, f func(*Timestamp) error) error {
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

			if err := f(dynamoToTimestamp(&rec)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *DynamoDB) Put(ctx context.Context, ts *Timestamp) error {
	item, err := attributevalue.MarshalMap(&DynamoDBRecord{
		Did:       ts.Did,
		LatestCid: ts.LatestCid,
		TS:        ts.Timestamp,
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
