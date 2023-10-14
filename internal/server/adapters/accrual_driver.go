package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/matthiasBT/gophermart/internal/infra/logging"
	"github.com/matthiasBT/gophermart/internal/server/entities"
)

type AccrualClient struct {
	logger            logging.ILogger
	baseURL           string
	lock              *sync.Mutex
	retryAfterDefault int
	maxAttempts       int
}

func NewAccrualClient(logger logging.ILogger, url string, retryAfterDefault int, maxAttempts int) *AccrualClient {
	return &AccrualClient{
		logger:            logger,
		baseURL:           url,
		lock:              &sync.Mutex{},
		retryAfterDefault: retryAfterDefault,
		maxAttempts:       maxAttempts,
	}
}

func (ac *AccrualClient) GetAccrual(ctx context.Context, orderID int) (*entities.AccrualResponse, error) {
	ac.logger.Infof("Sending request for order accrual: %d", orderID)
	if err := ac.Lock(ctx); err != nil {
		return nil, errors.New("mutex locking was cancelled")
	}
	defer ac.lock.Unlock()
	client := &http.Client{}
	req, err := ac.constructRequest(ctx, orderID)
	if err != nil {
		return nil, err
	}
	for i := 0; i < ac.maxAttempts; i++ {
		resp, err := client.Do(req)
		if err != nil {
			ac.logger.Errorf("Request failed: %v", err.Error())
			return nil, err
		}
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusTooManyRequests {
			ac.logger.Errorf("Non-OK and non-retriable response from the accrual system: %d", resp.StatusCode)
			return nil, errors.New("accrual request failed")
		}
		if resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				ac.logger.Errorf("Failed to read response body: %s", err.Error())
				return nil, err
			}
			var accrual entities.AccrualResponse
			if err := json.Unmarshal(body, &accrual); err != nil {
				ac.logger.Errorf("Failed to parse response: %s", err.Error())
				return nil, err
			}
			return &accrual, nil
		}
		var retryAfterDuration int
		if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
			retryAfterDuration, err = strconv.Atoi(retryAfter)
			if err != nil {
				ac.logger.Errorf("Invalid Retry-After header value: %s", retryAfter)
				return nil, errors.New("invalid Retry-After header")
			}
		} else {
			retryAfterDuration = ac.retryAfterDefault
		}
		timeout := time.Duration(retryAfterDuration) * time.Second
		ac.logger.Warningf("Retrying after %d seconds", retryAfterDuration)
		select {
		case <-ctx.Done():
			return nil, errors.New("request aborted")
		case <-time.After(timeout):
			ac.logger.Infof("It's time to retry the request")
		}
	}
	return nil, errors.New("unreachable code")
}

func (ac *AccrualClient) constructRequest(ctx context.Context, orderID int) (*http.Request, error) {
	path := fmt.Sprintf("%s%s/%d", ac.baseURL, "/api/orders", orderID)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		ac.logger.Errorf("Failed to construct a request: %s", err.Error())
		return nil, err
	}
	req = req.WithContext(ctx)
	return req, nil
}

func (ac *AccrualClient) Lock(ctx context.Context) error {
	lockAcquired := make(chan struct{})
	go func() {
		ac.lock.Lock()
		close(lockAcquired)
	}()
	select {
	case <-lockAcquired:
		ac.logger.Infoln("Lock acquired")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}