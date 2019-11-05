// +build windows

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

package credentialspec

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	apicontainer "github.com/aws/amazon-ecs-agent/agent/api/container"
	"github.com/aws/amazon-ecs-agent/agent/api/task/status"
	"github.com/aws/amazon-ecs-agent/agent/credentials"
	"github.com/aws/amazon-ecs-agent/agent/s3"
	s3factory "github.com/aws/amazon-ecs-agent/agent/s3/factory"
	"github.com/aws/amazon-ecs-agent/agent/ssm"
	ssmfactory "github.com/aws/amazon-ecs-agent/agent/ssm/factory"
	"github.com/aws/amazon-ecs-agent/agent/taskresource"
	resourcestatus "github.com/aws/amazon-ecs-agent/agent/taskresource/status"
	"github.com/aws/amazon-ecs-agent/agent/utils/ioutilwrapper"
	"github.com/aws/amazon-ecs-agent/agent/utils/oswrapper"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/cihub/seelog"
	"github.com/pkg/errors"
)

// CredentialSpecResource is the abstraction for credentialspec resources
type CredentialSpecResource struct {
	taskARN string
	region  string

	credentialsManager     credentials.Manager
	executionCredentialsID string

	ioutil ioutilwrapper.IOUtil
	os     oswrapper.OS

	createdAt           time.Time
	desiredStatusUnsafe resourcestatus.ResourceStatus
	knownStatusUnsafe   resourcestatus.ResourceStatus
	// appliedStatus is the status that has been "applied" (e.g., we've called some
	// operation such as 'Create' on the resource) but we don't yet know that the
	// application was successful, which may then change the known status. This is
	// used while progressing resource states in progressTask() of task manager
	appliedStatus                      resourcestatus.ResourceStatus
	resourceStatusToTransitionFunction map[resourcestatus.ResourceStatus]func() error

	// terminalReason should be set for resource creation failures. This ensures
	// the resource object carries some context for why provisioning failed.
	terminalReason     string
	terminalReasonOnce sync.Once

	// ssmClientCreator is a factory interface that creates new SSM clients. This is
	// needed mostly for testing.
	ssmClientCreator ssmfactory.SSMClientCreator

	// s3ClientCreator is a factory interface that creates new S3 clients. This is
	// needed mostly for testing.
	s3ClientCreator s3factory.S3ClientCreator

	// required for processing credentialspecs, key is input credentialspec
	// Example key := credentialspec:file://credentialspec.json
	requiredCredentialSpecs map[string][]*apicontainer.Container

	// map to transform credentialspec values, key is a input credentialspec
	// Examples:
	// * key := credentialspec:file://credentialspec.json, value := credentialspec=file://credentialspec.json
	// * key := credentialspec:s3ARN, value := credentialspec=file://CredentialSpecResourceDir/s3_taskARN_fileName.json
	// * key := credentialspec:ssmARN, value := credentialspec=file://CredentialSpecResourceDir/ssm_taskARN_param.json
	credSpecMap map[string]string

	// lock is used for fields that are accessed and updated concurrently
	lock sync.RWMutex
}

// NewCredentialSpecResource creates a new CredentialSpecResource object
func NewCredentialSpecResource(taskARN, region string,
	credentialSpecs map[string][]*apicontainer.Container,
	executionCredentialsID string,
	credentialsManager credentials.Manager,
	ssmClientCreator ssmfactory.SSMClientCreator,
	s3ClientCreator s3factory.S3ClientCreator) *CredentialSpecResource {

	s := &CredentialSpecResource{
		taskARN:                 taskARN,
		region:                  region,
		requiredCredentialSpecs: credentialSpecs,
		credentialsManager:      credentialsManager,
		executionCredentialsID:  executionCredentialsID,
		ssmClientCreator:        ssmClientCreator,
		s3ClientCreator:         s3ClientCreator,
		credSpecMap:             make(map[string]string),
	}

	s.initStatusToTransition()
	return s
}

