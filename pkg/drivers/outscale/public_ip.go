package outscale

import (
	"errors"
	"fmt"

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

	response, httpRes, err := oscApi.client.PublicIpApi.CreatePublicIp(oscApi.context).CreatePublicIpRequest(request).Execute()
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

	response, httpRes, err := oscApi.client.PublicIpApi.LinkPublicIp(oscApi.context).LinkPublicIpRequest(request).Execute()
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
