package shopify

import (
	"context"
	"fmt"

	"github.com/gempages/go-shopify-graphql/graph/models"
)

//go:generate mockgen -destination=./mock/variant_service.go -package=mock . VariantService
type VariantService interface {
	Update(variant models.ProductVariantInput) error
}

type VariantServiceOp struct {
	client *Client
}

var _ VariantService = &VariantServiceOp{}

type mutationProductVariantUpdate struct {
	ProductVariantUpdateResult struct {
		UserErrors []models.UserError `json:"userErrors,omitempty"`
	} `graphql:"productVariantUpdate(input: $input)" json:"productVariantUpdate"`
}

func (s *VariantServiceOp) Update(variant models.ProductVariantInput) error {
	m := mutationProductVariantUpdate{}

	vars := map[string]interface{}{
		"input": variant,
	}
	err := s.client.gql.Mutate(context.Background(), &m, vars)
	if err != nil {
		return fmt.Errorf("mutation: %w", err)
	}

	if len(m.ProductVariantUpdateResult.UserErrors) > 0 {
		return fmt.Errorf("%+v", m.ProductVariantUpdateResult.UserErrors)
	}

	return nil
}