func (cs *CredentialSpecResource) initStatusToTransition() {
	resourceStatusToTransitionFunction := map[resourcestatus.ResourceStatus]func() error{
		resourcestatus.ResourceStatus(CredentialSpecCreated): cs.Create,
	}
	cs.resourceStatusToTransitionFunction = resourceStatusToTransitionFunction
}

func (cs *CredentialSpecResource) Initialize(resourceFields *taskresource.ResourceFields,
	taskKnownStatus status.TaskStatus,
	taskDesiredStatus status.TaskStatus) {
	cs.initStatusToTransition()
	cs.credentialsManager = resourceFields.CredentialsManager
	cs.ssmClientCreator = resourceFields.SSMClientCreator

	// if task hasn't turn to 'created' status, and it's desire status is 'running'
	// the resource status needs to be reset to 'NONE' status so the cs value
	// will be retrieved again
	if taskKnownStatus < status.TaskCreated &&
		taskDesiredStatus <= status.TaskRunning {
		cs.SetKnownStatus(resourcestatus.ResourceStatusNone)
	}
}

// GetTerminalReason returns an error string to propagate up through to task
// state change messages
func (cs *CredentialSpecResource) GetTerminalReason() string {
	return cs.terminalReason
}

func (cs *CredentialSpecResource) setTerminalReason(reason string) {
	cs.terminalReasonOnce.Do(func() {
		seelog.Infof("credentialspec resource: setting terminal reason for credentialspec resource in task: [%s]", cs.taskARN)
		cs.terminalReason = reason
	})
}

// GetDesiredStatus safely returns the desired status of the task
func (cs *CredentialSpecResource) GetDesiredStatus() resourcestatus.ResourceStatus {
	cs.lock.RLock()
	defer cs.lock.RUnlock()

	return cs.desiredStatusUnsafe
}

// SetDesiredStatus safely sets the desired status of the resource
func (cs *CredentialSpecResource) SetDesiredStatus(status resourcestatus.ResourceStatus) {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	cs.desiredStatusUnsafe = status
}

// DesiredTerminal returns true if the credentialspec's desired status is REMOVED
func (cs *CredentialSpecResource) DesiredTerminal() bool {
	cs.lock.RLock()
	defer cs.lock.RUnlock()

	return cs.desiredStatusUnsafe == resourcestatus.ResourceStatus(CredentialSpecRemoved)
}

// KnownCreated returns true if the credentialspec's known status is CREATED
func (cs *CredentialSpecResource) KnownCreated() bool {
	cs.lock.RLock()
	defer cs.lock.RUnlock()

	return cs.knownStatusUnsafe == resourcestatus.ResourceStatus(CredentialSpecCreated)
}

// TerminalStatus returns the last transition state of credentialspec
func (cs *CredentialSpecResource) TerminalStatus() resourcestatus.ResourceStatus {
	return resourcestatus.ResourceStatus(CredentialSpecRemoved)
}

// NextKnownState returns the state that the resource should
// progress to based on its `KnownState`.
func (cs *CredentialSpecResource) NextKnownState() resourcestatus.ResourceStatus {
	return cs.GetKnownStatus() + 1
}

// ApplyTransition calls the function required to move to the specified status
func (cs *CredentialSpecResource) ApplyTransition(nextState resourcestatus.ResourceStatus) error {
	transitionFunc, ok := cs.resourceStatusToTransitionFunction[nextState]
	if !ok {
		return errors.Errorf("resource [%s]: transition to %s impossible", cs.GetName(),
			cs.StatusString(nextState))
	}
	return transitionFunc()
}

// SteadyState returns the transition state of the resource defined as "ready"
func (cs *CredentialSpecResource) SteadyState() resourcestatus.ResourceStatus {
	return resourcestatus.ResourceStatus(CredentialSpecCreated)
}

