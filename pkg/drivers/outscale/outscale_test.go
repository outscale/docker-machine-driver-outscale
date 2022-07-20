package outscale

import (
	"os"
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/stretchr/testify/assert"
)

func TestSetConfigFromEnv(t *testing.T) {
	os.Clearenv()
	driver := NewDriver("", "")

	os.Setenv("OSC_ACCESS_KEY", "OUTSCALE_ACCESS_KEY")
	os.Setenv("OSC_SECRET_KEY", "OUTSCALE_SECRET_KEY")
	os.Setenv("OSC_REGION", "OUTSCALE_REGION")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.NoError(t, err)
	assert.Equal(t, "OUTSCALE_ACCESS_KEY", driver.Ak)
	assert.Equal(t, "OUTSCALE_SECRET_KEY", driver.Sk)
	assert.Equal(t, "OUTSCALE_REGION", driver.Region)
	assert.Empty(t, checkFlags.InvalidFlags)
}

func TestSetConfigFromEnvWithConfigNotEmpty(t *testing.T) {
	os.Clearenv()
	driver := NewDriver("", "")

	os.Setenv("OSC_ACCESS_KEY", "OSC_ACCESS_KEY")
	os.Setenv("OSC_SECRET_KEY", "OSC_SECRET_KEY")
	os.Setenv("OSC_REGION", "OSC_REGION")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			flagAccessKey: "OUTSCALE_ACCESS_KEY",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.NoError(t, err)
	assert.Equal(t, "OUTSCALE_ACCESS_KEY", driver.Ak)
	assert.Equal(t, "OSC_SECRET_KEY", driver.Sk)
	assert.Equal(t, "OSC_REGION", driver.Region)
	assert.Empty(t, checkFlags.InvalidFlags)
}

func TestSetConfigDefaultRegion(t *testing.T) {
	os.Clearenv()
	driver := NewDriver("", "")

	os.Setenv("OSC_ACCESS_KEY", "OSC_ACCESS_KEY")
	os.Setenv("OSC_SECRET_KEY", "OSC_SECRET_KEY")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.NoError(t, err)
	assert.Equal(t, defaultOscRegion, driver.Region)
	assert.Empty(t, checkFlags.InvalidFlags)
}

func TestDiskType(t *testing.T) {
	assert.Equal(t, true, validateDiskType("io1"))
	assert.Equal(t, true, validateDiskType("standard"))
	assert.Equal(t, true, validateDiskType("gp2"))
	assert.Equal(t, false, validateDiskType("notADiskType"))
}

func TestDiskSize(t *testing.T) {
	os.Clearenv()
	driver := NewDriver("", "")

	os.Setenv("OSC_ACCESS_KEY", "OSC_ACCESS_KEY")
	os.Setenv("OSC_SECRET_KEY", "OSC_SECRET_KEY")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			flagRootDiskSize: -1,
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)
	assert.Error(t, err)
	assert.Equal(t, "the disk size (-1) is not accepted, it must be > 0", err.Error())

	checkFlags.FlagsValues[flagRootDiskSize] = 12
	err = driver.SetConfigFromFlags(checkFlags)
	assert.NoError(t, err)

}

func TestDiskIops(t *testing.T) {
	os.Clearenv()
	driver := NewDriver("", "")

	os.Setenv("OSC_ACCESS_KEY", "OSC_ACCESS_KEY")
	os.Setenv("OSC_SECRET_KEY", "OSC_SECRET_KEY")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			flagRootDiskIo1Iops: -1,
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)
	assert.Error(t, err)
	assert.Equal(t, "the disk iops (-1) is not accepted, it must between 1 and 13000", err.Error())

	checkFlags.FlagsValues[flagRootDiskIo1Iops] = 12
	err = driver.SetConfigFromFlags(checkFlags)
	assert.NoError(t, err)

}
