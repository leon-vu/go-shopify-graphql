package utils

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

func IsOperationUrlEmptyError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "Operation result URL is empty") || strings.Contains(err.Error(), "no Host in request URL"))
}

func IsInvalidTokenError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Invalid API key or access token")
}

func IsInvalidStorefrontTokenError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "401 Unauthorized body")
}

func IsMaxCostLimitError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "max cost limit")
}

func IsPermissionError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "403 Forbidden")
}

func ExecWithRetries(retryCount int, f func() error) error {
	var (
		retries = 0
		err     error
	)
	for {
		err = f()
		if IsInvalidTokenError(err) || IsInvalidStorefrontTokenError(err) || IsOperationUrlEmptyError(err) || IsPermissionError(err) ||
			IsMaxCostLimitError(err) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		} else if err != nil {
			retries++
			if retries > retryCount {
				return fmt.Errorf("after %v tries: %w", retries, err)
			}
			time.Sleep(time.Duration(retries) * time.Second)
			continue
		}
		break
	}
	return nil
}
