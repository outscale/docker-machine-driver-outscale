package outscale

import (
	"errors"
	"fmt"
	"net/http"

	retry "github.com/avast/retry-go"
	"github.com/docker/machine/libmachine/log"
	osc "github.com/outscale/osc-sdk-go/v2"
)

func createPublicIp(d *OscDriver) error {
	log.Debug("Creating the Public Ip")

	// Get the client
	oscApi, err := d.getClient()
	if err != nil {
		return err
	}

	request := osc.CreatePublicIpRequest{}

	var httpRes *http.Response
	var response osc.CreatePublicIpResponse
	err = retry.Do(
		func() error {
			var response_error error
			response, httpRes, response_error = oscApi.client.PublicIpApi.CreatePublicIp(oscApi.context).CreatePublicIpRequest(request).Execute()
			return response_error
		},
		defaultThrottlingRetryOption...,
	)

	if err != nil {
		log.Error("Error while submitting the Public IP creation request: ")
		if httpRes != nil {
			fmt.Printf(httpRes.Status)
		}
		return err
	}

	if !response.HasPublicIp() {
		return errors.New("Error  while creating public Ip ")
	}

	d.IPAddress = response.PublicIp.GetPublicIp()
	d.PublicIpId = response.PublicIp.GetPublicIpId()

	return nil
}

func linkPublicIp(d *OscDriver) error {
	log.Debug("Linking the Public Ip")

	// Get the client
	oscApi, err := d.getClient()
	if err != nil {
		return err
	}

	request := osc.LinkPublicIpRequest{
		PublicIpId: &d.PublicIpId,
		VmId:       &d.VmId,
	}

	var httpRes *http.Response
	var response osc.LinkPublicIpResponse
	err = retry.Do(
		func() error {
			var response_error error
			response, httpRes, response_error = oscApi.client.PublicIpApi.LinkPublicIp(oscApi.context).LinkPublicIpRequest(request).Execute()
			return response_error
		},
		defaultThrottlingRetryOption...,
	)

	if err != nil {
		log.Error("Error while submitting the Public IP link request: ")
		if httpRes != nil {
			fmt.Printf(httpRes.Status)
		}
		return err
	}

	if !response.HasLinkPublicIpId() {
		return errors.New("Error  while creating public Ip ")
	}

	if err := addTag(d, d.VmId, "osc.fcu.eip.auto-attach", d.IPAddress); err != nil {
		return err
	}

	return nil
}

func deletePublicIp(d *OscDriver, resourceId string) error {
	log.Debug("Deletion of the Public Ip")

	// Get the client
	oscApi, err := d.getClient()
	if err != nil {
		return err
	}

	if resourceId == "" {
		log.Warn("Skipping deletion of the public IP because none was stored.")
		return nil
	}

	request := osc.DeletePublicIpRequest{
		PublicIpId: &resourceId,
	}

	var httpRes *http.Response
	err = retry.Do(
		func() error {
			var response_error error
			_, httpRes, response_error = oscApi.client.PublicIpApi.DeletePublicIp(oscApi.context).DeletePublicIpRequest(request).Execute()
			return response_error
		},
		defaultThrottlingRetryOption...,
	)

	if err != nil {
		log.Error("Error while submitting the Public IP link deletion request: ")
		if httpRes != nil {
			fmt.Printf(httpRes.Status)
		}
		return err
	}

	return nil

}
