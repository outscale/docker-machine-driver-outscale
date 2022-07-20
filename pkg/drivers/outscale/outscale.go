package outscale

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	retry "github.com/avast/retry-go"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	osc "github.com/outscale/osc-sdk-go/v2"
)

const (
	defaultOscRegion       = "eu-west-2"
	defaultOscOMI          = "ami-504e6b16"  // Debian
	defaultOscVmType       = "tinav2.c1r2p3" //t2.small
	defaultDockerPort      = 2376
	defaultSSHPort         = 22
	defaultSSHUsername     = "outscale"
	defaultRootDiskType    = "gp2"
	defaultRootDiskSize    = 15
	defaultRootDiskIo1Iops = 1500

	flagAccessKey          = "outscale-access-key"
	flagSecretKey          = "outscale-secret-key"
	flagRegion             = "outscale-region"
	flagInstanceType       = "outscale-instance-type"
	flagSourceOmi          = "outscale-source-omi"
	flagExtraTagsAll       = "outscale-extra-tags-all"
	flagExtraTagsInstances = "outscale-extra-tags-instances"
	flagSecurityGroupIds   = "outscale-security-group-ids"
	flagRootDiskType       = "outscale-root-disk-type"
	flagRootDiskSize       = "outscale-root-disk-size"
	flagRootDiskIo1Iops    = "outscale-root-disk-io1-iops"
)

