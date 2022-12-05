package shopify

import (
	"context"
	"fmt"

	"github.com/gempages/go-shopify-graphql/graph/models"
	"github.com/sirupsen/logrus"
)

type WebhookService interface {
	CreateWebhookSubscription(topic models.WebhookSubscriptionTopic, input models.WebhookSubscriptionInput) *models.WebhookSubscriptionCreatePayload
	CreateEventBridgeWebhookSubscription(topic models.WebhookSubscriptionTopic, input models.EventBridgeWebhookSubscriptionInput) *models.EventBridgeWebhookSubscriptionCreatePayload

	ListAll() ([]models.WebhookSubscription, error)
	Delete(webhookID string) (*models.WebhookSubscriptionDeletePayload, error)
}

type WebhookServiceOp struct {
	client *Client
}

type mutationWebhookCreate struct {
	WebhookCreateResult models.WebhookSubscriptionCreatePayload `graphql:"webhookSubscriptionCreate(topic: $topic, input: $input)" json:"webhookSubscriptionCreate"`
}

type mutationWebhookDelete struct {
	WebhookDeleteResult models.WebhookSubscriptionDeletePayload `graphql:"webhookSubscriptionDelete(id: $id)" json:"webhookSubscriptionDelete"`
}

type mutationEventBridgeWebhookCreate struct {
	EventBridgeWebhookCreateResult models.EventBridgeWebhookSubscriptionCreatePayload `graphql:"eventBridgeWebhookSubscriptionCreate(topic: $topic, input: $input)" json:"eventBridgeWebhookSubscriptionCreate"`
}

func (s *WebhookServiceOp) CreateWebhookSubscription(topic models.WebhookSubscriptionTopic, input models.WebhookSubscriptionInput) *models.WebhookSubscriptionCreatePayload {
	m := mutationWebhookCreate{}
	vars := map[string]interface{}{
		"topic": topic,
		"input": input,
	}
	err := s.client.gql.Mutate(context.Background(), &m, vars)
	if err != nil {
		return &m.WebhookCreateResult
	}

	if len(m.WebhookCreateResult.UserErrors) > 0 {
		err = fmt.Errorf("%+v", m.WebhookCreateResult.UserErrors)
		logrus.Info(err)
		return &m.WebhookCreateResult
	}

	return &m.WebhookCreateResult
}

func (s *WebhookServiceOp) CreateEventBridgeWebhookSubscription(topic models.WebhookSubscriptionTopic,
	input models.EventBridgeWebhookSubscriptionInput) *models.EventBridgeWebhookSubscriptionCreatePayload {

	m := mutationEventBridgeWebhookCreate{}
	vars := map[string]interface{}{
		"topic": topic,
		"input": input,
	}

	err := s.client.gql.Mutate(context.Background(), &m, vars)
	if err != nil {
		logrus.Info(err)
		return &m.EventBridgeWebhookCreateResult
	}

	if len(m.EventBridgeWebhookCreateResult.UserErrors) > 0 {
		err = fmt.Errorf("%+v", m.EventBridgeWebhookCreateResult.UserErrors)
		logrus.Info(err)
		return &m.EventBridgeWebhookCreateResult
	}

	return &m.EventBridgeWebhookCreateResult
}

func (s *WebhookServiceOp) Delete(webhookID string) (*models.WebhookSubscriptionDeletePayload, error) {
	m := mutationWebhookDelete{}
	vars := map[string]interface{}{
		"id": webhookID,
	}
	err := s.client.gql.Mutate(context.Background(), &m, vars)
	if err != nil {
		logrus.Info(err)
		return &m.WebhookDeleteResult, err
	}

	if len(m.WebhookDeleteResult.UserErrors) > 0 {
		err = fmt.Errorf("%+v", m.WebhookDeleteResult.UserErrors)
		logrus.Info(err)
	}

	return &m.WebhookDeleteResult, err
}

func (s *WebhookServiceOp) ListAll() ([]models.WebhookSubscription, error) {
	query := fmt.Sprintf(`{
    webhookSubscriptions(first: $first) {
      edges {
        node {
          id,
          topic,
          endpoint {
            __typename
            ... on WebhookHttpEndpoint {
              callbackUrl
            }
            ... on WebhookEventBridgeEndpoint{
              arn
            }
          }
          callbackUrl
          format
          topic
          includeFields
          createdAt
          updatedAt
        }
      }
    }
  }
  `)

	vars := map[string]interface{}{
		"first": 50,
	}

	var out []models.WebhookSubscription
	err := s.client.gql.QueryString(context.Background(), query, vars, &out)
	if err != nil {
		return out, err
	}

	return out, nil
}