// SetKnownStatus safely sets the currently known status of the resource
func (cs *CredentialSpecResource) SetKnownStatus(status resourcestatus.ResourceStatus) {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	cs.knownStatusUnsafe = status
	cs.updateAppliedStatusUnsafe(status)
}

// updateAppliedStatusUnsafe updates the resource transitioning status
func (cs *CredentialSpecResource) updateAppliedStatusUnsafe(knownStatus resourcestatus.ResourceStatus) {
	if cs.appliedStatus == resourcestatus.ResourceStatus(CredentialSpecStatusNone) {
		return
	}

	// Check if the resource transition has already finished
	if cs.appliedStatus <= knownStatus {
		cs.appliedStatus = resourcestatus.ResourceStatus(CredentialSpecStatusNone)
	}
}

// SetAppliedStatus sets the applied status of resource and returns whether
// the resource is already in a transition
func (cs *CredentialSpecResource) SetAppliedStatus(status resourcestatus.ResourceStatus) bool {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	if cs.appliedStatus != resourcestatus.ResourceStatus(CredentialSpecStatusNone) {
		// return false to indicate the set operation failed
		return false
	}

	cs.appliedStatus = status
	return true
}

// GetKnownStatus safely returns the currently known status of the task
func (cs *CredentialSpecResource) GetKnownStatus() resourcestatus.ResourceStatus {
	cs.lock.RLock()
	defer cs.lock.RUnlock()

	return cs.knownStatusUnsafe
}

// StatusString returns the string of the cgroup resource status
func (cs *CredentialSpecResource) StatusString(status resourcestatus.ResourceStatus) string {
	return CredentialSpecStatus(status).String()
}

// SetCreatedAt sets the timestamp for resource's creation time
func (cs *CredentialSpecResource) SetCreatedAt(createdAt time.Time) {
	if createdAt.IsZero() {
		return
	}
	cs.lock.Lock()
	defer cs.lock.Unlock()

	cs.createdAt = createdAt
}

// GetCreatedAt sets the timestamp for resource's creation time
func (cs *CredentialSpecResource) GetCreatedAt() time.Time {
	cs.lock.RLock()
	defer cs.lock.RUnlock()

	return cs.createdAt
}

// getRequiredCredentialSpecs returns the requiredCredentialSpecs field of credentialspec task resource
func (cs *CredentialSpecResource) getRequiredCredentialSpecs() map[string][]*apicontainer.Container {
	cs.lock.RLock()
	defer cs.lock.RUnlock()

	return cs.requiredCredentialSpecs
}

// getExecutionCredentialsID returns the execution role's credential ID
func (cs *CredentialSpecResource) getExecutionCredentialsID() string {
	cs.lock.RLock()
	defer cs.lock.RUnlock()

	return cs.executionCredentialsID
}

// GetName safely returns the name of the resource
func (cs *CredentialSpecResource) GetName() string {
	cs.lock.RLock()
	defer cs.lock.RUnlock()

	return ResourceName
}

