package eventddb_test

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm/broker/eventddb"
)

// Test structures
type TestUser struct {
	ID    string `dynamodbav:"id"`
	Name  string `dynamodbav:"name"`
	Age   int    `dynamodbav:"age"`
	Email string `dynamodbav:"email,omitempty"`
}

type TestProduct struct {
	ProductID string  `dynamodbav:"product_id"`
	Title     string  `dynamodbav:"title"`
	Price     float64 `dynamodbav:"price"`
	InStock   bool    `dynamodbav:"in_stock"`
}

type TestComplexStruct struct {
	ID       string            `dynamodbav:"id"`
	Tags     []string          `dynamodbav:"tags"`
	Metadata map[string]string `dynamodbav:"metadata"`
	Count    int64             `dynamodbav:"count"`
}

func TestUnmarshalMap(t *testing.T) {
	t.Run("UnmarshalSimpleStruct", func(t *testing.T) {
		// Create DynamoDB AttributeValue map
		attributeMap := map[string]events.DynamoDBAttributeValue{
			"id":   events.NewStringAttribute("user-123"),
			"name": events.NewStringAttribute("John Doe"),
			"age":  events.NewNumberAttribute("30"),
		}

		var user TestUser
		err := eventddb.UnmarshalMap(attributeMap, &user)

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(user.ID, "user-123"),
			it.Equal(user.Name, "John Doe"),
			it.Equal(user.Age, 30),
		)
	})

	t.Run("UnmarshalWithOptionalFields", func(t *testing.T) {
		// Test with optional email field present
		attributeMap := map[string]events.DynamoDBAttributeValue{
			"id":    events.NewStringAttribute("user-456"),
			"name":  events.NewStringAttribute("Jane Smith"),
			"age":   events.NewNumberAttribute("25"),
			"email": events.NewStringAttribute("jane@example.com"),
		}

		var user TestUser
		err := eventddb.UnmarshalMap(attributeMap, &user)

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(user.ID, "user-456"),
			it.Equal(user.Name, "Jane Smith"),
			it.Equal(user.Age, 25),
			it.Equal(user.Email, "jane@example.com"),
		)
	})

	t.Run("UnmarshalWithMissingOptionalFields", func(t *testing.T) {
		// Test with optional email field missing
		attributeMap := map[string]events.DynamoDBAttributeValue{
			"id":   events.NewStringAttribute("user-789"),
			"name": events.NewStringAttribute("Bob Wilson"),
			"age":  events.NewNumberAttribute("35"),
		}

		var user TestUser
		err := eventddb.UnmarshalMap(attributeMap, &user)

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(user.ID, "user-789"),
			it.Equal(user.Name, "Bob Wilson"),
			it.Equal(user.Age, 35),
			it.Equal(user.Email, ""), // Should be empty string
		)
	})

	t.Run("UnmarshalBooleanAndFloat", func(t *testing.T) {
		attributeMap := map[string]events.DynamoDBAttributeValue{
			"product_id": events.NewStringAttribute("prod-123"),
			"title":      events.NewStringAttribute("Test Product"),
			"price":      events.NewNumberAttribute("29.99"),
			"in_stock":   events.NewBooleanAttribute(true),
		}

		var product TestProduct
		err := eventddb.UnmarshalMap(attributeMap, &product)

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(product.ProductID, "prod-123"),
			it.Equal(product.Title, "Test Product"),
			it.Equal(product.Price, 29.99),
			it.Equal(product.InStock, true),
		)
	})

	t.Run("UnmarshalComplexTypes", func(t *testing.T) {
		// Create complex DynamoDB AttributeValue map with lists and maps
		tags := events.NewStringSetAttribute([]string{"tag1", "tag2", "tag3"})
		metadata := events.NewMapAttribute(map[string]events.DynamoDBAttributeValue{
			"key1": events.NewStringAttribute("value1"),
			"key2": events.NewStringAttribute("value2"),
		})

		attributeMap := map[string]events.DynamoDBAttributeValue{
			"id":       events.NewStringAttribute("complex-123"),
			"tags":     tags,
			"metadata": metadata,
			"count":    events.NewNumberAttribute("42"),
		}

		var complex TestComplexStruct
		err := eventddb.UnmarshalMap(attributeMap, &complex)

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(complex.ID, "complex-123"),
			it.Equal(len(complex.Tags), 3),
			it.Equal(complex.Tags[0], "tag1"),
			it.Equal(complex.Metadata["key1"], "value1"),
			it.Equal(complex.Metadata["key2"], "value2"),
			it.Equal(complex.Count, int64(42)),
		)
	})

	t.Run("UnmarshalEmptyMap", func(t *testing.T) {
		attributeMap := map[string]events.DynamoDBAttributeValue{}

		var user TestUser
		err := eventddb.UnmarshalMap(attributeMap, &user)

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(user.ID, ""),
			it.Equal(user.Name, ""),
			it.Equal(user.Age, 0),
		)
	})

	t.Run("UnmarshalNilTarget", func(t *testing.T) {
		attributeMap := map[string]events.DynamoDBAttributeValue{
			"id": events.NewStringAttribute("test"),
		}

		err := eventddb.UnmarshalMap(attributeMap, nil)

		it.Then(t).ShouldNot(
			it.Nil(err),
		)
	})

	t.Run("UnmarshalNonPointerTarget", func(t *testing.T) {
		attributeMap := map[string]events.DynamoDBAttributeValue{
			"id": events.NewStringAttribute("test"),
		}

		var user TestUser
		err := eventddb.UnmarshalMap(attributeMap, user) // Pass by value instead of pointer

		it.Then(t).ShouldNot(
			it.Nil(err),
		)
	})

	t.Run("UnmarshalInvalidNumberFormat", func(t *testing.T) {
		// Create an attribute with invalid number format
		attributeMap := map[string]events.DynamoDBAttributeValue{
			"id":   events.NewStringAttribute("user-123"),
			"name": events.NewStringAttribute("John Doe"),
			"age":  events.NewStringAttribute("not-a-number"), // Invalid number
		}

		var user TestUser
		err := eventddb.UnmarshalMap(attributeMap, &user)

		it.Then(t).ShouldNot(
			it.Nil(err),
		)
	})

	t.Run("UnmarshalWithNullValues", func(t *testing.T) {
		attributeMap := map[string]events.DynamoDBAttributeValue{
			"id":    events.NewStringAttribute("user-123"),
			"name":  events.NewStringAttribute("John Doe"),
			"age":   events.NewNumberAttribute("30"),
			"email": events.NewNullAttribute(),
		}

		var user TestUser
		err := eventddb.UnmarshalMap(attributeMap, &user)

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(user.ID, "user-123"),
			it.Equal(user.Name, "John Doe"),
			it.Equal(user.Age, 30),
			it.Equal(user.Email, ""), // Null should result in zero value
		)
	})

	t.Run("UnmarshalBinaryData", func(t *testing.T) {
		type BinaryStruct struct {
			ID   string `json:"id"`
			Data []byte `json:"data"`
		}

		binaryData := []byte("binary test data")
		attributeMap := map[string]events.DynamoDBAttributeValue{
			"id":   events.NewStringAttribute("binary-123"),
			"data": events.NewBinaryAttribute(binaryData),
		}

		var binary BinaryStruct
		err := eventddb.UnmarshalMap(attributeMap, &binary)

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(binary.ID, "binary-123"),
			it.Equal(string(binary.Data), "binary test data"),
		)
	})
}

