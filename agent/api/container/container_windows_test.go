// +build windows,unit

// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequiresCredentialSpec(t *testing.T) {
	getContainer := func(hostConfig string) *Container {
		c := &Container{
			Name: "c",
		}
		c.DockerConfig.HostConfig = &hostConfig
		return c
	}

	testCases := []struct {
		name           string
		container      *Container
		expectedOutput bool
	}{
		{
			name:           "hostconfig_nil",
			container:      &Container{},
			expectedOutput: false,
		},
		{
			name:           "invalid_case",
			container:      getContainer("invalid"),
			expectedOutput: false,
		},
		{
			name:           "empty_sec_opt",
			container:      getContainer("{\"NetworkMode\":\"bridge\"}"),
			expectedOutput: false,
		},
		{
			name:           "missing_credentialspec",
			container:      getContainer("{\"SecurityOpt\": [\"invalid-sec-opt\"]}"),
			expectedOutput: false,
		},
		{
			name:           "valid_credentialspec",
			container:      getContainer("{\"SecurityOpt\": [\"credentialspec:file://gmsa_gmsa-acct.json\"]}"),
			expectedOutput: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedOutput, tc.container.RequiresCredentialSpec())
		})
	}
}

func TestGetCredentialSpecErr(t *testing.T) {
	getContainer := func(hostConfig string) *Container {
		c := &Container{
			Name: "c",
		}
		c.DockerConfig.HostConfig = &hostConfig
		return c
	}

	testCases := []struct {
		name                 string
		container            *Container
		expectedOutputString string
		expectedErrorString  string
	}{
		{
			name:                 "hostconfig nil",
			container:            &Container{},
			expectedOutputString: "",
			expectedErrorString:  "empty container hostConfig",
		},
		{
			name:                 "invalid case",
			container:            getContainer("invalid"),
			expectedOutputString: "",
			expectedErrorString:  "unable to unmarshal container hostConfig",
		},
		{
			name:                 "empty sec_opt",
			container:            getContainer("{\"NetworkMode\":\"bridge\"}"),
			expectedOutputString: "",
			expectedErrorString:  "unable to find container security options",
		},
		{
			name:                 "missing credentialspec",
			container:            getContainer("{\"SecurityOpt\": [\"invalid-sec-opt\"]}"),
			expectedOutputString: "",
			expectedErrorString:  "unable to obtain credentialspec",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expectedOutputStr, err := tc.container.GetCredentialSpec()
			assert.Equal(t, tc.expectedOutputString, expectedOutputStr)
			assert.EqualError(t, err, tc.expectedErrorString)
		})
	}
}

func TestGetCredentialSpecHappyPath(t *testing.T) {
	getContainer := func(hostConfig string) *Container {
		c := &Container{
			Name: "c",
		}
		c.DockerConfig.HostConfig = &hostConfig
		return c
	}

	c := getContainer("{\"SecurityOpt\": [\"credentialspec:file://gmsa_gmsa-acct.json\"]}")

	expectedCredentialSpec := "credentialspec:file://gmsa_gmsa-acct.json"

	credentialspec, err := c.GetCredentialSpec()
	assert.NoError(t, err)
	assert.EqualValues(t, expectedCredentialSpec, credentialspec)
}
