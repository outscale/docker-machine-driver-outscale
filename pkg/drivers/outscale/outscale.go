package outscale

import (
	"context"
	"errors"
	"fmt"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	osc "github.com/outscale/osc-sdk-go/v2"
)

const (
	defaultOscRegion   = "eu-west-2"
	defaultOScOMI      = "ami-504e6b16" // Debian
	defaultDockerPort  = 2376
	defaultSSHPort     = 22
	defaultSSHUsername = "outscale"
	defaultKeyPairPath = "/tmp/keypair"
)

type OscDriver struct {
	*drivers.BaseDriver

	oscApi *OscApiData

	Ak     string
	Sk     string
	Region string

	VmId            string
	KeypairName     string
	SecurityGroupId string
	PublicIpId      string
}

type OscApiData struct {
	client  *osc.APIClient
	context context.Context
}

// NewDriver creates and returns a new instance of the Outscale driver
func NewDriver(hostName, storePath string) *OscDriver {
	return &OscDriver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

func (d *OscDriver) getClient() (*OscApiData, error) {
	if d.oscApi == nil {
		config := osc.NewConfiguration()

		config.Debug = true
		client := osc.NewAPIClient(config)

		ctx := context.WithValue(context.Background(), osc.ContextAWSv4, osc.AWSv4{
			AccessKey: d.Ak,
			SecretKey: d.Sk,
		})

		ctx = context.WithValue(ctx, osc.ContextServerIndex, 0)
		ctx = context.WithValue(ctx, osc.ContextServerVariables, map[string]string{"region": d.Region})

		d.oscApi = &OscApiData{
			client:  client,
			context: ctx,
		}
	}

	return d.oscApi, nil

}

// Create a host using the driver's config
func (d *OscDriver) Create() error {
	log.Debug("Creating a Vm")

	// Get the client
	oscApi, err := d.getClient()
	if err != nil {
		return err
	}

	// Create a keypair
	if err := createKeyPair(d); err != nil {
		return err
	}

	// Create a SG
	if err := createSecurityGroup(d); err != nil {
		return err
	}

	// (TODO) Assign an Public IP
	if err := createPublicIp(d); err != nil {
		return err
	}
	// Create an Instance
	createVmRequest := osc.CreateVmsRequest{
		ImageId:     defaultOScOMI,
		KeypairName: &d.KeypairName,
		SecurityGroupIds: &[]string{
			d.SecurityGroupId,
		},
	}

	createVmResponse, httpRes, err := oscApi.client.VmApi.CreateVms(oscApi.context).CreateVmsRequest(createVmRequest).Execute()
	if err != nil {
		log.Error("Error while submitting the Vm creation request: ")
		if httpRes != nil {
			fmt.Printf(httpRes.Status)
		}
		return err
	}

	if !createVmResponse.HasVms() || len(createVmResponse.GetVms()) != 1 {
		return errors.New("Error while creating the Vm: the number of VM created is wrong")
	}

	// Store the VM Id
	d.VmId = createVmResponse.GetVms()[0].GetVmId()

	// Wait for the VM to be started
	log.Debug("Waiting for the Vm to be running...")
	if err := d.waitForState(d.VmId, "running"); err != nil {
		return errors.New("Error while waiting that the VM is running")
	}

	// Retrieve the Public IP
	readVmRequest := osc.ReadVmsRequest{
		Filters: &osc.FiltersVm{
			VmIds: &[]string{
				d.VmId,
			},
		},
	}

	response, httpRes, err := oscApi.client.VmApi.ReadVms(oscApi.context).ReadVmsRequest(readVmRequest).Execute()
	if err != nil {
		fmt.Printf("Error while submitting the Vm creation request: ")
		if httpRes != nil {
			fmt.Printf(httpRes.Status)
		}
		return err
	}

	if !response.HasVms() {
		return errors.New("Error while reading the VM: there is no VM")
	}

	// Link the Public Ip
	if err := linkPublicIp(d); err != nil {
		return err
	}

	// Add the tag of the Vm name
	if err := addTag(d, d.VmId, "name", d.GetMachineName()); err != nil {
		return err
	}

	return nil
}

// DriverName returns the name of the driver
func (d *OscDriver) DriverName() string {
	return "outscale"
}

// GetCreateFlags returns the mcnflag.Flag slice representing the flags
// that can be set, their descriptions and defaults.
func (d *OscDriver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "OUTSCALE_ACCESSKEYID",
			Name:   "outscale-access-key",
			Usage:  "Outscale Access Key",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OUTSCALE_SECRETKEYID",
			Name:   "outscale-secret-key",
			Usage:  "Outscale Secret Key",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OUTSCALE_REGION",
			Name:   "outscale-region",
			Usage:  "Outscale Region (e.g. eu-west-2)",
			Value:  defaultOscRegion,
		},
	}
}

// GetIP returns an IP or hostname that this host is available at
// e.g. 1.2.3.4 or docker-host-d60b70a14d3a.cloudapp.net
func (d *OscDriver) GetIP() (string, error) {
	if d.IPAddress == "" {
		return "", errors.New("IP address is not set")
	}
	return d.IPAddress, nil
}

func (d *OscDriver) GetSSHHostname() (string, error) {
	return d.IPAddress, nil
}

// GetSSHKeyPath returns key path for use with ssh
func (d *OscDriver) GetSSHKeyPath() string {
	return fmt.Sprintf("/tmp/%s", d.GetMachineName())
}

// GetSSHPort returns port for use with ssh
func (d *OscDriver) GetSSHPort() (int, error) {
	return defaultSSHPort, nil
}

// GetSSHUsername returns username for use with ssh
func (d *OscDriver) GetSSHUsername() string {
	return defaultSSHUsername
}

