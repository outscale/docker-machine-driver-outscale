package outscale

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

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

	response, httpRes, err := oscApi.client.KeypairApi.CreateKeypair(oscApi.context).CreateKeypairRequest(request).Execute()
	if err != nil {
		log.Error("Error while submitting the Keypair creation request: ")
		if httpRes != nil {
			fmt.Printf(httpRes.Status)
		}
		return err
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

	request := osc.DeleteKeypairRequest{
		KeypairName: keypairName,
	}

	_, httpRes, err := oscApi.client.KeypairApi.DeleteKeypair(oscApi.context).DeleteKeypairRequest(request).Execute()
	if err != nil {
		log.Error("Error while submitting the Keypair deletetion request: ")
		if httpRes != nil {
			fmt.Printf(httpRes.Status)
		}
		return err
	}

	return nil
}
