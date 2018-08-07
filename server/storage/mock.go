package storage

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// DataLoader

var _ DataLoader = (*DataLoaderMock)(nil)

type DataLoaderMock struct {
	mock.Mock
}

func (m *DataLoaderMock) Load(ctx context.Context) (<-chan *LoadResult, error) {
	args := m.Called(ctx)
	return args.Get(0).(chan *LoadResult), args.Error(1)
}

func (m *DataLoaderMock) Close() error { return m.Called().Error(0) }