// GetURL returns a Docker compatible host URL for connecting to this host
// e.g. tcp://1.2.3.4:2376
func (d *OscDriver) GetURL() (string, error) {
	return fmt.Sprintf("tcp://%s:%d", d.IPAddress, defaultDockerPort), nil
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *OscDriver) GetState() (state.State, error) {
	oscApi, err := d.getClient()
	if err != nil {
		return state.None, err
	}

	readVmRequest := osc.ReadVmsRequest{
		Filters: &osc.FiltersVm{
			VmIds: &[]string{
				d.VmId,
			},
		},
	}

	readVmResponse, httpRes, err := oscApi.client.VmApi.ReadVms(oscApi.context).ReadVmsRequest(readVmRequest).Execute()
	if err != nil {
		fmt.Printf("Error while submitting the Vm creation request: ")
		if httpRes != nil {
			fmt.Printf(httpRes.Status)
		}
		return state.None, err
	}

	if !readVmResponse.HasVms() {
		return state.None, errors.New("Error while reading the VM: there is no VM")
	}

	switch vmState := readVmResponse.GetVms()[0].GetState(); vmState {
	case "pending":
		return state.Starting, nil
	case "running":
		return state.Running, nil
	case "stopping", "shutting-down":
		return state.Stopping, nil
	case "stopped", "terminated", "quarantine":
		return state.Stopped, nil
	default:
		return state.None, nil
	}
}

// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
func (d *OscDriver) PreCreateCheck() error {
	oscApi, err := d.getClient()
	if err != nil {
		return err
	}

	request := osc.ReadAccountsRequest{}

	_, httpRes, err := oscApi.client.AccountApi.ReadAccounts(oscApi.context).ReadAccountsRequest(request).Execute()
	if err != nil {
		fmt.Printf("Error while submitting the ReadAcoount request: ")
		if httpRes != nil {
			fmt.Printf(httpRes.Status)
		}
		return err
	}

	return nil

}

// Remove a host
func (d *OscDriver) Remove() error {
	oscApi, err := d.getClient()
	if err != nil {
		return err
	}

	request := osc.DeleteVmsRequest{
		VmIds: []string{
			d.VmId,
		},
	}

	_, httpRes, err := oscApi.client.VmApi.DeleteVms(oscApi.context).DeleteVmsRequest(request).Execute()
	if err != nil {
		fmt.Printf("Error while submitting the DeleteVm request: ")
		if httpRes != nil {
			fmt.Printf(httpRes.Status)
		}
		return err
	}

	if err := d.waitForState(d.VmId, "terminated"); err != nil {
		return err
	}

	if err := deletePublicIp(d, d.PublicIpId); err != nil {
		return err
	}

	if err := deleteSecurityGroup(d, d.SecurityGroupId); err != nil {
		return err
	}

	if err := deleteKeyPair(d, d.KeypairName); err != nil {
		return err
	}

	return nil
}

// Restart a host. This may just call Stop(); Start() if the provider does not
// have any special restart behaviour.
func (d *OscDriver) Restart() error {
	oscApi, err := d.getClient()
	if err != nil {
		return err
	}

	request := osc.RebootVmsRequest{
		VmIds: []string{
			d.VmId,
		},
	}

	_, httpRes, err := oscApi.client.VmApi.RebootVms(oscApi.context).RebootVmsRequest(request).Execute()
	if err != nil {
		fmt.Printf("Error while submitting the RebootVm request: ")
		if httpRes != nil {
			fmt.Printf(httpRes.Status)
		}
		return err
	}

	if err := d.waitForState(d.VmId, "running"); err != nil {
		return err
	}

	return nil
}

// SetConfigFromFlags configures the driver with the object that was returned
// by RegisterCreateFlags
func (d *OscDriver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	if d.Ak = flags.String("outscale-access-key"); d.Ak == "" {
		return errors.New("Outscale Access Key is required")
	}

	if d.Sk = flags.String("outscale-secret-key"); d.Sk == "" {
		return errors.New("Outscale Secret key is required")
	}

	d.Region = flags.String("outscale-region")

	d.SSHKeyPath = d.GetSSHKeyPath()
	d.SSHUser = d.GetSSHUsername()
	d.SSHPort, _ = d.GetSSHPort()

	return nil
}

// Start a host
func (d *OscDriver) Start() error {
	oscApi, err := d.getClient()
	if err != nil {
		return err
	}

	request := osc.StartVmsRequest{
		VmIds: []string{
			d.VmId,
		},
	}

	_, httpRes, err := oscApi.client.VmApi.StartVms(oscApi.context).StartVmsRequest(request).Execute()
	if err != nil {
		fmt.Printf("Error while submitting the StartVm request: ")
		if httpRes != nil {
			fmt.Printf(httpRes.Status)
		}
		return err
	}

	if err := d.waitForState(d.VmId, "running"); err != nil {
		return err
	}

	return nil

}

func (d *OscDriver) innerStop(force bool) error {
	oscApi, err := d.getClient()
	if err != nil {
		return err
	}

	request := osc.StopVmsRequest{
		VmIds: []string{
			d.VmId,
		},
	}
	request.SetForceStop(force)

	_, httpRes, err := oscApi.client.VmApi.StopVms(oscApi.context).StopVmsRequest(request).Execute()
	if err != nil {
		fmt.Printf("Error while submitting the StopVm request: ")
		if httpRes != nil {
			fmt.Printf(httpRes.Status)
		}
		return err
	}

	if err := d.waitForState(d.VmId, "stopped"); err != nil {
		return err
	}

	return nil
}

// Stop a host gracefully
func (d *OscDriver) Stop() error {
	return d.innerStop(false)
}

// Kill stops a host forcefully
func (d *OscDriver) Kill() error {
	return d.innerStop(true)
}
