// Copyright (C) 2016-Present Pivotal Software, Inc. All rights reserved.
// This program and the accompanying materials are made available under the terms of the under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package serviceadapter

import (
	"encoding/json"
	"log"

	"github.com/pivotal-cf/brokerapi"
	sdk "github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"
)

func (c *Client) GeneratePlanSchema(plan sdk.Plan, logger *log.Logger) (brokerapi.ServiceSchemas, error) {
	plan.Properties = SanitiseForJSON(plan.Properties)
	serialisedPlan, err := json.Marshal(plan)
	if err != nil {
		return brokerapi.ServiceSchemas{}, err
	}

	stdout, stderr, exitCode, err := c.CommandRunner.Run(
		c.ExternalBinPath, "generate-plan-schemas", "--plan-json", string(serialisedPlan),
	)

	if err != nil {
		return brokerapi.ServiceSchemas{}, adapterError(c.ExternalBinPath, stdout, stderr, err)
	}

	if err := ErrorForExitCode(*exitCode, string(stdout)); err != nil {
		logger.Printf(adapterFailedMessage(*exitCode, c.ExternalBinPath, stdout, stderr))
		return brokerapi.ServiceSchemas{}, err
	}

	logger.Printf("service adapter ran generate-plan-schema successfully, stderr logs: %s", string(stderr))

	var schemas brokerapi.ServiceSchemas
	err = json.Unmarshal(stdout, &schemas)
	return schemas, err
}