func TestUnmarshalMapEdgeCases(t *testing.T) {
	t.Run("UnmarshalLargeNumbers", func(t *testing.T) {
		type NumberStruct struct {
			BigInt   int64   `dynamodbav:"big_int"`
			BigFloat float64 `dynamodbav:"big_float"`
		}

		attributeMap := map[string]events.DynamoDBAttributeValue{
			"big_int":   events.NewNumberAttribute("9223372036854775807"),     // Max int64
			"big_float": events.NewNumberAttribute("1.7976931348623157e+308"), // Large float64
		}

		var numbers NumberStruct
		err := eventddb.UnmarshalMap(attributeMap, &numbers)

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(numbers.BigInt, int64(9223372036854775807)),
		)
	})

	t.Run("UnmarshalNestedMaps", func(t *testing.T) {
		type NestedStruct struct {
			ID     string                 `json:"id"`
			Config map[string]interface{} `json:"config"`
		}

		nestedMap := events.NewMapAttribute(map[string]events.DynamoDBAttributeValue{
			"setting1": events.NewStringAttribute("value1"),
			"setting2": events.NewNumberAttribute("42"),
			"setting3": events.NewBooleanAttribute(true),
		})

		attributeMap := map[string]events.DynamoDBAttributeValue{
			"id":     events.NewStringAttribute("nested-123"),
			"config": nestedMap,
		}

		var nested NestedStruct
		err := eventddb.UnmarshalMap(attributeMap, &nested)

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(nested.ID, "nested-123"),
		).ShouldNot(
			it.Nil(nested.Config),
		)
	})
}
