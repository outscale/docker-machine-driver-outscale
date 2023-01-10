package outscale

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	retry "github.com/avast/retry-go"
	"github.com/docker/machine/libmachine/log"
	osc "github.com/outscale/osc-sdk-go/v2"
)

const (
	defaultReadMaxAttempts       = 180
	defaultReadDelay             = time.Duration(1) * time.Second
	defaultThrottlingDelay       = time.Duration(15) * time.Second
	defaultThrottlingMaxAttempts = 60
)

var (
	defaultThrottlingRetryOption = []retry.Option{
		retry.MaxJitter(defaultThrottlingDelay),
		retry.DelayType(retry.RandomDelay),
		retry.Attempts(defaultThrottlingMaxAttempts),
		retry.OnRetry(func(n uint, err error) {
			log.Debug("Retry number %v after throttling.", n)
		}),
		retry.RetryIf(isThrottlingError),
		retry.LastErrorOnly(true),
	}

	// Throtlling
	ThrottlingErrors = []int{503, 429}
)

func (d *OscDriver) waitForState(vmId string, state string) error {
	err := retry.Do(
		func() error {
			oscApi, err := d.getClient()
			if err != nil {
				return err
			}

			readVmRequest := osc.ReadVmsRequest{
				Filters: &osc.FiltersVm{
					VmIds: &[]string{
						vmId,
					},
				},
			}

			readVmResponse, httpRes, err := oscApi.client.VmApi.ReadVms(oscApi.context).ReadVmsRequest(readVmRequest).Execute()
			if err != nil {
				return fmt.Errorf("Error while submitting the Vm read request: %s", getErrorInfo(err, httpRes))
			}

			if !readVmResponse.HasVms() {
				return errors.New("Error while reading the VM: there is no VM")
			}

			if readVmResponse.GetVms()[0].GetState() != state {
				return errors.New("The VM is not (yet) in the wanted state")
			}
			return nil
		},
		retry.Attempts(defaultReadMaxAttempts),
		retry.Delay(defaultReadDelay),
		retry.OnRetry(func(n uint, err error) {
			log.Debug("Vm is not in the wanted state, retrying...")
		}),
	)

	if err != nil {
		return err
	}

	return nil
}

func isThrottlingError(err error) bool {
	cloudError, ok := err.(CloudError)
	if ! ok {
		return false
	}
	if cloudError.httpRes != nil {
		for _, errorCode := range ThrottlingErrors {
			if errorCode == cloudError.httpRes.StatusCode {
				return true
			}
		}
	}
	return false
}

func cleanUp(d *OscDriver) {
	d.Remove()
}

func extractApiError(err error) (bool, *osc.ErrorResponse) {
	genericError, ok := err.(osc.GenericOpenAPIError)
	if ok {
		errorsResponse, ok := genericError.Model().(osc.ErrorResponse)
		if ok {
			return true, &errorsResponse
		}
		return false, nil
	}
	return false, nil
}

func getErrorInfo(err error, httpRes *http.Response) string {
	if ok, apiError := extractApiError(err); ok {
		return fmt.Sprintf("%v - '%v %v' - '%v'", httpRes.Status, apiError.GetErrors()[0].GetCode(), apiError.GetErrors()[0].GetType(), apiError.GetErrors()[0].GetDetails())
	}
	if httpRes != nil {
		return httpRes.Status
	}

	return fmt.Sprintf("%v", err)
}

type CloudError struct {
	httpRes *http.Response
	cloudError error
}

func (e CloudError) Error() string {
    return getErrorInfo(e.cloudError, e.httpRes)
}

func wrapError(err error, httpRes *http.Response) error {
	if err == nil {
		return nil
	}

	return CloudError{
		httpRes: httpRes,
		cloudError: err,
	}
}
