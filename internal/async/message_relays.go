package async

import (
	hermesTopics "github.com/AlpacaLabs/api-hermes/pkg/topic"
	"github.com/AlpacaLabs/api-password/internal/configuration"
	"github.com/AlpacaLabs/api-password/internal/db"
)

func RelayMessagesForSendEmail(config configuration.Config, dbClient db.Client) {
	relayMessages(config, dbClient, relayMessagesInput{
		topic:                    hermesTopics.TopicForSendEmailRequest,
		transactionalOutboxTable: db.TableForSendEmailRequest,
	})
}

func RelayMessagesForSendSms(config configuration.Config, dbClient db.Client) {
	relayMessages(config, dbClient, relayMessagesInput{
		topic:                    hermesTopics.TopicForSendSmsRequest,
		transactionalOutboxTable: db.TableForSendSmsRequest,
	})
}
