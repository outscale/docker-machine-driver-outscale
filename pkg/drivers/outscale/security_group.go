package outscale

import (
	"errors"
	"fmt"
	"time"

	"github.com/docker/machine/libmachine/log"
	osc "github.com/outscale/osc-sdk-go/v2"
)

var (
	etcdPort                    = []int32{2379, 2380}
	kubeApiPort           int32 = 6443
	nginxIngressHttpPort  int32 = 80
	nginxIngressHttpsPort int32 = 443
	nodePort                    = []int32{30000, 32767}
)

func addSecurityGroupRule(d *OscDriver, sgId string, request *osc.CreateSecurityGroupRuleRequest) error {
	// Get the client
	oscApi, err := d.getClient()
	if err != nil {
		return err
	}

	response, httpRes, err := oscApi.client.SecurityGroupRuleApi.CreateSecurityGroupRule(oscApi.context).CreateSecurityGroupRuleRequest(*request).Execute()
	if err != nil {
		log.Error("Error while submitting the Security Group Rule creation request: ")
		if httpRes != nil {
			fmt.Printf(httpRes.Status)
		}
		return err
	}

	if !response.HasSecurityGroup() {
		return errors.New("Error while creating the SecurityGroupRule")
	}

	return nil

}

func createSecurityGroup(d *OscDriver) error {
	log.Debug("Creating the Security Group")

	// Get the client
	oscApi, err := d.getClient()
	if err != nil {
		return err
	}

	request := osc.CreateSecurityGroupRequest{
		Description:       fmt.Sprintf("Security Group for docker-machine %s", d.GetMachineName()),
		SecurityGroupName: fmt.Sprintf("docker-machine-%s-%d", d.GetMachineName(), time.Now().Unix()),
	}

	response, httpRes, err := oscApi.client.SecurityGroupApi.CreateSecurityGroup(oscApi.context).CreateSecurityGroupRequest(request).Execute()
	if err != nil {
		log.Error("Error while submitting the Security Group creation request: ")
		if httpRes != nil {
			fmt.Printf(httpRes.Status)
		}
		return err
	}

	if !response.HasSecurityGroup() {
		return errors.New("Error while creating the SecurityGroup")
	}

	d.SecurityGroupId = response.SecurityGroup.GetSecurityGroupId()

	// Add SSH rule
	sshRuleRequest := osc.CreateSecurityGroupRuleRequest{}
	sshRuleRequest.SetIpProtocol("tcp")
	sshRuleRequest.SetFlow("Inbound")
	sshRuleRequest.SetSecurityGroupId(d.SecurityGroupId)
	sshRuleRequest.SetFromPortRange(22)
	sshRuleRequest.SetToPortRange(22)
	sshRuleRequest.SetIpRange("0.0.0.0/0")

	if err := addSecurityGroupRule(d, d.SecurityGroupId, &sshRuleRequest); err != nil {
		log.Error("Error while adding the ssh rule in the SecurityGroup")
		return err
	}

	// Add TCP Docker rule
	dockerPortRuleRequest := osc.CreateSecurityGroupRuleRequest{}
	dockerPortRuleRequest.SetIpProtocol("tcp")
	dockerPortRuleRequest.SetFlow("Inbound")
	dockerPortRuleRequest.SetSecurityGroupId(d.SecurityGroupId)
	dockerPortRuleRequest.SetFromPortRange(defaultDockerPort)
	dockerPortRuleRequest.SetToPortRange(defaultDockerPort)
	dockerPortRuleRequest.SetIpRange("0.0.0.0/0")
	if err := addSecurityGroupRule(d, d.SecurityGroupId, &dockerPortRuleRequest); err != nil {
		log.Error("Error while adding the docker rule in the SecurityGroup")
		return err
	}

	// Add ETCD Port
	ruleRequest := osc.CreateSecurityGroupRuleRequest{}
	ruleRequest.SetIpProtocol("tcp")
	ruleRequest.SetFlow("Inbound")
	ruleRequest.SetSecurityGroupId(d.SecurityGroupId)
	ruleRequest.SetFromPortRange(etcdPort[0])
	ruleRequest.SetToPortRange(etcdPort[1])
	ruleRequest.SetIpRange("0.0.0.0/0")
	if err := addSecurityGroupRule(d, d.SecurityGroupId, &ruleRequest); err != nil {
		log.Error("Error while adding the etcd rule in the SecurityGroup")
		return err
	}

	// Add Kube api Port
	ruleRequest = osc.CreateSecurityGroupRuleRequest{}
	ruleRequest.SetIpProtocol("tcp")
	ruleRequest.SetFlow("Inbound")
	ruleRequest.SetSecurityGroupId(d.SecurityGroupId)
	ruleRequest.SetFromPortRange(kubeApiPort)
	ruleRequest.SetToPortRange(kubeApiPort)
	ruleRequest.SetIpRange("0.0.0.0/0")
	if err := addSecurityGroupRule(d, d.SecurityGroupId, &ruleRequest); err != nil {
		log.Error("Error while adding the kubeApi rule in the SecurityGroup")
		return err
	}

	// Add nginxIngress HTTP Port
	ruleRequest = osc.CreateSecurityGroupRuleRequest{}
	ruleRequest.SetIpProtocol("tcp")
	ruleRequest.SetFlow("Inbound")
	ruleRequest.SetSecurityGroupId(d.SecurityGroupId)
	ruleRequest.SetFromPortRange(nginxIngressHttpPort)
	ruleRequest.SetToPortRange(nginxIngressHttpPort)
	ruleRequest.SetIpRange("0.0.0.0/0")
	if err := addSecurityGroupRule(d, d.SecurityGroupId, &ruleRequest); err != nil {
		log.Error("Error while adding the nginx ingress HTTP rule in the SecurityGroup")
		return err
	}

	// Add nginxIngress HTTPS Port
	ruleRequest = osc.CreateSecurityGroupRuleRequest{}
	ruleRequest.SetIpProtocol("tcp")
	ruleRequest.SetFlow("Inbound")
	ruleRequest.SetSecurityGroupId(d.SecurityGroupId)
	ruleRequest.SetFromPortRange(nginxIngressHttpsPort)
	ruleRequest.SetToPortRange(nginxIngressHttpsPort)
	ruleRequest.SetIpRange("0.0.0.0/0")
	if err := addSecurityGroupRule(d, d.SecurityGroupId, &ruleRequest); err != nil {
		log.Error("Error while adding the nginx ingress HTTPS rule in the SecurityGroup")
		return err
	}

	// Add node Port
	ruleRequest = osc.CreateSecurityGroupRuleRequest{}
	ruleRequest.SetIpProtocol("tcp")
	ruleRequest.SetFlow("Inbound")
	ruleRequest.SetSecurityGroupId(d.SecurityGroupId)
	ruleRequest.SetFromPortRange(nodePort[0])
	ruleRequest.SetToPortRange(nodePort[1])
	ruleRequest.SetIpRange("0.0.0.0/0")
	if err := addSecurityGroupRule(d, d.SecurityGroupId, &ruleRequest); err != nil {
		log.Error("Error while adding the node port rule in the SecurityGroup")
		return err
	}

	ruleRequest = osc.CreateSecurityGroupRuleRequest{}
	ruleRequest.SetIpProtocol("udp")
	ruleRequest.SetFlow("Inbound")
	ruleRequest.SetSecurityGroupId(d.SecurityGroupId)
	ruleRequest.SetFromPortRange(nodePort[0])
	ruleRequest.SetToPortRange(nodePort[1])
	ruleRequest.SetIpRange("0.0.0.0/0")
	if err := addSecurityGroupRule(d, d.SecurityGroupId, &ruleRequest); err != nil {
		log.Error("Error while adding the node port rule in the SecurityGroup")
		return err
	}

	return nil
}

func deleteSecurityGroup(d *OscDriver, resourceId string) error {
	log.Debug("Deletion the Security Group")

	// Get the client
	oscApi, err := d.getClient()
	if err != nil {
		return err
	}

	request := osc.DeleteSecurityGroupRequest{
		SecurityGroupId: &resourceId,
	}

	_, httpRes, err := oscApi.client.SecurityGroupApi.DeleteSecurityGroup(oscApi.context).DeleteSecurityGroupRequest(request).Execute()
	if err != nil {
		log.Error("Error while submitting the Security Group deletion request: ")
		if httpRes != nil {
			fmt.Printf(httpRes.Status)
		}
		return err
	}

	return nil
}
