# docker-machine-driver-outscale

[![Go Report Card](https://goreportcard.com/badge/github.com/outscale-mdr/docker-machine-driver-outscale)](https://goreportcard.com/report/github.com/outscale-mdr/docker-machine-driver-outscale)
[![GitHub release](https://img.shields.io/github/release/outscale-mdr/docker-machine-driver-outscale.svg)](https://github.com/outscale-mdr/docker-machine-driver-outscale/releases/)

Outscale Driver plugin for docker-machine

## TODO
- [ ] Test CRUD operations
- [ ] Handle operation failure (remove created resources)
- [ ] Allow private IP / VPC
- [ ] Handle additional parameters (KeyPair, VPC, SG)

## Install
If you would rather build from source, you will need to have a working `go` 1.17+ environment,

```bash
eval $(go env)
export PATH="$PATH:$GOPATH/bin"
```

You can then install `docker-machine` from source by running:

```bash
go get github.com/docker/machine
cd $GOPATH/src/github.com/docker/machine
make build
```

And then compile the `docker-machine-driver-outscale` driver:

```bash
go get github.com/outscale-mdr/docker-machine-driver-outscale
cd $GOPATH/src/github.com/outscale-mdr/docker-machine-driver-outscale
make install
```

## Run
In order to create a machine, you will need to have you AK/SK . You can find information here: [Documentation](https://docs.outscale.com/en/userguide/Getting-Information-About-Your-Access-Keys.html)

```bash
docker-machine create -d outscale --outscale-access-key=<outscale-access-key>  --outscale-secret-key=<outscale-secret-key> --outscale-region=<outscale-region> outscale
```

### Options
| Argument | Env | Default | Description
| --- | --- | --- | ---
| `outscale-access-key` | `OSC_ACCESS_KEY` | None | **required** Outscale Access Key (see [here](https://docs.outscale.com/en/userguide/Getting-Information-About-Your-Access-Keys.html))
| `outscale-secret-key` | `OSC_SECRET_KEY` | None | **required** Outscale Secret Key (see [here](https://docs.outscale.com/en/userguide/Getting-Information-About-Your-Access-Keys.html))
| `outscale-region` | `OSC_REGION` | eu-west-2 | Outscale Region
| `outscale-instance-type` | `OSC_INSTANCE_TYPE` | tinav2.c1r2p3 (t2.small) | Outscale VM Instance Type (see [here](https://docs.outscale.com/en/userguide/Instance-Types.html))
| `outscale-source-omi`    | `OSC_SOURCE_OMI`    | ami-504e6b16 (Debian-10-2021.05.12-3) | Outscale Machine Image to use as bootstrap for the VM (see [here](https://docs.outscale.com/en/userguide/Official-OMIs-Reference.html#_supported_official_images)) |


## Debugging
Detailed run output will be emitted when using  the `docker-machine` `--debug` option.

```bash
docker-machine --debug  create -d outscale --outscale-access-key=<outscale-access-key>  --outscale-secret-key=<outscale-secret-key> --outscale-region=<outscale-region> outscale
```

## License

> Copyright Outscale SAS
>
> BSD-3-Clause

This project is compliant with [REUSE](https://reuse.software/).