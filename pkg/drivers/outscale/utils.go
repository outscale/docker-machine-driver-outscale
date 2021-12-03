package outscale

import (
	"errors"
	"fmt"
	"log"
	"time"

	retry "github.com/avast/retry-go"
	osc "github.com/outscale/osc-sdk-go/v2"
)

const (
	defaultMaxAttempts = 180
	defaultDelay       = time.Duration(1) * time.Second
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
				fmt.Printf("Error while submitting the Vm creation request: ")
				if httpRes != nil {
					fmt.Printf(httpRes.Status)
				}
				return err
			}

			if !readVmResponse.HasVms() {
				return errors.New("Error while reading the VM: there is no VM")
			}

			if readVmResponse.GetVms()[0].GetState() != state {
				return errors.New("The VM is not (yet) in the wanted state")
			}
			return nil
		},
		retry.Attempts(defaultMaxAttempts),
		retry.Delay(defaultDelay),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("[DEBUG] Vm is not in the wanted state, retrying...")
		}),
	)

	if err != nil {
		return err
	}

	return nil
}
