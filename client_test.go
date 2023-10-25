package dynago_test

import (
	"context"
	"fmt"
	"github.com/oolio-group/dynago/v1"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ory/dockertest/v3"
)

func startLocalDatabase(t *testing.T) (addr string, purge func()) {
	t.Helper()
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("could not connect to docker: %s", err)
	}

	resource, err := pool.Run("amazon/dynamodb-local", "latest", []string{})
	if err != nil {
		log.Fatalf("could not start container: %s", err)
	}
	addr = fmt.Sprintf("http://localhost:%s", resource.GetPort("8000/tcp"))
	return addr, func() {
		if err := pool.Purge(resource); err != nil {
			t.Fatalf("could not purge container: %s", err)
		}
	}
}

func createTestTable(t *dynago.Client) error {
	_, err := t.GetDynamoDBClient().CreateTable(context.TODO(), &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{{
			AttributeName: aws.String("pk"),
			AttributeType: types.ScalarAttributeTypeS,
		}, {AttributeName: aws.String("sk"), AttributeType: types.ScalarAttributeTypeS}},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("pk"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("sk"), KeyType: types.KeyTypeRange},
		},
		TableName:   &t.TableName,
		BillingMode: types.BillingModePayPerRequest,
		TableClass:  types.TableClassStandard,
	})
	return err
}

func TestNewClientLocalEndpoint(t *testing.T) {
	endpoint, purge := startLocalDatabase(t)
	defer purge()

	table, err := dynago.NewClient(context.TODO(), dynago.ClientOptions{
		TableName: "test",
		Endpoint: &dynago.EndpointResolver{
			EndpointURL:     endpoint,
			AccessKeyID:     "dummy",
			SecretAccessKey: "dummy",
		},
		PartitionKeyName: "pk",
		SortKeyName:      "sk",
		Region:           "us-east-1",
	})
	if err != nil {
		t.Fatalf("expected configuration to succeed, got %s", err)
	}

	err = createTestTable(table)
	if err != nil {
		t.Fatalf("expected create table on local table to succeed, got %s", err)
	}
}
