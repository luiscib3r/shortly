package datasources

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/luiscib3r/shortly/app/internal/data/models"
	"github.com/luiscib3r/shortly/app/internal/domain/entities"
)

type ShortcutDynamoDB struct {
	TableName string
	svc       *dynamodb.Client
}

func NewShortcutDynamoDB() (*ShortcutDynamoDB, error) {
	tableName := os.Getenv("ShortcutsTableName")

	if tableName == "" {
		log.Printf("ShortcutsTableName environment variable not set")
		return nil, errors.New("ShortcutsTableName environment variable not set")
	}

	// Load SDK config
	if cfg, err := config.LoadDefaultConfig(context.TODO()); err == nil {
		// Create DynamoDB client
		svc := dynamodb.NewFromConfig(cfg)

		return &ShortcutDynamoDB{
			TableName: tableName,
			svc:       svc,
		}, nil
	} else {
		log.Printf("unable to load SDK config, %v", err)
		return nil, err
	}
}

func (s ShortcutDynamoDB) Find(
	limit int32,
) ([]entities.Shortcut, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(s.TableName),
		Limit:     aws.Int32(limit),
	}

	scanner := dynamodb.NewScanPaginator(
		s.svc,
		input,
	)

	output, scanError := scanner.NextPage(context.TODO())

	if scanError != nil {
		log.Printf("Got error calling Scan: %v", scanError)
		return nil, scanError
	}

	var items []models.ShortcutItem

	err := attributevalue.UnmarshalListOfMaps(output.Items, &items)

	if err != nil {
		log.Printf("Got error unmarshalling items: %v", err)
	}

	// Convert to entities
	var shortcuts []entities.Shortcut
	for _, item := range items {
		shortcuts = append(shortcuts, *entities.NewShortcut(item.Id, item.Url))
	}

	return shortcuts, nil
}

func (s ShortcutDynamoDB) FindById(id string) (entities.Shortcut, bool) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(s.TableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	}

	result, err := s.svc.GetItem(context.TODO(), input)

	if err != nil {
		log.Printf("Got error calling GetItem: %v", err)
		return entities.Shortcut{}, false
	}

	if result.Item == nil {
		return entities.Shortcut{}, false
	}

	var item models.ShortcutItem

	err = attributevalue.UnmarshalMap(result.Item, &item)

	if err != nil {
		log.Printf("Got error unmarshalling item: %v", err)
		return entities.Shortcut{}, false
	}

	return *entities.NewShortcut(item.Id, item.Url), true
}

func (s ShortcutDynamoDB) Save(entity entities.Shortcut) (bool, error) {
	input := &dynamodb.PutItemInput{
		TableName: aws.String(s.TableName),
		Item: map[string]types.AttributeValue{
			"id":  &types.AttributeValueMemberS{Value: entity.Id()},
			"url": &types.AttributeValueMemberS{Value: entity.Url()},
		},
	}

	if _, err := s.svc.PutItem(context.TODO(), input); err != nil {
		log.Printf("Got error saving Shortcut: %v", err)
		return false, err
	}

	return true, nil
}

func (s ShortcutDynamoDB) Delete(id string) bool {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(s.TableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	}

	if _, err := s.svc.DeleteItem(context.TODO(), input); err != nil {
		log.Printf("Got error deleting Shortcut: %v", err)
		return false
	}

	return true
}
