// Copyright (C) 2016-Present Pivotal Software, Inc. All rights reserved.
// This program and the accompanying materials are made available under the terms of the under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package broker_response

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-cf/on-demand-service-broker/broker"
	"github.com/pivotal-cf/on-demand-service-broker/mgmtapi"
)

type UpgradeOperation struct {
	Type UpgradeOperationType
	Data broker.OperationData
}

type UpgradeOperationType int

const (
	ResultAccepted            UpgradeOperationType = iota
	ResultOperationInProgress UpgradeOperationType = iota
	ResultNotFound            UpgradeOperationType = iota
	ResultOrphan              UpgradeOperationType = iota
)

func UpgradeOperationFrom(response *http.Response) (UpgradeOperation, error) {
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusAccepted:
		var operationData broker.OperationData
		if err := json.NewDecoder(response.Body).Decode(&operationData); err != nil {
			return UpgradeOperation{}, fmt.Errorf("cannot parse upgrade response: %s", err)
		}
		return UpgradeOperation{Type: ResultAccepted, Data: operationData}, nil
	case http.StatusNotFound:
		return UpgradeOperation{Type: ResultNotFound}, nil
	case http.StatusGone:
		return UpgradeOperation{Type: ResultOrphan}, nil
	case http.StatusConflict:
		return UpgradeOperation{Type: ResultOperationInProgress}, nil
	case http.StatusInternalServerError:
		var errorResponse brokerapi.ErrorResponse
		body, _ := ioutil.ReadAll(response.Body)
		if err := json.Unmarshal(body, &errorResponse); err != nil {
			return UpgradeOperation{}, fmt.Errorf(
				"unexpected status code: %d. cannot parse upgrade response: '%s'", response.StatusCode, body,
			)
		}

		return UpgradeOperation{}, fmt.Errorf(
			"unexpected status code: %d. description: %s", response.StatusCode, errorResponse.Description,
		)
	default:
		body, _ := ioutil.ReadAll(response.Body)
		return UpgradeOperation{}, fmt.Errorf(
			"unexpected status code: %d. body: %s", response.StatusCode, string(body),
		)
	}
}

func ListInstancesFrom(response *http.Response) ([]string, error) {
	var instances []mgmtapi.Instance
	err := decodeBodyInto(response, &instances)
	if err != nil {
		return nil, err
	}

	return instanceIDsIn(instances), nil
}

func LastOperationFrom(response *http.Response) (brokerapi.LastOperation, error) {
	var lastOperation brokerapi.LastOperation
	err := decodeBodyInto(response, &lastOperation)
	if err != nil {
		return brokerapi.LastOperation{}, err
	}

	return lastOperation, nil
}

func decodeBodyInto(response *http.Response, contents interface{}) error {
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP response status: %s", response.Status)
	}

	err := json.NewDecoder(response.Body).Decode(contents)
	if err != nil {
		return err
	}

	return nil
}

func instanceIDsIn(instances []mgmtapi.Instance) []string {
	var instanceIDs []string
	for _, instance := range instances {
		instanceIDs = append(instanceIDs, instance.InstanceID)
	}
	return instanceIDs
}
