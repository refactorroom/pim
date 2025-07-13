package main

import (
	"fmt"
	"time"

	pim "github.com/refactorroom/pim"
)

// User represents a user with various field types
type User struct {
	ID         int       `json:"id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Age        int       `json:"age"`
	Score      float64   `json:"score"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	EmptyField string    `json:"empty_field,omitempty"`
	ZeroField  int       `json:"zero_field"`
	unexported string    // This field will be included if IncludeUnexported is true
}

// Product represents a product with validation examples
type Product struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
}

func main() {
	// Create a user for demonstration
	user := User{
		ID:         123,
		Username:   "john_doe",
		Email:      "john.doe@example.com",
		FirstName:  "John",
		LastName:   "Doe",
		Age:        30,
		Score:      95.5,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		EmptyField: "",
		ZeroField:  0,
		unexported: "private data",
	}

	// Create a product for validation examples
	product := Product{
		ID:          "PROD-001",
		Name:        "Laptop",
		Description: "High-performance laptop",
		Price:       1299.99,
		Category:    "Electronics",
		Tags:        []string{"computer", "portable", "gaming"},
	}

	fmt.Println("=== Advanced JSON Serializer Examples ===\n")

	// Example 1: Basic serialization with default options
	fmt.Println("1. Basic Serialization:")
	serializer := pim.NewJsonSerializer()
	jsonData, err := serializer.MarshalToString(user)
	if err != nil {
		pim.Error("Failed to marshal user", err)
		return
	}
	pim.Json(jsonData)

	// Example 2: Custom field mapping
	fmt.Println("\n2. Custom Field Mapping:")
	mappingSerializer := pim.NewJsonSerializer()
	mappingSerializer.AddFieldMapping("Username", "user_name")
	mappingSerializer.AddFieldMapping("Email", "email_address")
	jsonData, err = mappingSerializer.MarshalToString(user)
	if err != nil {
		pim.Error("Failed to marshal with field mapping", err)
		return
	}
	pim.Json(jsonData)

	// Example 3: Field transformations
	fmt.Println("\n3. Field Transformations:")
	transformSerializer := pim.NewJsonSerializer()
	transformSerializer.AddFieldTransformer("Username", pim.StringToUpper)
	transformSerializer.AddFieldTransformer("FirstName", pim.StringToLower)
	transformSerializer.AddFieldTransformer("LastName", pim.StringTrim)
	transformSerializer.AddFieldTransformer("Email", pim.StringReplace("@", "[at]"))
	transformSerializer.AddFieldTransformer("Score", pim.NumberFormat(2))
	transformSerializer.AddFieldTransformer("CreatedAt", pim.TimeFormat("2006-01-02 15:04:05"))
	jsonData, err = transformSerializer.MarshalToString(user)
	if err != nil {
		pim.Error("Failed to marshal with transformations", err)
		return
	}
	pim.Json(jsonData)

	// Example 4: Field validations
	fmt.Println("\n4. Field Validations:")
	validationSerializer := pim.NewJsonSerializer()
	validationSerializer.AddFieldValidator("Username", pim.StringMinLength(3))
	validationSerializer.AddFieldValidator("Username", pim.StringMaxLength(20))
	validationSerializer.AddFieldValidator("Email", pim.StringPattern(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`))
	validationSerializer.AddFieldValidator("Age", pim.NumberMin(18))
	validationSerializer.AddFieldValidator("Age", pim.NumberMax(120))
	validationSerializer.AddFieldValidator("Price", pim.NumberMin(0))
	jsonData, err = validationSerializer.MarshalToString(product)
	if err != nil {
		pim.Error("Failed to marshal with validations", err)
		return
	}
	pim.Json(jsonData)

	// Example 5: Omit empty and zero values
	fmt.Println("\n5. Omit Empty and Zero Values:")
	omitSerializer := pim.NewJsonSerializer()
	omitSerializer.SetOmitEmpty(true)
	omitSerializer.SetOmitZero(true)
	jsonData, err = omitSerializer.MarshalToString(user)
	if err != nil {
		pim.Error("Failed to marshal with omit options", err)
		return
	}
	pim.Json(jsonData)

	// Example 6: Include unexported fields
	fmt.Println("\n6. Include Unexported Fields:")
	unexportedSerializer := pim.NewJsonSerializer()
	unexportedSerializer.SetIncludeUnexported(true)
	jsonData, err = unexportedSerializer.MarshalToString(user)
	if err != nil {
		pim.Error("Failed to marshal with unexported fields", err)
		return
	}
	pim.Json(jsonData)

	// Example 7: Null empty strings
	fmt.Println("\n7. Null Empty Strings:")
	nullSerializer := pim.NewJsonSerializer()
	nullSerializer.SetNullEmptyStrings(true)
	jsonData, err = nullSerializer.MarshalToString(user)
	if err != nil {
		pim.Error("Failed to marshal with null empty strings", err)
		return
	}
	pim.Json(jsonData)

	// Example 8: Use Number for precise number handling
	fmt.Println("\n8. Use Number for Precise Numbers:")
	numberSerializer := pim.NewJsonSerializer()
	numberSerializer.SetUseNumber(true)
	jsonData, err = numberSerializer.MarshalToString(user)
	if err != nil {
		pim.Error("Failed to marshal with number precision", err)
		return
	}
	pim.Json(jsonData)

	// Example 9: Custom time format
	fmt.Println("\n9. Custom Time Format:")
	timeSerializer := pim.NewJsonSerializer()
	timeSerializer.SetTimeFormat("2006-01-02")
	jsonData, err = timeSerializer.MarshalToString(user)
	if err != nil {
		pim.Error("Failed to marshal with custom time format", err)
		return
	}
	pim.Json(jsonData)

	// Example 10: Complex transformation with regex
	fmt.Println("\n10. Complex Regex Transformation:")
	regexSerializer := pim.NewJsonSerializer()
	regexSerializer.AddFieldTransformer("Email", pim.StringRegexReplace(`@([^.]+)\.`, "@[domain]."))
	regexSerializer.AddFieldTransformer("Username", pim.StringRegexReplace(`_`, "-"))
	jsonData, err = regexSerializer.MarshalToString(user)
	if err != nil {
		pim.Error("Failed to marshal with regex transformations", err)
		return
	}
	pim.Json(jsonData)

	// Example 11: Unmarshal with transformations
	fmt.Println("\n11. Unmarshal with Reverse Transformations:")
	jsonInput := `{
		"id": 456,
		"user_name": "jane_smith",
		"email_address": "jane.smith@example.com",
		"first_name": "JANE",
		"last_name": "SMITH",
		"age": 25,
		"score": "88.75",
		"is_active": true,
		"created_at": "2024-01-15 10:30:00",
		"updated_at": "2024-01-15 10:30:00"
	}`

	var newUser User
	err = mappingSerializer.UnmarshalFromString(jsonInput, &newUser)
	if err != nil {
		pim.Error("Failed to unmarshal user", err)
		return
	}

	pim.Info("Unmarshaled user:", newUser)

	// Example 12: Validation error handling
	fmt.Println("\n12. Validation Error Handling:")
	invalidUser := User{
		ID:        789,
		Username:  "ab",            // Too short
		Email:     "invalid-email", // Invalid format
		Age:       15,              // Too young
		FirstName: "Test",
		LastName:  "User",
	}

	_, err = validationSerializer.MarshalToString(invalidUser)
	if err != nil {
		pim.Error("Validation failed as expected:", err)
	} else {
		pim.Info("Validation should have failed")
	}

	// Example 13: Batch processing with different configurations
	fmt.Println("\n13. Batch Processing with Different Configurations:")
	users := []User{user, newUser}

	// Create different serializers for different use cases
	compactSerializer := pim.NewJsonSerializer()
	compactSerializer.SetIndent(0)
	compactSerializer.SetPrettyPrint(false)
	compactSerializer.SetOmitEmpty(true)

	verboseSerializer := pim.NewJsonSerializer()
	verboseSerializer.SetIndent(4)
	verboseSerializer.SetPrettyPrint(true)
	verboseSerializer.SetIncludeUnexported(true)

	for i, u := range users {
		pim.Info("User", i+1, "Compact:")
		compact, _ := compactSerializer.MarshalToString(u)
		pim.Json(compact)

		pim.Info("User", i+1, "Verbose:")
		verbose, _ := verboseSerializer.MarshalToString(u)
		pim.Json(verbose)
	}

	// Example 14: Custom complex transformation
	fmt.Println("\n14. Custom Complex Transformation:")
	customSerializer := pim.NewJsonSerializer()

	// Custom transformer that combines first and last name
	customSerializer.AddFieldTransformer("FirstName", func(value interface{}) (interface{}, error) {
		if firstName, ok := value.(string); ok {
			return firstName + " " + user.LastName, nil
		}
		return value, nil
	})

	// Custom validator that checks for profanity (simplified)
	customSerializer.AddFieldValidator("Username", func(value interface{}) error {
		if username, ok := value.(string); ok {
			profanity := []string{"bad", "inappropriate", "spam"}
			for _, word := range profanity {
				if username == word {
					return fmt.Errorf("username contains inappropriate content")
				}
			}
		}
		return nil
	})

	jsonData, err = customSerializer.MarshalToString(user)
	if err != nil {
		pim.Error("Failed to marshal with custom transformations", err)
		return
	}
	pim.Json(jsonData)

	fmt.Println("\n=== Advanced JSON Serializer Examples Complete ===")
}
