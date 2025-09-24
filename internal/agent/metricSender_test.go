package agent

import (
	"context"
	"errors"
	"github.com/stretchr/testify/mock"
	"sync"
	"testing"
	"time"
)

type MockSender struct {
	mock.Mock
	mu sync.Mutex
}

func (m *MockSender) Send(data []byte) error {
	args := m.Called(data)
	return args.Error(0)
}

func TestMetricSender_Push_SendsData(t *testing.T) {
	mockSender := new(MockSender)
	mChan := make(chan []byte, 10)

	sender := NewMetricSender(mockSender, mChan)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sender.Push(ctx)

	data1 := []byte("test-metric-1")
	data2 := []byte("test-metric-2")

	mockSender.On("Send", data1).Return(nil)
	mockSender.On("Send", data2).Return(nil)

	mChan <- data1
	mChan <- data2

	time.Sleep(200 * time.Millisecond)

	mockSender.AssertExpectations(t)
	close(mChan)
}

func TestMetricSender_Push_StopsOnContextCancel(t *testing.T) {
	mockSender := new(MockSender)
	mChan := make(chan []byte, 10)
	defer close(mChan)

	sender := NewMetricSender(mockSender, mChan)
	ctx, cancel := context.WithCancel(context.Background())

	go sender.Push(ctx)
	cancel()
	time.Sleep(50 * time.Millisecond)

	mChan <- []byte("should-not-be-sent")

	time.Sleep(50 * time.Millisecond)

	mockSender.AssertNotCalled(t, "Send")
}

func TestMetricSender_Push_ContinuesOnSendError(t *testing.T) {
	mockSender := new(MockSender)
	mChan := make(chan []byte, 10)

	sender := NewMetricSender(mockSender, mChan)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sender.Push(ctx)

	data1 := []byte("error-metric")
	data2 := []byte("next-metric")

	mockSender.On("Send", data1).Return(errors.New("send failed"))
	mockSender.On("Send", data2).Return(nil)

	mChan <- data1
	mChan <- data2

	close(mChan)
	time.Sleep(50 * time.Millisecond)

	mockSender.AssertExpectations(t)
}