// Create is used to create all the credentialspec resources for a given task
func (cs *CredentialSpecResource) Create() error {
	// To fail fast, check execution role first
	executionCredentials, ok := cs.credentialsManager.GetTaskCredentials(cs.getExecutionCredentialsID())
	if !ok {
		// No need to log here. managedTask.applyResourceState already does that
		err := errors.New("credentialspec resource: unable to find execution role credentials")
		cs.setTerminalReason(err.Error())
		return err
	}
	iamCredentials := executionCredentials.GetIAMRoleCredentials()

	for credSpecStr, _ := range cs.requiredCredentialSpecs {
		credSpecSplit := strings.SplitAfterN(credSpecStr, "credentialspec:", 2)
		credSpecValue := credSpecSplit[1]

		if strings.HasPrefix(credSpecValue, "file://") {
			dockerHostconfigSecOptCredSpec := strings.Replace(credSpecStr, "credentialspec:", "credentialspec=", 1)
			cs.updateCredSpecMapping(credSpecStr, dockerHostconfigSecOptCredSpec)

			return nil
		}

		parsedARN, err := arn.Parse(credSpecValue)
		if err != nil {
			cs.setTerminalReason(err.Error())
			return err
		}

		parsedARNService := parsedARN.Service
		if parsedARNService == "s3" {
			s3ResourceARN := parsedARN
			s3CredSpecValue := credSpecValue

			bucket, key, err := s3.ParseS3ARN(s3CredSpecValue)
			if err != nil {
				cs.setTerminalReason(err.Error())
				return err
			}

			s3Client, err := cs.s3ClientCreator.NewS3ClientForBucket(bucket, cs.region, iamCredentials)
			if err != nil {
				cs.setTerminalReason(err.Error())
				return errors.Wrapf(err, "unable to initialize s3 client for bucket %s", bucket)
			}

			resourceBase := filepath.Base(s3ResourceARN.Resource)
			localCredSpecFilePath := fmt.Sprintf("%s/s3_%s_%s.json", CredentialSpecResourceDir, cs.taskARN, resourceBase)

			err = cs.writeS3File(func(file oswrapper.File) error {
				return s3.DownloadFile(bucket, key, s3DownloadTimeout, file, s3Client)
			}, localCredSpecFilePath)
			if err != nil {
				cs.setTerminalReason(err.Error())
				return errors.Wrapf(err, "unable to download s3 file %s from bucket %s", key, bucket)
			}

			dockerHostconfigSecOptCredSpec := fmt.Sprintf("credentialspec=file://%s", localCredSpecFilePath)
			cs.updateCredSpecMapping(credSpecValue, dockerHostconfigSecOptCredSpec)

		} else if parsedARNService == "ssm" {
			ssmResourceARN := parsedARN

			ssmClient := cs.ssmClientCreator.NewSSMClient(cs.region, iamCredentials)

			ssmParam := filepath.Base(ssmResourceARN.Resource)
			ssmParams := []string{ssmParam}

			ssmParamMap, err := ssm.GetParametersFromSSM(ssmParams, ssmClient)
			if err != nil {
				cs.setTerminalReason(err.Error())
				return err
			}

			ssmParamData := ssmParamMap[ssmParam]

			localCredSpecFilePath := fmt.Sprintf("%s/ssm_%s_%s.json", CredentialSpecResourceDir, cs.taskARN, ssmParam)

			err = cs.writeSSMFile(ssmParamData, localCredSpecFilePath)
			if err != nil {
				cs.setTerminalReason(err.Error())
				return err
			}

			dockerHostconfigSecOptCredSpec := fmt.Sprintf("credentialspec=file://%s", localCredSpecFilePath)
			cs.updateCredSpecMapping(credSpecValue, dockerHostconfigSecOptCredSpec)

		} else {
			err := errors.New("unsupported credentialspec ARN dependency, only s3/ssm ARNs are valid")
			cs.setTerminalReason(err.Error())
			return err
		}
	}

	return nil
}

func (cs *CredentialSpecResource) writeS3File(writeFunc func(file oswrapper.File) error, filePath string) error {
	temp, err := cs.ioutil.TempFile(CredentialSpecResourceDir, tempFileName)
	if err != nil {
		return err
	}
	defer temp.Close()

	err = writeFunc(temp)
	if err != nil {
		return err
	}

	err = temp.Chmod(os.FileMode(filePerm))
	if err != nil {
		return err
	}

	// Persist the file to disk.
	err = temp.Sync()
	if err != nil {
		return err
	}

	err = cs.os.Rename(temp.Name(), filePath)
	if err != nil {
		return err
	}

	return nil
}

func (cs *CredentialSpecResource) writeSSMFile(ssmParamData, filePath string) error {
	return cs.ioutil.WriteFile(filePath, []byte(ssmParamData), filePerm)
}

