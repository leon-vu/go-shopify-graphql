package shopify

import (
	"context"
	"fmt"
	"strings"

	"github.com/gempages/go-shopify-graphql/graph/models"
	"github.com/gempages/go-shopify-graphql/graphql"
)

//go:generate mockgen -destination=./mock/order_service.go -package=mock . OrderService
type OrderService interface {
	Get(id graphql.ID) (*models.Order, error)

	List(opts ListOptions) ([]models.Order, error)
	ListAll() ([]models.Order, error)

	ListAfterCursor(opts ListOptions) ([]models.Order, *string, *string, error)

	Update(input models.OrderInput) error
}

type OrderServiceOp struct {
	client *Client
}

var _ OrderService = &OrderServiceOp{}

type mutationOrderUpdate struct {
	OrderUpdateResult struct {
		UserErrors []models.UserError `json:"userErrors,omitempty"`
	} `graphql:"orderUpdate(input: $input)" json:"orderUpdate"`
}

const orderBaseQuery = `
	id
	legacyResourceId
	name
	createdAt
	customer{
		id
		legacyResourceId
		firstName
		displayName
		email
	}
	clientIp
	shippingAddress{
		address1
		address2
		city
		province
		country
		zip
	}
	shippingLine{
		originalPriceSet{
			presentmentMoney{
				amount
				currencyCode
			}
			shopMoney{
				amount
				currencyCode
			}
		}
		title
	}
	taxLines{
		priceSet{
			presentmentMoney{
				amount
				currencyCode
			}
			shopMoney{
				amount
				currencyCode
			}
		}
		rate
		ratePercentage
		title
	}
	totalReceivedSet{
		presentmentMoney{
			amount
			currencyCode
		}
		shopMoney{
			amount
			currencyCode
		}
	}
	note
	tags
	transactions {
		processedAt
		status
		kind
		test
		amountSet {
			shopMoney {
				amount
				currencyCode
			}
		}
	}
`

const orderLightQuery = `
	id
	legacyResourceId
	name
	createdAt
	customer{
		id
		legacyResourceId
		firstName
		displayName
		email
	}
	shippingAddress{
		address1
		address2
		city
		province
		country
		zip
	}
	shippingLine{
		title
	}
	totalReceivedSet{
		shopMoney{
			amount
		}
	}
	note
	tags
`

const lineItemFragment = `
fragment lineItem on LineItem {
	id
	sku
	quantity
	fulfillableQuantity
	fulfillmentStatus
	product{
		id
		legacyResourceId
	}
	vendor
	title
	variantTitle
	variant{
		id
		legacyResourceId
		selectedOptions{
			name
			value
		}
	}
	originalTotalSet{
		presentmentMoney{
			amount
			currencyCode
		}
		shopMoney{
			amount
			currencyCode
		}
	}
	originalUnitPriceSet{
		presentmentMoney{
			amount
			currencyCode
		}
		shopMoney{
			amount
			currencyCode
		}
	}
	discountedUnitPriceSet{
		presentmentMoney{
			amount
			currencyCode
		}
		shopMoney{
			amount
			currencyCode
		}
	}
	discountedTotalSet{
		presentmentMoney{
			amount
			currencyCode
		}
		shopMoney{
			amount
			currencyCode
		}
	}
}
`

const lineItemFragmentLight = `
fragment lineItem on LineItem {
	id
	sku
	quantity
	fulfillableQuantity
	fulfillmentStatus
	vendor
	title
	variantTitle
}
`

func (s *OrderServiceOp) Get(id graphql.ID) (*models.Order, error) {
	q := fmt.Sprintf(`
		query order($id: ID!) {
			node(id: $id){
				... on Order {
					%s
					lineItems(first:50){
						edges{
							node{
								...lineItem
							}
						}
					}
					fulfillmentOrders(first:5){
						edges {
							node {
								id
								status
								lineItems(first:50){
									edges {
										node {
											id
											remainingQuantity
											totalQuantity
											lineItem{
												sku
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
		%s
	`, orderBaseQuery, lineItemFragment)

	vars := map[string]interface{}{
		"id": id,
	}

	out := struct {
		Order *models.Order `json:"node"`
	}{}
	err := s.client.gql.QueryString(context.Background(), q, vars, &out)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return out.Order, nil
}

