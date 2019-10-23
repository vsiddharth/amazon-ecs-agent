// +build windows

// Copyright 2014-2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package app

import (
	"github.com/aws/amazon-ecs-agent/agent/ecs_client/model/ecs"
	"github.com/aws/amazon-ecs-agent/agent/taskresource/volume"
	"github.com/aws/aws-sdk-go/aws"
)

func (agent *ecsAgent) appendVolumeDriverCapabilities(capabilities []*ecs.Attribute) []*ecs.Attribute {
	// "local" is default docker driver
	return appendNameOnlyAttribute(capabilities, attributePrefix+capabilityDockerPluginInfix+volume.DockerLocalVolumeDriver)
}

func (agent *ecsAgent) appendNvidiaDriverVersionAttribute(capabilities []*ecs.Attribute) []*ecs.Attribute {
	return capabilities
}

func (agent *ecsAgent) appendENITrunkingCapabilities(capabilities []*ecs.Attribute) []*ecs.Attribute {
	return capabilities
}

func (agent *ecsAgent) appendPIDAndIPCNamespaceSharingCapabilities(capabilities []*ecs.Attribute) []*ecs.Attribute {
	return capabilities
}

func (agent *ecsAgent) appendAppMeshCapabilities(capabilities []*ecs.Attribute) []*ecs.Attribute {
	return capabilities
}

func (agent *ecsAgent) appendTaskEIACapabilities(capabilities []*ecs.Attribute) []*ecs.Attribute {
	return capabilities
}

func (agent *ecsAgent) appendFirelensFluentdCapabilities(capabilities []*ecs.Attribute) []*ecs.Attribute {
	return capabilities
}

func (agent *ecsAgent) appendFirelensFluentbitCapabilities(capabilities []*ecs.Attribute) []*ecs.Attribute {
	return capabilities
}

func (agent *ecsAgent) appendFirelensLoggingDriverCapabilities(capabilities []*ecs.Attribute) []*ecs.Attribute {
	return capabilities
}

func (agent *ecsAgent) appendFirelensConfigCapabilities(capabilities []*ecs.Attribute) []*ecs.Attribute {
	return capabilities
}

func (agent *ecsAgent) appendBranchENIPluginVersionAttribute(capabilities []*ecs.Attribute) []*ecs.Attribute {
	// NOTE: dummy value for poc
	version := "2019.06.0"

	return append(capabilities, &ecs.Attribute{
		Name:  aws.String(attributePrefix + branchCNIPluginVersionSuffix),
		Value: aws.String(version),
	})
}

func (agent *ecsAgent) appendTaskENICapabilities(capabilities []*ecs.Attribute) []*ecs.Attribute {
	if agent.cfg.TaskENIEnabled {
		// The assumption here is that all of the dependencies for supporting the
		// Task ENI in the Agent have already been validated prior to the invocation of
		// the `agent.capabilities()` call
		capabilities = append(capabilities, &ecs.Attribute{
			Name: aws.String(attributePrefix + taskENIAttributeSuffix),
		})

		taskENIVersionAttribute, err := agent.getTaskENIPluginVersionAttribute()
		if err != nil {
			return capabilities
		}
		capabilities = append(capabilities, taskENIVersionAttribute)

		// We only care about AWSVPCBlockInstanceMetdata if Task ENI is enabled
		if agent.cfg.AWSVPCBlockInstanceMetdata {
			// If the Block Instance Metadata flag is set for AWS VPC networking mode, register a capability
			// indicating the same
			capabilities = append(capabilities, &ecs.Attribute{
				Name: aws.String(attributePrefix + taskENIBlockInstanceMetadataAttributeSuffix),
			})
		}
	}

	return capabilities
}
