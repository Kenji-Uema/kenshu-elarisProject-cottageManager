package db

import (
	"context"

	"github.com/Kenji-Uema/cottageManager/internal/port"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type mongoTransactionManager struct {
	client *mongo.Client
}

func NewMongoTxManager(client *mongo.Client) port.TransactionManager {
	return &mongoTransactionManager{client: client}
}

func (m *mongoTransactionManager) WithTransaction(ctx context.Context, callback func(ctx context.Context) (any, error)) (any, error) {
	transactionSession, err := m.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer transactionSession.EndSession(ctx)

	return transactionSession.WithTransaction(ctx, callback)
}