func (s *OrderServiceOp) List(opts ListOptions) ([]models.Order, error) {
	q := fmt.Sprintf(`
		{
			orders(query: "$query"){
				edges{
					node{
						%s
						lineItems{
							edges{
								node{
									...lineItem
								}
							}
						}
					}
				}
			}
		}
		%s
	`, orderBaseQuery, lineItemFragment)

	q = strings.ReplaceAll(q, "$query", opts.Query)

	res := []models.Order{}
	err := s.client.BulkOperation.BulkQuery(q, &res)
	if err != nil {
		return nil, fmt.Errorf("bulk query: %w", err)
	}

	return res, nil
}

func (s *OrderServiceOp) ListAll() ([]models.Order, error) {
	q := fmt.Sprintf(`
		{
			orders(query: "$query"){
				edges{
					node{
						%s
						lineItems{
							edges{
								node{
									...lineItem
								}
							}
						}
					}
				}
			}
		}
		%s
	`, orderBaseQuery, lineItemFragment)

	res := []models.Order{}
	err := s.client.BulkOperation.BulkQuery(q, &res)
	if err != nil {
		return nil, fmt.Errorf("bulk query: %w", err)
	}

	return res, nil
}

func (s *OrderServiceOp) ListAfterCursor(opts ListOptions) ([]models.Order, *string, *string, error) {
	q := fmt.Sprintf(`
		query orders($query: String, $first: Int, $last: Int, $before: String, $after: String, $reverse: Boolean) {
			orders(query: $query, first: $first, last: $last, before: $before, after: $after, reverse: $reverse){
				edges{
					node{
						%s
						lineItems(first:25){
							edges{
								node{
									...lineItem
								}
							}
						}
					}
					cursor
				}
				pageInfo{
					hasNextPage
				}
			}
		}
		%s
	`, orderLightQuery, lineItemFragmentLight)

	vars := map[string]interface{}{
		"query":   opts.Query,
		"reverse": opts.Reverse,
	}

	if opts.After != "" {
		vars["after"] = opts.After
	} else if opts.Before != "" {
		vars["before"] = opts.Before
	}

	if opts.First > 0 {
		vars["first"] = opts.First
	} else if opts.Last > 0 {
		vars["last"] = opts.Last
	}

	out := struct {
		Orders struct {
			Edges []struct {
				OrderQueryResult *models.Order `json:"node,omitempty"`
				Cursor           string        `json:"cursor,omitempty"`
			} `json:"edges,omitempty"`
			PageInfo struct {
				HasNextPage bool `json:"hasNextPage,omitempty"`
			} `json:"pageInfo,omitempty"`
		} `json:"orders,omitempty"`
	}{}
	err := s.client.gql.QueryString(context.Background(), q, vars, &out)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("query: %w", err)
	}

	res := []models.Order{}
	var firstCursor *string
	var lastCursor *string
	if len(out.Orders.Edges) > 0 {
		firstCursor = &out.Orders.Edges[0].Cursor
		lastCursor = &out.Orders.Edges[len(out.Orders.Edges)-1].Cursor
		for _, o := range out.Orders.Edges {
			res = append(res, *o.OrderQueryResult)
		}
	}

	return res, firstCursor, lastCursor, nil
}

func (s *OrderServiceOp) Update(input models.OrderInput) error {
	m := mutationOrderUpdate{}

	vars := map[string]interface{}{
		"input": input,
	}
	err := s.client.gql.Mutate(context.Background(), &m, vars)
	if err != nil {
		return fmt.Errorf("mutation: %w", err)
	}

	if len(m.OrderUpdateResult.UserErrors) > 0 {
		return fmt.Errorf("%+v", m.OrderUpdateResult.UserErrors)
	}

	return nil
}
