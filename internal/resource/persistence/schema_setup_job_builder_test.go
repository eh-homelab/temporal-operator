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

package persistence_test

import (
	"testing"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/resource/persistence"
	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// The schema jobs migrate with the admin tools image: a pinned reference is used as
// written, and an untagged one gets the tag for the server's minor.
func TestSchemaJobAdminToolsImage(t *testing.T) {
	for adminTools, want := range map[string]string{
		"temporalio/admin-tools":                    "temporalio/admin-tools:1.28",
		"temporalio/admin-tools:1.28@sha256:abc123": "temporalio/admin-tools:1.28@sha256:abc123",
	} {
		c := &v1beta1.TemporalCluster{
			TypeMeta:   v1beta1.TemporalClusterTypeMeta,
			ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
			Spec: v1beta1.TemporalClusterSpec{
				Image:      "temporalio/server:1.28.4",
				AdminTools: &v1beta1.TemporalAdminToolsSpec{Image: adminTools},
				Persistence: v1beta1.TemporalPersistenceSpec{
					DefaultStore:    &v1beta1.DatastoreSpec{Name: "default", SQL: &v1beta1.SQLSpec{PluginName: "postgres12"}},
					VisibilityStore: &v1beta1.DatastoreSpec{Name: "visibility", SQL: &v1beta1.SQLSpec{PluginName: "postgres12"}},
				},
			},
		}
		c.Default()

		job := persistence.NewSchemaJobBuilder(c, runtime.NewScheme(), "update-default-schema", []string{"true"}).Build().(*batchv1.Job)
		assert.Equal(t, want, job.Spec.Template.Spec.Containers[0].Image, adminTools)
	}
}
