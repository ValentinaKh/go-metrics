package retry

import (
	"context"
	"time"

	"github.com/ValentinaKh/go-metrics/internal/apperror"
)

type (
	RetryPolicy interface {
		ShouldRetry(attempt int, err error) bool
	}

	DelayStrategy interface {
		GetDelay(attempt int) time.Duration
	}

	TimeProvider interface {
		Sleep(ctx context.Context, d time.Duration)
	}
)

type ClassifierRetryPolicy struct {
	classifier apperror.ErrorClassifier
	maxAttempt int
}

func NewClassifierRetryPolicy(classifier apperror.ErrorClassifier, maxAttempt int) *ClassifierRetryPolicy {
	return &ClassifierRetryPolicy{
		classifier: classifier,
		maxAttempt: maxAttempt,
	}

}

func (c *ClassifierRetryPolicy) ShouldRetry(attempt int, err error) bool {
	if err == nil {
		return false
	}
	if attempt >= c.maxAttempt {
		return false
	}
	return c.classifier.Classify(err) == apperror.Retriable
}

type StaticDelayStrategy struct {
	delays []time.Duration
}

func NewStaticDelayStrategy(delays []time.Duration) *StaticDelayStrategy {
	return &StaticDelayStrategy{delays: delays}
}

func (s *StaticDelayStrategy) GetDelay(attempt int) time.Duration {
	if attempt < 0 {
		return time.Duration(0)
	}

	//если номер попытки больше массива, повторяем последний интервал
	if attempt > len(s.delays)-1 {
		return s.delays[len(s.delays)-1]
	}

	return s.delays[attempt]
}

type SleepTimeProvider struct {
}

func (*SleepTimeProvider) Sleep(ctx context.Context, d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return
	case <-timer.C:
		return
	}
}

type Retrier struct {
	retryPolicy   RetryPolicy
	delayStrategy DelayStrategy
	timeProvider  TimeProvider
}

func NewRetrier(retryPolicy RetryPolicy, delayStrategy DelayStrategy, timeProvider TimeProvider) *Retrier {
	return &Retrier{
		retryPolicy:   retryPolicy,
		delayStrategy: delayStrategy,
		timeProvider:  timeProvider,
	}
}

func DoWithRetry[T any](ctx context.Context, r *Retrier, doWork func() (T, error)) (T, error) {
	var empty T

	for i := 0; ; i++ {
		//проверяем контекст, чтобы не выполнять вызов функции, если контекст уже истек
		if ctx.Err() != nil {
			return empty, ctx.Err()
		}

		result, err := doWork()
		if err == nil {
			return result, nil
		}

		if r.retryPolicy.ShouldRetry(i, err) {
			r.timeProvider.Sleep(ctx, r.delayStrategy.GetDelay(i))
		} else {
			return empty, err
		}
	}
}
