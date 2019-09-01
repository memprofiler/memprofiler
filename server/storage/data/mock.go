package data

import (
	"context"

	"github.com/stretchr/testify/mock"
)

var _ Loader = (*LoaderMock)(nil)

// LoaderMock ...
type LoaderMock struct {
	mock.Mock
}

// Load ...
func (m *LoaderMock) Load(ctx context.Context) (<-chan *LoadResult, error) {
	args := m.Called(ctx)
	return args.Get(0).(chan *LoadResult), args.Error(1)
}

// Close ...
func (m *LoaderMock) Close() error { return m.Called().Error(0) }
