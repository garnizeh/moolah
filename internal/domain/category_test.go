package domain

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestCreateCategoryInput_Validation(t *testing.T) {
	t.Parallel()

	validate := validator.New()

	tests := []struct {
		name    string
		input   CreateCategoryInput
		wantErr bool
	}{
		{
			name: "valid root category",
			input: CreateCategoryInput{
				Name: "Groceries",
				Type: CategoryTypeExpense,
			},
			wantErr: false,
		},
		{
			name: "valid sub-category",
			input: CreateCategoryInput{
				ParentID: "01H7XRM1Z8P8P8P8P8P8P8P8P8",
				Name:     "Fruits",
				Type:     CategoryTypeExpense,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			input: CreateCategoryInput{
				Type: CategoryTypeExpense,
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			input: CreateCategoryInput{
				Name: "Invalid",
				Type: "something-else",
			},
			wantErr: true,
		},
		{
			name: "invalid color hex",
			input: CreateCategoryInput{
				Name:  "Blue",
				Type:  CategoryTypeIncome,
				Color: "not-a-hex",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validate.Struct(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCategory_IsRoot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		parentID string
		want     bool
	}{
		{"root category (empty string)", "", true},
		{"child category (has ID)", "ulid_parent_123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &Category{ParentID: tt.parentID}
			assert.Equal(t, tt.want, c.IsRoot())
		})
	}
}
