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

package config_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/resource/config"
	"github.com/alexandrevilain/temporal-operator/pkg/version"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/utils/ptr"
)

var update = flag.Bool("update", false, "rewrite the golden files in testdata")

// newCluster is a defaulted postgres cluster shaped like eh-homelab's scadbuddy-temporal.
// This file must keep compiling against upstream v0.22.0, which is what generated the golden file.
func newCluster() *v1beta1.TemporalCluster {
	store := func(name, db string) *v1beta1.DatastoreSpec {
		return &v1beta1.DatastoreSpec{
			Name: name,
			SQL: &v1beta1.SQLSpec{
				PluginName:      "postgres12",
				User:            "temporal",
				DatabaseName:    db,
				ConnectAddr:     "db-rw:5432",
				ConnectProtocol: "tcp",
				MaxConns:        8,
				MaxIdleConns:    4,
			},
			PasswordSecretRef: &v1beta1.SecretKeyReference{Name: "db-app", Key: "password"},
			SkipCreate:        true,
		}
	}
	c := &v1beta1.TemporalCluster{
		TypeMeta:   v1beta1.TemporalClusterTypeMeta,
		ObjectMeta: metav1.ObjectMeta{Name: "temporal", Namespace: "default"},
		Spec: v1beta1.TemporalClusterSpec{
			Version:          version.MustNewVersionFromString("1.28.4"),
			NumHistoryShards: 4,
			Persistence: v1beta1.TemporalPersistenceSpec{
				DefaultStore:    store("default", "temporal"),
				VisibilityStore: store("visibility", "temporal_visibility"),
			},
			Metrics: &v1beta1.MetricsSpec{
				Enabled:    true,
				Prometheus: &v1beta1.PrometheusSpec{ListenPort: ptr.To[int32](9090)},
			},
		},
	}
	c.Default()
	return c
}

func render(t *testing.T, c *v1beta1.TemporalCluster) string {
	t.Helper()
	scheme := runtime.NewScheme()
	utilruntime.Must(v1beta1.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))

	b := config.NewConfigmapBuilder(c, scheme)
	cm := b.Build()
	require.NoError(t, b.Update(cm))
	return cm.(*corev1.ConfigMap).Data["config_template.yaml"]
}

// TestConfigmapBuilderMatchesUpstream: without clusterMetadata the fork must render exactly
// what upstream v0.22.0 renders. Regenerate the golden file only from upstream's code:
// check out v0.22.0, copy this file in, and run `go test ./internal/resource/config/ -update`.
func TestConfigmapBuilderMatchesUpstream(t *testing.T) {
	got := render(t, newCluster())
	golden := filepath.Join("testdata", "config_template.v0.22.0.yaml")
	if *update {
		require.NoError(t, os.MkdirAll("testdata", 0o755))
		require.NoError(t, os.WriteFile(golden, []byte(got), 0o600))
	}
	want, err := os.ReadFile(golden)
	require.NoError(t, err)
	require.Equal(t, string(want), got)
}
