// Licensed to Alexandre VILAIN under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Alexandre VILAIN licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package v1beta1_test

import (
	"testing"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestImageTag(t *testing.T) {
	for ref, want := range map[string]string{
		"temporalio/server":                         "",
		"temporalio/server:1.28.4":                  "1.28.4",
		"temporalio/admin-tools:1.28@sha256:abc123": "1.28",
		"temporalio/server@sha256:abc123":           "",
		"registry:5000/temporalio/server":           "",
		"registry:5000/temporalio/server:1.29.7":    "1.29.7",
	} {
		assert.Equal(t, want, v1beta1.ImageTag(ref), ref)
	}
}

func TestImageWithTag(t *testing.T) {
	for ref, want := range map[string]string{
		"temporalio/server":                         "temporalio/server:1.28.4",
		"temporalio/server:1.29.7":                  "temporalio/server:1.29.7",
		"temporalio/admin-tools:1.28@sha256:abc123": "temporalio/admin-tools:1.28@sha256:abc123",
		"temporalio/server@sha256:abc123":           "temporalio/server@sha256:abc123",
		"registry:5000/temporalio/server":           "registry:5000/temporalio/server:1.28.4",
	} {
		assert.Equal(t, want, v1beta1.ImageWithTag(ref, "1.28.4"), ref)
	}
}
