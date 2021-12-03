package outscale

import (
	"context"
	"errors"
	"fmt"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	osc "github.com/outscale/osc-sdk-go/v2"
)

const (
	defaultOscRegion = "eu-west-2"
	defaultOScOMI    = "ami-504e6b16" // Debian
)

type OscDriver struct {
	*drivers.BaseDriver

	oscApi *OscApiData

	ak     string
	sk     string
	region string

	vmId string
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
			AccessKey: d.ak,
			SecretKey: d.sk,
		})

		ctx = context.WithValue(ctx, osc.ContextServerIndex, 0)
		ctx = context.WithValue(ctx, osc.ContextServerVariables, map[string]string{"region": d.region})

		d.oscApi = &OscApiData{
			client:  client,
			context: ctx,
		}
	}

	return d.oscApi, nil

}

// Create a host using the driver's config
func (d *OscDriver) Create() error {

	// Get the client
	oscApi, err := d.getClient()
	if err != nil {
		return err
	}

	// (TODO) Create a keypair

	// (TODO) Create a SG

	// (TODO) Assign an Public IP

	// Create an Instance
	createVmRequest := osc.CreateVmsRequest{
		ImageId: defaultOScOMI,
	}

	createVmResponse, httpRes, err := oscApi.client.VmApi.CreateVms(oscApi.context).CreateVmsRequest(createVmRequest).Execute()
	if err != nil {
		fmt.Printf("Error while submitting the Vm creation request: ")
		if httpRes != nil {
			fmt.Printf(httpRes.Status)
		}
		return err
	}

	if !createVmResponse.HasVms() || len(createVmResponse.GetVms()) != 1 {
		return errors.New("Error while creating the Vm: the number of VM created is wrong")
	}

	// Store the VM Id
	d.vmId = createVmResponse.GetVms()[0].GetVmId()

	// Wait for the VM to be started
	fmt.Println("Waiting for the Vm to be running...")
	if err := d.waitForState(d.vmId, "running"); err != nil {
		return errors.New("Error while waiting that the VM is running")
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
			EnvVar: "OSC_ACCESS_KEY",
			Name:   "osc-access-key",
			Usage:  "Outscale Access Key",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OSC_SECRET_KEY",
			Name:   "osc-secret-key",
			Usage:  "Outscale Secret Key",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OSC_REGION",
			Name:   "osc-region",
			Usage:  "Outscale Region (e.g. eu-west-2)",
			Value:  defaultOscRegion,
		},
	}
}

// GetMachineName returns the name of the machine
func (d *OscDriver) GetMachineName() string {
	return d.vmId
}

// GetIP returns an IP or hostname that this host is available at
// e.g. 1.2.3.4 or docker-host-d60b70a14d3a.cloudapp.net
func (d *OscDriver) GetSSHHostname() (string, error) {
	return "", nil
}

// GetSSHKeyPath returns key path for use with ssh
func (d *OscDriver) GetSSHKeyPath() string {
	return ""
}

// GetSSHPort returns port for use with ssh
func (d *OscDriver) GetSSHPort() (int, error) {
	return -1, nil
}

// GetSSHUsername returns username for use with ssh
func (d *OscDriver) GetSSHUsername() string {
	return ""
}

// GetURL returns a Docker compatible host URL for connecting to this host
// e.g. tcp://1.2.3.4:2376
func (d *OscDriver) GetURL() (string, error) {
	return "", nil
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *OscDriver) GetState() (state.State, error) {
	return state.None, nil
}

// Kill stops a host forcefully
func (d *OscDriver) Kill() error {
	return nil
}

// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
func (d *OscDriver) PreCreateCheck() error {
	return nil
}

// Remove a host
func (d *OscDriver) Remove() error {
	return nil
}

// Restart a host. This may just call Stop(); Start() if the provider does not
// have any special restart behaviour.
func (d *OscDriver) Restart() error {
	return nil
}

// SetConfigFromFlags configures the driver with the object that was returned
// by RegisterCreateFlags
func (d *OscDriver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	if d.ak = flags.String("osc-access-key"); d.ak == "" {
		return errors.New("Outscale Access Key is required")
	}

	if d.sk = flags.String("osc-secret-key"); d.sk == "" {
		return errors.New("Outscale Secret key is required")
	}

	d.region = flags.String("osc-region")

	return nil
}

// Start a host
func (d *OscDriver) Start() error {
	return nil
}

// Stop a host gracefully
func (d *OscDriver) Stop() error {
	return nil
}
