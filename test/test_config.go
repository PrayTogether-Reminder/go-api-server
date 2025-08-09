package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
)

// TestConfig provides test configuration (matching Java IntegrateTestConfig)
type TestConfig struct {
	MockFirebaseApp       *MockFirebaseApp
	MockFirebaseMessaging *MockFirebaseMessaging
}

// NewTestConfig creates a new test configuration
func NewTestConfig() *TestConfig {
	return &TestConfig{
		MockFirebaseApp:       new(MockFirebaseApp),
		MockFirebaseMessaging: new(MockFirebaseMessaging),
	}
}

// MockFirebaseApp mocks Firebase App (matching Java mockFirebaseApp)
type MockFirebaseApp struct {
	mock.Mock
}

// MockFirebaseMessaging mocks Firebase Messaging (matching Java mockFirebaseMessaging)
type MockFirebaseMessaging struct {
	mock.Mock
}

// Send mocks sending a message
func (m *MockFirebaseMessaging) Send(ctx context.Context, message interface{}) (string, error) {
	args := m.Called(ctx, message)
	return args.String(0), args.Error(1)
}

// SendAll mocks sending multiple messages
func (m *MockFirebaseMessaging) SendAll(ctx context.Context, messages []interface{}) ([]string, error) {
	args := m.Called(ctx, messages)
	return args.Get(0).([]string), args.Error(1)
}

// SetupFirebaseMocks sets up Firebase mocks for testing
func SetupFirebaseMocks(t *testing.T) *TestConfig {
	config := NewTestConfig()

	// Setup default expectations
	config.MockFirebaseMessaging.On("Send", mock.Anything, mock.Anything).Return("mock-message-id", nil)
	config.MockFirebaseMessaging.On("SendAll", mock.Anything, mock.Anything).Return([]string{"mock-message-id-1", "mock-message-id-2"}, nil)

	return config
}
