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

package controllers

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// The API server must reject a duration the operator cannot decode: a stored
// one fails the informer's list and stops the watch for every object of that kind.
var _ = Describe("Duration fields", func() {
	newNamespace := func(name, retention string) *unstructured.Unstructured {
		u := &unstructured.Unstructured{}
		u.SetAPIVersion("temporal.io/v1beta1")
		u.SetKind("TemporalNamespace")
		u.SetNamespace("default")
		u.SetName(name)
		u.Object["spec"] = map[string]any{
			"clusterRef":      map[string]any{"name": "cluster"},
			"retentionPeriod": retention,
		}
		return u
	}

	DescribeTable("retentionPeriod",
		func(name, retention string, valid bool) {
			err := k8sClient.Create(context.Background(), newNamespace(name, retention))
			if valid {
				Expect(err).NotTo(HaveOccurred())
				return
			}
			Expect(apierrors.IsInvalid(err)).To(BeTrue(), "%v", err)
			Expect(err.Error()).To(ContainSubstring("spec.retentionPeriod"))
		},
		Entry("hours", "valid-hours", "720h", true),
		Entry("compound", "valid-compound", "1h30m0.5s", true),
		Entry("zero", "valid-zero", "0", true),
		Entry("days", "days", "30d", false),
		Entry("empty", "empty", "", false),
		Entry("no unit", "no-unit", "5", false),
		Entry("negative", "negative", "-1h", false),
		Entry("overflow", "overflow", "9999999999h", false),
	)
})
