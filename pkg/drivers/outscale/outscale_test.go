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
