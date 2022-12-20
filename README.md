# docker-machine-driver-outscale

[![Go Report Card](https://goreportcard.com/badge/github.com/outscale-dev/docker-machine-driver-outscale)](https://goreportcard.com/report/github.com/outscale-dev/docker-machine-driver-outscale)
[![GitHub release](https://img.shields.io/github/release/outscale-dev/docker-machine-driver-outscale.svg)](https://github.com/outscale-dev/docker-machine-driver-outscale/releases/)

Outscale Driver plugin for docker-machine

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
go get github.com/outscale-dev/docker-machine-driver-outscale
cd $GOPATH/src/github.com/outscale-dev/docker-machine-driver-outscale
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
| `outscale-access-key` | `OUTSCALE_ACCESS_KEY\|OSC_ACCESS_KEY` | None | **required** Outscale Access Key (see [here](https://docs.outscale.com/en/userguide/Getting-Information-About-Your-Access-Keys.html))
| `outscale-secret-key` | `OUTSCALE_SECRET_KEY\|OSC_SECRET_KEY` | None | **required** Outscale Secret Key (see [here](https://docs.outscale.com/en/userguide/Getting-Information-About-Your-Access-Keys.html))
| `outscale-region` | `OUTSCALE_REGION\|OSC_REGION` | eu-west-2 | Outscale Region
| `outscale-instance-type` | `OUTSCALE_INSTANCE_TYPE` | tinav2.c1r2p3 (t2.small) | Outscale VM Instance Type (see [here](https://docs.outscale.com/en/userguide/Instance-Types.html))
| `outscale-source-omi`    | `OUTSCALE_SOURCE_OMI`    | ami-2cf1fa3e (Debian-10-2021.05.12-3) | Outscale Machine Image to use as bootstrap for the VM (see [here](https://docs.outscale.com/en/userguide/Official-OMIs-Reference.html#_supported_official_images)) |
| `outscale-extra-tags-all` | `` | nil| Extra tags for all created resources. Format "key=value". Can be set multiple times
| `outscale-extra-tags-instances` | `` | nil | Extra tags only for instances. Format "key=value". Can be set multiple times
| `outscale-security-group-ids` | `` | nil | Ids of user defined Security Groups to add to the machine. Can be set multiple times
| `outscale-root-disk-type` | `` | gp2 | Type of volume for the root disk ('standard', 'io1' or 'gp2')
| `outscale-root-disk-size` | `` | 15 | Size of the root disk in GB (> 0)
| `outscale-root-disk-iops` | `` | 1500 | Iops for the io1 root disk type (ignore if it is not io1). Value between 1 and 13000.
| `outscale-subnet-id` | `` | `` | Id of the Net use to create all resources when a private network is requested.


## Security group
If no Security group is provided, a security group will be created with theses rules
| Type | Protocol | From Port | To Port | CIDR | Description
| --- | --- | --- | --- | --- | ---
| Inbound | TCP | 22 | 22 | 0.0.0.0/0 | SSH
| Inbound | TCP | 80 | 80 | 0.0.0.0/0 | nginx Ingress Http
| Inbound | TCP | 443 | 443 | 0.0.0.0/0 | nginx Ingress Https
| Inbound | TCP | 2376 | 2376 | 0.0.0.0/0 | Docker daemon
| Inbound | TCP | 2379 | 2380 | 0.0.0.0/0 | ETCD (client request and peer communication)
| Inbound | TCP | 6443 | 6443 | 0.0.0.0/0 | Kube-api 
| Inbound | TCP | 10250 | 10250 | 0.0.0.0/0 | Kubelet
| Inbound | TCP | 10251 | 10251 | 0.0.0.0/0 | Kube-scheduler
| Inbound | TCP | 10252 | 10252 | 0.0.0.0/0 | Kube-controller-manager
| Inbound | TCP | 10256 | 10256 | 0.0.0.0/0 | Kube-proxy
| Inbound | TCP | 30000 | 32767 | 0.0.0.0/0 | Node port
| Inbound | UDP | 30000 | 32767 | 0.0.0.0/0 | Node port
| Inbound | UDP | 8472 | 8472 | 0.0.0.0/0 | Canal/Flannel overlay
| Inbound | UDP | 4789 | 4789 | 0.0.0.0/0 | Canal/Flannel overlay

In the example section, there are some exampe of minimal Security Group preprovisionned for different use-cased:
- [Rancher Cluster with calico network](example/calico/README.md)
- [Rancher Cluster with canal network](example/canal/README.md)

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