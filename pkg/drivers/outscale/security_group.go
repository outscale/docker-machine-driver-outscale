package outscale

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	retry "github.com/avast/retry-go"
	"github.com/docker/machine/libmachine/log"
	osc "github.com/outscale/osc-sdk-go/v2"
)

var (
	etcdPort                    = []int32{2379, 2380}
	kubeApiPort           int32 = 6443
	nginxIngressHttpPort  int32 = 80
	nginxIngressHttpsPort int32 = 443
	nodePort                    = []int32{30000, 32767}
	kubePort              int32 = 10250
)

func addSecurityGroupRule(d *OscDriver, sgId string, request *osc.CreateSecurityGroupRuleRequest) error {
	// Get the client
	oscApi, err := d.getClient()
	if err != nil {
		return err
	}

	var httpRes *http.Response
	var response osc.CreateSecurityGroupRuleResponse
	err = retry.Do(
		func() error {
			var response_error error
			response, httpRes, response_error = oscApi.client.SecurityGroupRuleApi.CreateSecurityGroupRule(oscApi.context).CreateSecurityGroupRuleRequest(*request).Execute()
			return response_error
		},
		defaultThrottlingRetryOption...,
	)

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

func buildSecurityGroupRule(ipProtocol string, flow string, securityGroupId string, fromPort int32, toPort int32, ipRange string) *osc.CreateSecurityGroupRuleRequest {
	securityGroupRuleRequest := osc.CreateSecurityGroupRuleRequest{}
	securityGroupRuleRequest.SetIpProtocol(ipProtocol)
	securityGroupRuleRequest.SetFlow(flow)
	securityGroupRuleRequest.SetSecurityGroupId(securityGroupId)
	securityGroupRuleRequest.SetFromPortRange(fromPort)
	securityGroupRuleRequest.SetToPortRange(toPort)
	securityGroupRuleRequest.SetIpRange(ipRange)

	return &securityGroupRuleRequest
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

	var httpRes *http.Response
	var response osc.CreateSecurityGroupResponse
	err = retry.Do(
		func() error {
			var response_error error
			response, httpRes, response_error = oscApi.client.SecurityGroupApi.CreateSecurityGroup(oscApi.context).CreateSecurityGroupRequest(request).Execute()
			return response_error
		},
		defaultThrottlingRetryOption...,
	)

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
	sshRuleRequest := buildSecurityGroupRule("tcp", "Inbound", d.SecurityGroupId, 22, 22, "0.0.0.0/0")
	if err := addSecurityGroupRule(d, d.SecurityGroupId, sshRuleRequest); err != nil {
		log.Error("Error while adding the ssh rule in the SecurityGroup")
		return err
	}

	// Add TCP Docker rule
	dockerPortRuleRequest := buildSecurityGroupRule("tcp", "Inbound", d.SecurityGroupId, defaultDockerPort, defaultDockerPort, "0.0.0.0/0")
	if err := addSecurityGroupRule(d, d.SecurityGroupId, dockerPortRuleRequest); err != nil {
		log.Error("Error while adding the docker rule in the SecurityGroup")
		return err
	}

	// Add ETCD Port
	ruleRequest := buildSecurityGroupRule("tcp", "Inbound", d.SecurityGroupId, etcdPort[0], etcdPort[1], "0.0.0.0/0")
	if err := addSecurityGroupRule(d, d.SecurityGroupId, ruleRequest); err != nil {
		log.Error("Error while adding the etcd rule in the SecurityGroup")
		return err
	}

	// Add Kube api Port
	ruleRequest = buildSecurityGroupRule("tcp", "Inbound", d.SecurityGroupId, kubeApiPort, kubeApiPort, "0.0.0.0/0")
	if err := addSecurityGroupRule(d, d.SecurityGroupId, ruleRequest); err != nil {
		log.Error("Error while adding the kubeApi rule in the SecurityGroup")
		return err
	}

	// Add nginxIngress HTTP Port
	ruleRequest = buildSecurityGroupRule("tcp", "Inbound", d.SecurityGroupId, nginxIngressHttpPort, nginxIngressHttpPort, "0.0.0.0/0")
	if err := addSecurityGroupRule(d, d.SecurityGroupId, ruleRequest); err != nil {
		log.Error("Error while adding the nginx ingress HTTP rule in the SecurityGroup")
		return err
	}

	// Add nginxIngress HTTPS Port
	ruleRequest = buildSecurityGroupRule("tcp", "Inbound", d.SecurityGroupId, nginxIngressHttpsPort, nginxIngressHttpsPort, "0.0.0.0/0")
	if err := addSecurityGroupRule(d, d.SecurityGroupId, ruleRequest); err != nil {
		log.Error("Error while adding the nginx ingress HTTPS rule in the SecurityGroup")
		return err
	}

	// Add node Port
	ruleRequest = buildSecurityGroupRule("tcp", "Inbound", d.SecurityGroupId, nodePort[0], nodePort[1], "0.0.0.0/0")
	if err := addSecurityGroupRule(d, d.SecurityGroupId, ruleRequest); err != nil {
		log.Error("Error while adding the node port rule in the SecurityGroup")
		return err
	}

	ruleRequest = buildSecurityGroupRule("udp", "Inbound", d.SecurityGroupId, nodePort[0], nodePort[1], "0.0.0.0/0")
	if err := addSecurityGroupRule(d, d.SecurityGroupId, ruleRequest); err != nil {
		log.Error("Error while adding the node port rule in the SecurityGroup")
		return err
	}

	// Kube Port
	ruleRequest = buildSecurityGroupRule("tcp", "Inbound", d.SecurityGroupId, kubePort, kubePort, "0.0.0.0/0")
	if err := addSecurityGroupRule(d, d.SecurityGroupId, ruleRequest); err != nil {
		log.Error("Error while adding the kube port rule in the SecurityGroup")
		return err
	}

	// Add extra tags
	if err := addExtraTags(d, d.SecurityGroupId, d.extraTagsAll); err != nil {
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

	if resourceId == "" {
		log.Warn("Skipping deletion of the security group because none was stored.")
		return nil
	}

	request := osc.DeleteSecurityGroupRequest{
		SecurityGroupId: &resourceId,
	}

	var httpRes *http.Response
	err = retry.Do(
		func() error {
			var response_error error
			_, httpRes, response_error = oscApi.client.SecurityGroupApi.DeleteSecurityGroup(oscApi.context).DeleteSecurityGroupRequest(request).Execute()
			return response_error
		},
		defaultThrottlingRetryOption...,
	)

	if err != nil {
		log.Error("Error while submitting the Security Group deletion request: ")
		if httpRes != nil {
			fmt.Printf(httpRes.Status)
		}
		return err
	}

	return nil
}
