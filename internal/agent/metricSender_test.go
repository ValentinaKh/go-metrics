package agent

import (
	"context"
	"errors"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/stretchr/testify/mock"
	"sync"
	"testing"
	"time"
)

type MockSender struct {
	mock.Mock
	mu sync.Mutex
}

func (m *MockSender) Send(data []*models.Metrics) error {
	args := m.Called(data)
	return args.Error(0)
}

func TestMetricSender_Push_SendsData(t *testing.T) {
	mockSender := new(MockSender)
	mChan := make(chan []*models.Metrics, 10)

	sender := NewMetricSender([]Sender{mockSender}, mChan)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sender.Push(ctx)

	data1 := []*models.Metrics{{
		ID:    "test",
		MType: "gauge",
		Value: toPtr(42.0),
	},
	}
	data2 := []*models.Metrics{{
		ID:    "test",
		MType: "gauge",
		Value: toPtr(45.0),
	},
	}

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
	mChan := make(chan []*models.Metrics, 10)
	defer close(mChan)

	sender := NewMetricSender([]Sender{mockSender}, mChan)
	ctx, cancel := context.WithCancel(context.Background())

	go sender.Push(ctx)
	cancel()
	time.Sleep(50 * time.Millisecond)

	mChan <- []*models.Metrics{{
		ID:    "test",
		MType: "gauge",
		Value: toPtr(42.0),
	},
	}

	time.Sleep(50 * time.Millisecond)

	mockSender.AssertNotCalled(t, "Send")
}

func TestMetricSender_Push_ContinuesOnSendError(t *testing.T) {
	mockSender := new(MockSender)
	mChan := make(chan []*models.Metrics, 10)

	sender := NewMetricSender([]Sender{mockSender}, mChan)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sender.Push(ctx)

	data1 := []*models.Metrics{{
		ID:    "error",
		MType: "gauge",
		Value: toPtr(42.0),
	},
	}
	data2 := []*models.Metrics{{
		ID:    "next",
		MType: "gauge",
		Value: toPtr(42.0),
	},
	}

	mockSender.On("Send", data1).Return(errors.New("send failed"))
	mockSender.On("Send", data2).Return(nil)

	mChan <- data1
	mChan <- data2

	close(mChan)
	time.Sleep(50 * time.Millisecond)

	mockSender.AssertExpectations(t)
}
