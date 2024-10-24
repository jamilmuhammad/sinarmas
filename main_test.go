package test_test

import (
	"context"
	test "sinarmas"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProductService_PostProduct(t *testing.T) {
	tests := []struct {
		name          string
		input         *test.PostProduct
		expectedTotal int
		expectedCats  map[string]int // category -> expected number of items
		wantErr       bool
	}{
		{
			name: "success_mixed_categories",
			input: &test.PostProduct{
				Product: []test.ModelProduct{
					{Id: 1, Name: "Computer", Price: 10000, Category: "Electronic"},
					{Id: 2, Name: "Soap", Price: 20000, Category: "Body Care"},
					{Id: 3, Name: "Handphone", Price: 30000, Category: "Electronic"},
				},
			},
			expectedTotal: 60000,
			expectedCats: map[string]int{
				"Electronic": 2,
				"Body Care":  1,
			},
			wantErr: false,
		},
		{
			name: "success_single_category",
			input: &test.PostProduct{
				Product: []test.ModelProduct{
					{Id: 1, Name: "Sprite", Price: 40000, Category: "Beverage"},
					{Id: 2, Name: "Fanta", Price: 50000, Category: "Beverage"},
				},
			},
			expectedTotal: 90000,
			expectedCats: map[string]int{
				"Beverage": 2,
			},
			wantErr: false,
		},
		{
			name: "success_empty_products",
			input: &test.PostProduct{
				Product: []test.ModelProduct{},
			},
			expectedTotal: 0,
			expectedCats:  map[string]int{},
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize service
			service := test.NewService()

			// Execute
			result, err := service.PostProduct(context.Background(), tt.input)

			// Error check
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Validate basic fields
			assert.NotEmpty(t, result.Id)
			assert.True(t, result.TransactionDate.Before(time.Now()))
			assert.Equal(t, tt.expectedTotal, result.TransactionTotal)

			// Validate categories and items
			categoryCount := make(map[string]int)
			for _, detail := range result.TransactionDetail {
				categoryCount[detail.Category] = len(detail.Items)

				// Check if items are sorted by ID
				for i := 1; i < len(detail.Items); i++ {
					assert.True(t, detail.Items[i-1].ID <= detail.Items[i].ID,
						"Items should be sorted by ID within category")
				}
			}
			assert.Equal(t, tt.expectedCats, categoryCount)
		})
	}
}

func TestRandomString(t *testing.T) {
	tests := []struct {
		name string
		n    int
	}{
		{
			name: "success_random_string_3",
			n:    3,
		},
		{
			name: "success_random_string_5",
			n:    5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := test.RandomString(tt.n)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "TR")
		})
	}
}

// Helper function to compare transaction details
func compareTransactionDetails(t *testing.T, expected, actual []test.ProductDetails) {
	assert.Equal(t, len(expected), len(actual))
	for i := range expected {
		assert.Equal(t, expected[i].Category, actual[i].Category)
		assert.Equal(t, len(expected[i].Items), len(actual[i].Items))
		for j := range expected[i].Items {
			assert.Equal(t, expected[i].Items[j].ID, actual[i].Items[j].ID)
			assert.Equal(t, expected[i].Items[j].Name, actual[i].Items[j].Name)
			assert.Equal(t, expected[i].Items[j].Price, actual[i].Items[j].Price)
		}
	}
}

func TestProductService_PostProduct_Integration(t *testing.T) {
	// This test case demonstrates a more complex integration scenario
	t.Run("complex_integration_test", func(t *testing.T) {
		service := test.NewService()
		input := &test.PostProduct{
			Product: []test.ModelProduct{
				{Id: 6, Name: "Laptop", Price: 80000, Category: "Electronic"},
				{Id: 1, Name: "Computer", Price: 10000, Category: "Electronic"},
				{Id: 4, Name: "Sprite", Price: 40000, Category: "Beverage"},
				{Id: 2, Name: "Soap", Price: 20000, Category: "Body Care"},
				{Id: 5, Name: "Fanta", Price: 50000, Category: "Beverage"},
				{Id: 3, Name: "Handphone", Price: 30000, Category: "Electronic"},
			},
		}

		result, err := service.PostProduct(context.Background(), input)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Validate total
		expectedTotal := 230000
		assert.Equal(t, expectedTotal, result.TransactionTotal)

		// Validate categories exist and have correct number of items
		categoryItemCount := map[string]int{
			"Electronic": 3,
			"Beverage":   2,
			"Body Care":  1,
		}

		for _, detail := range result.TransactionDetail {
			expectedCount, exists := categoryItemCount[detail.Category]
			assert.True(t, exists, "Category should exist: "+detail.Category)
			assert.Equal(t, expectedCount, len(detail.Items))

			// Verify items are sorted within each category
			for i := 1; i < len(detail.Items); i++ {
				assert.True(t, detail.Items[i-1].ID < detail.Items[i].ID,
					"Items should be sorted by ID within category")
			}
		}
	})
}

func TestProductService_PostProduct_Negative(t *testing.T) {
	tests := []struct {
		name    string
		input   *test.PostProduct
		wantErr bool
	}{
		{
			name:    "nil_input",
			input:   nil,
			wantErr: true,
		},
		{
			name: "negative_price",
			input: &test.PostProduct{
				Product: []test.ModelProduct{
					{Id: 1, Name: "Invalid", Price: -1000, Category: "Test"},
				},
			},
			wantErr: true,
		},
		{
			name: "empty_category",
			input: &test.PostProduct{
				Product: []test.ModelProduct{
					{Id: 1, Name: "Test", Price: 1000, Category: ""},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := test.NewService()
			result, err := service.PostProduct(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}
