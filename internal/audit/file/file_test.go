package file

import (
	"errors"
	"github.com/ValentinaKh/go-metrics/internal/audit"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockWriter struct {
	mock.Mock
}

func (m *MockWriter) Write(v any) error {
	args := m.Called(v)
	return args.Error(0)
}

func (m *MockWriter) Close() error {
	return nil
}

func TestAuditHandler_Update_Success(t *testing.T) {
	mockWriter := new(MockWriter)
	handler := NewFileAuditHandler(mockWriter)

	request := []models.Metrics{{ID: "test_metric", Delta: nil, Value: nil}}
	ip := "192.168.1.1"

	dtoArg := audit.Dto{}
	mockWriter.On("Write", mock.AnythingOfType("audit.Dto")).Run(func(args mock.Arguments) {
		dtoArg = args.Get(0).(audit.Dto)
	}).Return(nil)

	handler.Update(request, ip)

	mockWriter.AssertExpectations(t)

	assert.Equal(t, ip, dtoArg.IPAddress)
	assert.Equal(t, request, dtoArg.Metrics)

}

func TestAuditHandler_Update_WriteError(t *testing.T) {
	mockWriter := new(MockWriter)
	handler := NewFileAuditHandler(mockWriter)

	expectedError := errors.New("write error")
	request := []models.Metrics{{ID: "error_metric", Delta: nil, Value: nil}}
	ip := "192.168.1.2"

	mockWriter.On("Write", mock.AnythingOfType("audit.Dto")).Return(expectedError)

	handler.Update(request, ip)

	mockWriter.AssertExpectations(t)
}
