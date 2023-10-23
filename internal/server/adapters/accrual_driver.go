package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/matthiasBT/gophermart/internal/infra/logging"
	"github.com/matthiasBT/gophermart/internal/server/entities"
)

type AccrualClient struct {
	logger            logging.ILogger
	baseURL           string
	retryAfterDefault int
	maxAttempts       int
}

func NewAccrualClient(logger logging.ILogger, url string, retryAfterDefault int, maxAttempts int) *AccrualClient {
	return &AccrualClient{
		logger:            logger,
		baseURL:           url,
		retryAfterDefault: retryAfterDefault,
		maxAttempts:       maxAttempts,
	}
}

func (ac *AccrualClient) GetAccrual(ctx context.Context, orderNumber string) (*entities.AccrualResponse, error) {
	ac.logger.Infof("Sending request for order accrual: %d", orderNumber)
	client := &http.Client{}
	req, err := ac.constructRequest(ctx, orderNumber)
	if err != nil {
		return nil, err
	}
	for i := 1; i <= ac.maxAttempts; i++ {
		ac.logger.Infof("Accrual system request: %s. Attempt: %d", req.URL.String(), i)
		resp, err := client.Do(req)
		if err != nil {
			ac.logger.Errorf("Request failed: %v", err.Error())
			return nil, err
		}
		if err := ac.checkResponseCode(resp); err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusOK {
			return ac.parseAccrualResponse(resp)
		}
		if resp.StatusCode == http.StatusNoContent {
			ac.logger.Infoln("Status no content")
			return nil, nil
		}
		timeout, err := ac.parseRetryAfterHeader(resp)
		if err != nil {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, errors.New("request aborted")
		case <-time.After(timeout):
			ac.logger.Infof("It's time to retry the request")
		}
	}
	ac.logger.Errorf("Failed to get data from the accrual system")
	return nil, errors.New("no accrual system response")
}

func (ac *AccrualClient) constructRequest(ctx context.Context, orderNumber string) (*http.Request, error) {
	path := fmt.Sprintf("%s%s/%s", ac.baseURL, "/api/orders", orderNumber)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		ac.logger.Errorf("Failed to construct a request: %s", err.Error())
		return nil, err
	}
	req = req.WithContext(ctx)
	return req, nil
}

func (ac *AccrualClient) checkResponseCode(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusTooManyRequests &&
		resp.StatusCode != http.StatusNoContent {
		ac.logger.Errorf("Non-OK and non-retriable response from the accrual system: %d", resp.StatusCode)
		return errors.New("accrual request failed")
	}
	return nil
}

func (ac *AccrualClient) parseAccrualResponse(resp *http.Response) (*entities.AccrualResponse, error) {
	ac.logger.Infoln("Status OK")
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
	ac.logger.Infof("Got accrual data for order: %v", accrual)
	return &accrual, nil
}

func (ac *AccrualClient) parseRetryAfterHeader(resp *http.Response) (time.Duration, error) {
	ac.logger.Infoln("Too many requests, need to wait for a while")
	var (
		retryAfterDuration int
		err                error
	)
	if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
		retryAfterDuration, err = strconv.Atoi(retryAfter)
		if err != nil {
			ac.logger.Errorf("Invalid Retry-After header value: %s", retryAfter)
			return 0, errors.New("invalid Retry-After header")
		}
	} else {
		retryAfterDuration = ac.retryAfterDefault
	}
	ac.logger.Warningf("Retrying after %d seconds", retryAfterDuration)
	return time.Duration(retryAfterDuration) * time.Second, nil
}
