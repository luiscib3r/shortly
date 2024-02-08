#!/bin/sh

# Start DynamoDB Local
java -jar DynamoDBLocal.jar -sharedDb -dbPath ./data &

# Wait for DynamoDB Local to start up
sleep 2

# Create the table
aws dynamodb create-table --region us-west-1 --table-name Shortcuts --attribute-definitions AttributeName=id,AttributeType=S --key-schema AttributeName=id,KeyType=HASH --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 --endpoint-url http://localhost:8000

# Keep the container running
tail -f /dev/null