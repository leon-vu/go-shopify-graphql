package shopify

import "github.com/gempages/go-shopify-graphql/graphql"

type UserErrors struct {
	Field   []graphql.String
	Message graphql.String
}

type Money string   // Serialized and truncated to 2 decimals decimal.Decimal
type Decimal string // Serialized decimal.Decimal

type MoneyV2 struct {
	Amount       Decimal      `json:"amount,omitempty"`
	CurrencyCode CurrencyCode `json:"currencyCode,omitempty"`
}

type MoneyBag struct {
	PresentmentMoney MoneyV2 `json:"presentmentMoney,omitempty"`
	ShopMoney        MoneyV2 `json:"shopMoney,omitempty"`
}

// CountryCode enum ISO 3166-1 alpha-2 country codes with some differences.
type CountryCode string

// CurrencyCode enum
// USD United States Dollars (USD).
// EUR Euro
// GBP British Pound
// ...
// see more at https://shopify.dev/docs/admin-api/graphql/reference/common-objects/currencycode
type CurrencyCode string

type DateTime string

type PageInfo struct {
	// Indicates if there are more pages to fetch.
	HasNextPage graphql.Boolean `json:"hasNextPage"`
	// Indicates if there are any pages prior to the current page.
	HasPreviousPage graphql.Boolean `json:"hasPreviousPage"`
}

// URL An RFC 3986 and RFC 3987 compliant URI string.
//
// Example value: "https://johns-apparel.myshopify.com".
type URL string
