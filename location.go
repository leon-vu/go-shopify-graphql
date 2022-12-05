package shopify

import (
	"context"
	"fmt"

	"github.com/gempages/go-shopify-graphql/graph/models"
)

//go:generate mockgen -destination=./mock/location_service.go -package=mock . LocationService
type LocationService interface {
	Get(id string) (*models.Location, error)
}

type LocationServiceOp struct {
	client *Client
}

var _ LocationService = &LocationServiceOp{}

func (s *LocationServiceOp) Get(id string) (*models.Location, error) {
	q := `query location($id: ID!) {
		location(id: $id){
			id
			name
		}
	}`

	vars := map[string]interface{}{
		"id": id,
	}

	var out struct {
		*models.Location `json:"location"`
	}
	err := s.client.gql.QueryString(context.Background(), q, vars, &out)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return out.Location, nil
}
