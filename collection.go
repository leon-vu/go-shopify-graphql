package shopify

import (
	"context"
	"fmt"

	"github.com/gempages/go-shopify-graphql/graph/models"
	"github.com/labstack/gommon/log"
)

//go:generate mockgen -destination=./mock/collection_service.go -package=mock . CollectionService
type CollectionService interface {
	ListAll() ([]models.Collection, error)
	//ListByCursor(first int, cursor string) (*CollectionsQueryResult, error)
	//ListWithFields(first int, cursor string, query string, fields string) (*CollectionsQueryResult, error)

	Get(id string) (*models.Collection, error)
	//GetSingleCollection(id graphql.ID, cursor string) (*CollectionQueryResult, error)

	Create(collection models.CollectionInput) (*string, error)
	CreateBulk(collections []models.CollectionInput) error

	Update(collection models.CollectionInput) error
}

type CollectionServiceOp struct {
	client *Client
}

var _ CollectionService = &CollectionServiceOp{}

type mutationCollectionCreate struct {
	CollectionCreateResult struct {
		Collection *struct {
			ID string `json:"id,omitempty"`
		} `json:"collection,omitempty"`

		UserErrors []models.UserError `json:"userErrors,omitempty"`
	} `graphql:"collectionCreate(input: $input)" json:"collectionCreate"`
}

type mutationCollectionUpdate struct {
	CollectionCreateResult struct {
		UserErrors []models.UserError `json:"userErrors,omitempty"`
	} `graphql:"collectionUpdate(input: $input)" json:"collectionUpdate"`
}

var collectionQuery = `
	id
	handle
	title
	products(first:250, after: $cursor){
		edges{
			node{
				id
			}
			cursor
		}
		pageInfo{
			hasNextPage
		}
	}
`

var collectionBulkQuery = `
	id
	handle
	title
`

func (s *CollectionServiceOp) ListAll() ([]models.Collection, error) {
	q := fmt.Sprintf(`
		{
			collections{
				edges{
					node{
						%s
					}
				}
			}
		}
	`, collectionBulkQuery)

	res := []models.Collection{}
	err := s.client.BulkOperation.BulkQuery(q, &res)
	if err != nil {
		return nil, fmt.Errorf("bulk query: %w", err)
	}

	return res, nil
}

//func (s *CollectionServiceOp) ListByCursor(first int, cursor string) (*CollectionsQueryResult, error) {
//	q := fmt.Sprintf(`
//		query collections($first: Int!, $cursor: String) {
//			collections(first: $first, after: $cursor){
//                edges{
//					node {
//						%s
//					}
//                    cursor
//                }
//                pageInfo {
//                      hasNextPage
//                }
//			}
//		}
//	`, collectionBulkQuery)
//
//	vars := map[string]interface{}{
//		"first": first,
//	}
//	if cursor != "" {
//		vars["cursor"] = cursor
//	}
//
//	out := CollectionsQueryResult{}
//
//	err := s.client.gql.QueryString(context.Background(), q, vars, &out)
//	if err != nil {
//		return nil, err
//	}
//
//	return &out, nil
//}

//func (s *CollectionServiceOp) ListWithFields(first int, cursor, query, fields string) (*CollectionsQueryResult, error) {
//	if fields == "" {
//		fields = `id`
//	}
//
//	q := fmt.Sprintf(`
//		query collections($first: Int!, $cursor: String, $query: String) {
//			collections(first: $first, after: $cursor, query:$query){
//				edges{
//					cursor
//					node {
//						%s
//					}
//				}
//			}
//		}
//	`, fields)
//
//	vars := map[string]interface{}{
//		"first": first,
//	}
//	if cursor != "" {
//		vars["cursor"] = cursor
//	}
//	if query != "" {
//		vars["query"] = query
//	}
//	out := &CollectionsQueryResult{}
//
//	err := s.client.gql.QueryString(context.Background(), q, vars, &out)
//	if err != nil {
//		return nil, err
//	}
//
//	return out, nil
//}

func (s *CollectionServiceOp) Get(id string) (*models.Collection, error) {
	out, err := s.getPage(id, "")
	if err != nil {
		return nil, err
	}

	nextPageData := out
	hasNextPage := out.Products.PageInfo.HasNextPage
	for hasNextPage && len(nextPageData.Products.Edges) > 0 {
		cursor := nextPageData.Products.Edges[len(nextPageData.Products.Edges)-1].Cursor
		nextPageData, err := s.getPage(id, cursor)
		if err != nil {
			return nil, err
		}
		out.Products.Edges = append(out.Products.Edges, nextPageData.Products.Edges...)
		hasNextPage = nextPageData.Products.PageInfo.HasNextPage
	}

	return out, nil
}

func (s *CollectionServiceOp) getPage(id string, cursor string) (*models.Collection, error) {
	q := fmt.Sprintf(`
		query collection($id: ID!, $cursor: String) {
			collection(id: $id){
				%s
			}
		}
	`, collectionQuery)

	vars := map[string]interface{}{
		"id": id,
	}
	if cursor != "" {
		vars["cursor"] = cursor
	}

	out := struct {
		Collection *models.Collection `json:"collection"`
	}{}
	err := s.client.gql.QueryString(context.Background(), q, vars, &out)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return out.Collection, nil
}

func (s *CollectionServiceOp) CreateBulk(collections []models.CollectionInput) error {
	for _, c := range collections {
		_, err := s.client.Collection.Create(c)
		if err != nil {
			log.Warnf("Couldn't create collection (%v): %s", c, err)
		}
	}

	return nil
}

func (s *CollectionServiceOp) Create(collection models.CollectionInput) (*string, error) {
	m := mutationCollectionCreate{}

	vars := map[string]interface{}{
		"input": collection,
	}
	err := s.client.gql.Mutate(context.Background(), &m, vars)
	if err != nil {
		return nil, fmt.Errorf("mutation: %w", err)
	}

	if len(m.CollectionCreateResult.UserErrors) > 0 {
		return nil, fmt.Errorf("%+v", m.CollectionCreateResult.UserErrors)
	}

	return &m.CollectionCreateResult.Collection.ID, nil
}

func (s *CollectionServiceOp) Update(collection models.CollectionInput) error {
	m := mutationCollectionUpdate{}

	vars := map[string]interface{}{
		"input": collection,
	}
	err := s.client.gql.Mutate(context.Background(), &m, vars)
	if err != nil {
		return fmt.Errorf("mutation: %w", err)
	}

	if len(m.CollectionCreateResult.UserErrors) > 0 {
		return fmt.Errorf("%+v", m.CollectionCreateResult.UserErrors)
	}

	return nil
}
