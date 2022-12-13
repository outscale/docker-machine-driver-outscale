package outscale

import (
	"fmt"
	"net/http"

	retry "github.com/avast/retry-go"
	osc "github.com/outscale/osc-sdk-go/v2"
)

func RetrieveNetFromSubnetId(d *OscDriver, subnetId string) (string, error) {
	oscApi, err := d.getClient()
	if err != nil {
		return "", err
	}

	request := osc.ReadSubnetsRequest{
		Filters: &osc.FiltersSubnet{
			SubnetIds: &[]string{subnetId},
		},
	}

	var httpRes *http.Response
	var response osc.ReadSubnetsResponse
	err = retry.Do(
		func() error {
			var response_error error
			response, httpRes, response_error = oscApi.client.SubnetApi.ReadSubnets(oscApi.context).ReadSubnetsRequest(request).Execute()
			return response_error
		},
		defaultThrottlingRetryOption...,
	)

	if err != nil {
		return "", fmt.Errorf("Error while submitting the Subnet read request: %s", getErrorInfo(err, httpRes))
	}

	if ! response.HasSubnets() {
		return "", fmt.Errorf("The subnet '%s' has not been found", subnetId)
	}

	subnet := response.GetSubnets()[0];

	return subnet.GetNetId(), nil
}