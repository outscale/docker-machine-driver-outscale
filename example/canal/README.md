# Cluster with Calico

This example proivides you information on how to bootstrap a cluster with canal as network layer. In this example, we will create security groups for:
- all nodes
- control-plane nodes
- etcd nodes
- worker nodes

# Requirements
 - Terraform
 - AK/SK

# Usage

```sh
terraform init
terraform apply
```

After the creation, we will have four outputs which are the id of every security groups to fill in the Rancher Cluster form.

# Clean up
```sh
terraform init
terraform destroy
```

# Details on all SG
## Node
It is a common group with all that is needed by all nodes 
| Type | Protocol | From Port | To Port | CIDR | Description
| --- | --- | --- | --- | --- | ---
| Inbound | TCP | 22 | 22 | 0.0.0.0/0 | SSH
| Inbound | TCP | 2376 | 2376 | 0.0.0.0/0 | Docker daemon
| Inbound | TCP | 9099 | 9099 | 0.0.0.0/0 | Liveness

## ETCD
| Type | Protocol | From Port | To Port | CIDR | Description
| --- | --- | --- | --- | --- | ---
| Inbound | TCP | 2379 | 2380 | 0.0.0.0/0 | ETCD (client request and peer communication)

## Control Plane
| Type | Protocol | From Port | To Port | CIDR | Description
| --- | --- | --- | --- | --- | ---
| Inbound | TCP | 6443 | 6443 | 0.0.0.0/0 | Kube-api 
| Inbound | UDP | 8472 | 8472 | sg-node | Canal/VXLAN

## Worker
| Type | Protocol | From Port | To Port | CIDR | Description
| --- | --- | --- | --- | --- | ---
| Inbound | UDP | 8472 | 8472 | sg-node | Canal/VXLAN
| Inbound | TCP | 80 | 80 | 0.0.0.0/0 | nginx Ingress Http
| Inbound | TCP | 443 | 443 | 0.0.0.0/0 | nginx Ingress Https
| Inbound | TCP | 30000 | 32767 | 0.0.0.0/0 | Node port
| Inbound | UDP | 30000 | 32767 | 0.0.0.0/0 | Node port
