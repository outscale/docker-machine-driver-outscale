package outscale

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	retry "github.com/avast/retry-go"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/ssh"
	osc "github.com/outscale/osc-sdk-go/v2"
)

// Create a SSH key for the VM
func (d *OscDriver) createSSHKey() (string, error) {
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return "", err
	}

	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return "", err
	}

	return string(publicKey), nil
}

// publicSSHKeyPath is always SSH Key Path appended with ".pub"
func (d *OscDriver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

// Create a Keypair for the VM
func createKeyPair(d *OscDriver) error {

	publicKey, err := d.createSSHKey()
	if err != nil {
		return nil
	}

	oscApi, err := d.getClient()
	if err != nil {
		return err
	}

	d.KeypairName = fmt.Sprintf("docker-machine-%s-%d", d.GetMachineName(), time.Now().Unix())

	request := osc.CreateKeypairRequest{
		KeypairName: d.KeypairName,
	}
	request.SetPublicKey(base64.StdEncoding.EncodeToString([]byte(publicKey)))

	var response osc.CreateKeypairResponse
	var httpRes *http.Response
	err = retry.Do(
		func() error {
			var response_error error
			response, httpRes, response_error = oscApi.client.KeypairApi.CreateKeypair(oscApi.context).CreateKeypairRequest(request).Execute()
			return wrapError(response_error, httpRes)
		},
		defaultThrottlingRetryOption...,
	)

	if err != nil {
		return fmt.Errorf("Error while submitting the Keypair creation request: %s", getErrorInfo(err, httpRes))
	}

	if !response.HasKeypair() {
		return errors.New("Error while creating the keypair: the response contains nothing")
	}

	return nil

}

func deleteKeyPair(d *OscDriver, keypairName string) error {
	oscApi, err := d.getClient()
	if err != nil {
		return err
	}

	if keypairName == "" {
		log.Warn("Skipping deletion of the keypair because none was stored.")
		return nil
	}

	request := osc.DeleteKeypairRequest{
		KeypairName: keypairName,
	}

	var httpRes *http.Response
	err = retry.Do(
		func() error {
			var response_error error
			_, httpRes, response_error = oscApi.client.KeypairApi.DeleteKeypair(oscApi.context).DeleteKeypairRequest(request).Execute()
			return wrapError(response_error, httpRes)
		},
		defaultThrottlingRetryOption...,
	)

	if err != nil {
		return fmt.Errorf("Error while submitting the Keypair deletetion request:  %s", getErrorInfo(err, httpRes))
	}

	return nil
}