type OscDriver struct {
	*drivers.BaseDriver

	oscApi *OscApiData

	// Stored
	Ak     string
	Sk     string
	Region string

	VmId            string
	KeypairName     string
	SecurityGroupId string
	PublicIpId      string

	// Unstored
	instanceType       string
	sourceOmi          string
	extraTagsAll       []string
	extraTagsInstances []string
	securityGroupIds   []string
	rootDiskType       string
	rootDiskSize       int32
	rootDiskIo1Iops    int32
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
		config.UserAgent = fmt.Sprintf("docker-machine-driver-outscale/%s", GetVersion())

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
		cleanUp(d)
		return err
	}

	// Create a SG
	if d.securityGroupIds == nil {
		// Create default SG
		if err := createDefaultSecurityGroup(d); err != nil {
			cleanUp(d)
			return err
		}

		d.securityGroupIds = []string{d.SecurityGroupId}
	}

	// (TODO) Assign an Public IP
	if err := createPublicIp(d); err != nil {
		cleanUp(d)
		return err
	}
	// Create an Instance
	deviceName := "/dev/sda1"
	rootDisk := osc.BlockDeviceMappingVmCreation{
		Bsu: &osc.BsuToCreate{
			VolumeType: &d.rootDiskType,
			VolumeSize: &d.rootDiskSize,
		},
		DeviceName: &deviceName,
	}

	if d.rootDiskType == "io1" {
		rootDisk.Bsu.SetIops(d.rootDiskIo1Iops)
	}

	createVmRequest := osc.CreateVmsRequest{
		ImageId:          d.sourceOmi,
		KeypairName:      &d.KeypairName,
		VmType:           &d.instanceType,
		SecurityGroupIds: &d.securityGroupIds,
		BlockDeviceMappings: &[]osc.BlockDeviceMappingVmCreation{
			rootDisk,
		},
	}

	var createVmResponse osc.CreateVmsResponse
	var httpRes *http.Response
	err = retry.Do(
		func() error {
			var response_error error
			createVmResponse, httpRes, response_error = oscApi.client.VmApi.CreateVms(oscApi.context).CreateVmsRequest(createVmRequest).Execute()
			return response_error
		},
		defaultThrottlingRetryOption...,
	)

	if err != nil {
		log.Error("Error while submitting the Vm creation request: ")
		if httpRes != nil {
			fmt.Printf(httpRes.Status)
		}
		cleanUp(d)
		return err
	}

	if !createVmResponse.HasVms() || len(createVmResponse.GetVms()) != 1 {
		cleanUp(d)
		return errors.New("Error while creating the Vm: the number of VM created is wrong")
	}

	// Store the VM Id
	d.VmId = createVmResponse.GetVms()[0].GetVmId()

	// Wait for the VM to be started
	log.Debug("Waiting for the Vm to be running...")
	if err := d.waitForState(d.VmId, "running"); err != nil {
		cleanUp(d)
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

	var response osc.ReadVmsResponse
	err = retry.Do(
		func() error {
			var response_error error
			response, httpRes, response_error = oscApi.client.VmApi.ReadVms(oscApi.context).ReadVmsRequest(readVmRequest).Execute()
			return response_error
		},
		defaultThrottlingRetryOption...,
	)
	if err != nil {
		fmt.Printf("Error while submitting the Vm creation request: ")
		if httpRes != nil {
			fmt.Printf(httpRes.Status)
		}
		cleanUp(d)
		return err
	}

	if !response.HasVms() {
		cleanUp(d)
		return errors.New("Error while reading the VM: there is no VM")
	}

	// Link the Public Ip
	if err := linkPublicIp(d); err != nil {
		cleanUp(d)
		return err
	}

	// Add the tag of the Vm name
	if err := addTag(d, d.VmId, "name", d.GetMachineName()); err != nil {
		cleanUp(d)
		return err
	}

	// Add extra tags to the Instances
	if err := addExtraTags(d, d.VmId, d.extraTagsAll); err != nil {
		cleanUp(d)
		return err
	}

	// Add extra tags only for the Instances
	if err := addExtraTags(d, d.VmId, d.extraTagsInstances); err != nil {
		cleanUp(d)
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
			EnvVar: "OUTSCALE_ACCESS_KEY",
			Name:   flagAccessKey,
			Usage:  "Outscale Access Key",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OUTSCALE_SECRET_KEY",
			Name:   flagSecretKey,
			Usage:  "Outscale Secret Key",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OUTSCALE_REGION",
			Name:   flagRegion,
			Usage:  "Outscale Region (e.g. eu-west-2)",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OUTSCALE_INSTANCE_TYPE",
			Name:   flagInstanceType,
			Usage:  "VM Instance type",
			Value:  defaultOscVmType,
		},
		mcnflag.StringFlag{
			EnvVar: "OUTSCALE_SOURCE_OMI",
			Name:   flagSourceOmi,
			Usage:  "OMI to use as bootstrap",
			Value:  defaultOscOMI,
		},
		mcnflag.StringSliceFlag{
			EnvVar: "",
			Name:   flagExtraTagsAll,
			Usage:  "Tags to set at all created resources",
			Value:  nil,
		},
		mcnflag.StringSliceFlag{
			EnvVar: "",
			Name:   flagExtraTagsInstances,
			Usage:  "Tags to set only to instances <key1=value1,key2=value2>",
			Value:  nil,
		},
		mcnflag.StringSliceFlag{
			EnvVar: "",
			Name:   flagSecurityGroupIds,
			Usage:  "Add machine into theses security groups",
			Value:  nil,
		},
		mcnflag.StringFlag{
			EnvVar: "",
			Name:   flagRootDiskType,
			Usage:  "Type of volume for the root disk ('standard', 'io1' or 'gp2')",
			Value:  defaultRootDiskType,
		},
		mcnflag.IntFlag{
			EnvVar: "",
			Name:   flagRootDiskSize,
			Usage:  "Size of the root disk in GB (> 0)",
			Value:  defaultRootDiskSize,
		},
		mcnflag.IntFlag{
			EnvVar: "",
			Name:   flagRootDiskIo1Iops,
			Usage:  "Iops for the io1 root disk type (ignore if it is not io1). Value between 1 and 13000.",
			Value:  defaultRootDiskIo1Iops,
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

	var readVmResponse osc.ReadVmsResponse
	var httpRes *http.Response
	err = retry.Do(
		func() error {
			var response_error error
			readVmResponse, httpRes, response_error = oscApi.client.VmApi.ReadVms(oscApi.context).ReadVmsRequest(readVmRequest).Execute()
			return response_error
		},
		defaultThrottlingRetryOption...,
	)

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

	var httpRes *http.Response
	err = retry.Do(
		func() error {
			var response_error error
			_, httpRes, response_error = oscApi.client.AccountApi.ReadAccounts(oscApi.context).ReadAccountsRequest(request).Execute()
			return response_error
		},
		defaultThrottlingRetryOption...,
	)

	if err != nil {
		fmt.Printf("Error while submitting the ReadAccount request: ")
		if httpRes != nil {
			fmt.Printf(httpRes.Status)
		}
		return err
	}

	// Check the SG
	for _, sgId := range d.securityGroupIds {
		sgExist, sgError := isSecurityGroupExist(d, sgId)
		if sgError != nil {
			return nil
		}

		if !sgExist {
			return fmt.Errorf("The Security Group '%v' does not exist.", sgId)
		}

		log.Debugf("The Security Group '%v' exists.", sgId)
	}

	return nil

}

// Remove a host
func (d *OscDriver) Remove() error {
	oscApi, err := d.getClient()
	if err != nil {
		return err
	}

	if d.VmId != "" {
		request := osc.DeleteVmsRequest{
			VmIds: []string{
				d.VmId,
			},
		}

		var httpRes *http.Response
		err = retry.Do(
			func() error {
				var response_error error
				_, httpRes, response_error = oscApi.client.VmApi.DeleteVms(oscApi.context).DeleteVmsRequest(request).Execute()
				return response_error
			},
			defaultThrottlingRetryOption...,
		)

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
	} else {
		log.Warn("Skipping deletion of the VM because none was stored.")
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

	var httpRes *http.Response
	err = retry.Do(
		func() error {
			var response_error error
			_, httpRes, response_error = oscApi.client.VmApi.RebootVms(oscApi.context).RebootVmsRequest(request).Execute()
			return response_error
		},
		defaultThrottlingRetryOption...,
	)

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
	if d.Ak = flags.String(flagAccessKey); d.Ak == "" {
		if d.Ak = os.Getenv("OSC_ACCESS_KEY"); d.Ak == "" {
			return errors.New("Outscale Access Key is required")
		}
	}

	if d.Sk = flags.String(flagSecretKey); d.Sk == "" {
		if d.Sk = os.Getenv("OSC_SECRET_KEY"); d.Ak == "" {
			return errors.New("Outscale Secret key is required")
		}
	}

	if d.Region = flags.String(flagRegion); d.Region == "" {
		if d.Region = os.Getenv("OSC_REGION"); d.Region == "" {
			d.Region = defaultOscRegion
		}
	}

	d.instanceType = flags.String(flagInstanceType)
	d.sourceOmi = flags.String(flagSourceOmi)

	// Root disk
	d.rootDiskType = flags.String(flagRootDiskType)
	if !validateDiskType(d.rootDiskType) {
		return fmt.Errorf("the disk type is not accepted (got: %s, expected: 'standard'|'io1'|'gp2')", d.rootDiskType)
	}
	if d.rootDiskSize = int32(flags.Int(flagRootDiskSize)); d.rootDiskSize <= 0 {
		return fmt.Errorf("the disk size (%v) is not accepted, it must be > 0", d.rootDiskSize)
	}

	if d.rootDiskIo1Iops = int32(flags.Int(flagRootDiskIo1Iops)); d.rootDiskIo1Iops <= 0 {
		return fmt.Errorf("the disk iops (%v) is not accepted, it must between 1 and 13000", d.rootDiskIo1Iops)
	}

	// Tags
	if d.extraTagsAll = flags.StringSlice(flagExtraTagsAll); !validateExtraTagsFormat(d.extraTagsAll) {
		return fmt.Errorf("--%v have not the expected syntax", flagExtraTagsAll)
	}

	if d.extraTagsInstances = flags.StringSlice(flagExtraTagsInstances); !validateExtraTagsFormat(d.extraTagsInstances) {
		return fmt.Errorf("--%v have not the expected syntax", flagExtraTagsInstances)
	}

	// Security Groups
	d.securityGroupIds = flags.StringSlice(flagSecurityGroupIds)

	// SSH
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

	var httpRes *http.Response
	err = retry.Do(
		func() error {
			var response_error error
			_, httpRes, response_error = oscApi.client.VmApi.StartVms(oscApi.context).StartVmsRequest(request).Execute()
			return response_error
		},
		defaultThrottlingRetryOption...,
	)

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

	var httpRes *http.Response
	err = retry.Do(
		func() error {
			var response_error error
			_, httpRes, response_error = oscApi.client.VmApi.StopVms(oscApi.context).StopVmsRequest(request).Execute()
			return response_error
		},
		defaultThrottlingRetryOption...,
	)

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

func validateDiskType(diskType string) bool {
	switch diskType {
	case "io1", "gp2", "standard":
		return true
	default:
		return false
	}
}
