package main

import (
	"github.com/outscale-dev/docker-machine-driver-outscale/pkg/drivers/outscale"

	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(outscale.NewDriver("", ""))
}
