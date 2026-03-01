package fakes

import "context"

type FakeTransactionManager struct {
	WithTransactionFunc func(ctx context.Context, callback func(ctx context.Context) (any, error)) (any, error)

	WithTransactionCalls int
}

func NewFakeTransactionManager() *FakeTransactionManager {
	return &FakeTransactionManager{}
}

func (f *FakeTransactionManager) WithTransaction(ctx context.Context, callback func(ctx context.Context) (any, error)) (any, error) {
	f.WithTransactionCalls++
	if f.WithTransactionFunc != nil {
		return f.WithTransactionFunc(ctx, callback)
	}
	return callback(ctx)
}