func (cs *CredentialSpecResource) getCredSpecMap() map[string]string {
	cs.lock.RLock()
	defer cs.lock.RUnlock()

	return cs.credSpecMap
}

func (cs *CredentialSpecResource) GetTargetMapping(credSpecInput string) (string, error) {
	cs.lock.RLock()
	defer cs.lock.RUnlock()

	targetCredSpecMapping, ok := cs.credSpecMap[credSpecInput]
	if !ok {
		return "", errors.New("unable to obtain credentialspec mapping")
	}

	return targetCredSpecMapping, nil
}

func (cs *CredentialSpecResource) updateCredSpecMapping(credSpecInput, targetCredSpec string) {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	cs.credSpecMap[credSpecInput] = targetCredSpec
}

// Cleanup removes the credentialspec created for the task
func (cs *CredentialSpecResource) Cleanup() error {
	cs.clearCredentialSpec()
	return nil
}

// clearCredentialSpec cycles through the collection of credentialspec data and
// removes them from the task
func (cs *CredentialSpecResource) clearCredentialSpec() {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	for key := range cs.credSpecMap {
		// TODO: Cleanup file on container instance
		delete(cs.credSpecMap, key)
	}
}

// CredentialSpecResourceJSON is the json representation of the credentialspec resource
type CredentialSpecResourceJSON struct {
	TaskARN                 string                               `json:"taskARN"`
	CreatedAt               *time.Time                           `json:"createdAt,omitempty"`
	DesiredStatus           *CredentialSpecStatus                `json:"desiredStatus"`
	KnownStatus             *CredentialSpecStatus                `json:"knownStatus"`
	RequiredCredentialSpecs map[string][]*apicontainer.Container `json:"credentialSpecResources"`
	CredSpecMap             map[string]string                    `json:"credSpecMap"`
	ExecutionCredentialsID  string                               `json:"executionCredentialsID"`
}

// MarshalJSON serialises the CredentialSpecResourceJSON struct to JSON
func (cs *CredentialSpecResource) MarshalJSON() ([]byte, error) {
	if cs == nil {
		return nil, errors.New("credential specresource is nil")
	}
	createdAt := cs.GetCreatedAt()
	return json.Marshal(CredentialSpecResourceJSON{
		TaskARN:   cs.taskARN,
		CreatedAt: &createdAt,
		DesiredStatus: func() *CredentialSpecStatus {
			desiredState := cs.GetDesiredStatus()
			s := CredentialSpecStatus(desiredState)
			return &s
		}(),
		KnownStatus: func() *CredentialSpecStatus {
			knownState := cs.GetKnownStatus()
			s := CredentialSpecStatus(knownState)
			return &s
		}(),
		RequiredCredentialSpecs: cs.getRequiredCredentialSpecs(),
		CredSpecMap:             cs.getCredSpecMap(),
		ExecutionCredentialsID:  cs.getExecutionCredentialsID(),
	})
}

// UnmarshalJSON deserialises the raw JSON to a CredentialSpecResourceJSON struct
func (cs *CredentialSpecResource) UnmarshalJSON(b []byte) error {
	temp := CredentialSpecResourceJSON{}

	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}

	if temp.DesiredStatus != nil {
		cs.SetDesiredStatus(resourcestatus.ResourceStatus(*temp.DesiredStatus))
	}
	if temp.KnownStatus != nil {
		cs.SetKnownStatus(resourcestatus.ResourceStatus(*temp.KnownStatus))
	}
	if temp.CreatedAt != nil && !temp.CreatedAt.IsZero() {
		cs.SetCreatedAt(*temp.CreatedAt)
	}
	if temp.RequiredCredentialSpecs != nil {
		cs.requiredCredentialSpecs = temp.RequiredCredentialSpecs
	}
	cs.taskARN = temp.TaskARN
	cs.executionCredentialsID = temp.ExecutionCredentialsID

	return nil
}
