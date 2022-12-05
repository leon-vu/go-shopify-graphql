package shopify

import (
	"context"
	"fmt"
	"strings"

	"github.com/gempages/go-shopify-graphql/graph/models"
	log "github.com/sirupsen/logrus"
)

//go:generate mockgen -destination=./mock/metafield_service.go -package=mock . MetafieldService
type MetafieldService interface {
	ListAllShopMetafields() ([]models.Metafield, error)
	ListShopMetafieldsByNamespace(namespace string) ([]models.Metafield, error)

	GetShopMetafieldByKey(namespace, key string) (*models.Metafield, error)

	Delete(metafield models.MetafieldDeleteInput) error
	DeleteBulk(metafield []models.MetafieldDeleteInput) error
}

type MetafieldServiceOp struct {
	client *Client
}

var _ MetafieldService = &MetafieldServiceOp{}

type mutationMetafieldDelete struct {
	MetafieldDeleteResult struct {
		UserErrors []models.UserError `json:"userErrors,omitempty"`
	} `graphql:"metafieldDelete(input: $input)" json:"metafieldDelete"`
}

func (s *MetafieldServiceOp) ListAllShopMetafields() ([]models.Metafield, error) {
	q := `
		{
			shop{
				metafields{
					edges{
						node{
							createdAt
							description
							id
							key
							legacyResourceId
							namespace
							ownerType
							updatedAt
							value
							type
						}
					}
				}
			}
		}
`

	res := []models.Metafield{}
	err := s.client.BulkOperation.BulkQuery(q, &res)
	if err != nil {
		return nil, fmt.Errorf("bulk query: %w", err)
	}

	return res, nil
}

func (s *MetafieldServiceOp) ListShopMetafieldsByNamespace(namespace string) ([]models.Metafield, error) {
	q := `
		{
			shop{
				metafields(namespace: "$namespace"){
					edges{
						node{
							createdAt
							description
							id
							key
							legacyResourceId
							namespace
							ownerType
							updatedAt
							value
							type
						}
					}
				}
			}
		}
`
	q = strings.ReplaceAll(q, "$namespace", namespace)

	res := []models.Metafield{}
	err := s.client.BulkOperation.BulkQuery(q, &res)
	if err != nil {
		return nil, fmt.Errorf("bulk query: %w", err)
	}

	return res, nil
}

func (s *MetafieldServiceOp) GetShopMetafieldByKey(namespace, key string) (*models.Metafield, error) {
	var q struct {
		Shop struct {
			Metafield models.Metafield `graphql:"metafield(namespace: $namespace, key: $key)"`
		} `graphql:"shop"`
	}
	vars := map[string]interface{}{
		"namespace": namespace,
		"key":       key,
	}

	err := s.client.gql.Query(context.Background(), &q, vars)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return &q.Shop.Metafield, nil
}

func (s *MetafieldServiceOp) DeleteBulk(metafields []models.MetafieldDeleteInput) error {
	for _, m := range metafields {
		err := s.Delete(m)
		if err != nil {
			log.Warnf("Couldn't delete metafield (%v): %s", m, err)
		}
	}

	return nil
}

func (s *MetafieldServiceOp) Delete(metafield models.MetafieldDeleteInput) error {
	m := mutationMetafieldDelete{}

	vars := map[string]interface{}{
		"input": metafield,
	}
	err := s.client.gql.Mutate(context.Background(), &m, vars)
	if err != nil {
		return fmt.Errorf("mutation: %w", err)
	}

	if len(m.MetafieldDeleteResult.UserErrors) > 0 {
		return fmt.Errorf("%+v", m.MetafieldDeleteResult.UserErrors)
	}

	return nil
}
