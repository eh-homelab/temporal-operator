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
	"testing"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/webhooks"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// A cluster stored without the webhook's defaults (written before an operator
// upgrade added them) is defaulted in memory, and nothing is written back.
func TestGetClusterDefaultsInMemory(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, v1beta1.AddToScheme(scheme))

	stored := &v1beta1.TemporalCluster{
		ObjectMeta: metav1.ObjectMeta{Name: "prod", Namespace: "demo"},
		Spec: v1beta1.TemporalClusterSpec{
			Image: "temporalio/server:1.28.4",
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(stored).Build()
	r := &TemporalClusterReconciler{Base: Base{Client: c}}
	key := types.NamespacedName{Name: "prod", Namespace: "demo"}

	got, err := r.getCluster(context.Background(), key)
	require.NoError(t, err)
	require.NotNil(t, got.Spec.AdminTools)
	require.Equal(t, "temporalio/admin-tools", got.Spec.AdminTools.Image)
	require.Equal(t, "1.28.4", got.Spec.Version.String())

	onServer := &v1beta1.TemporalCluster{}
	require.NoError(t, c.Get(context.Background(), key, onServer))
	require.Nil(t, onServer.Spec.AdminTools)
}

// The reconciler must see a stored cluster exactly as the mutating webhook would
// have written it, or a cluster's resources change when it is next written.
func TestGetClusterMatchesWebhookDefaults(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, v1beta1.AddToScheme(scheme))

	tests := map[string]v1beta1.TemporalClusterSpec{
		"empty":        {},
		"tagged image": {Image: "temporalio/server:1.28.4"},
		"deprecated prometheus listenAddress": {
			Metrics: &v1beta1.MetricsSpec{
				Enabled:    true,
				Prometheus: &v1beta1.PrometheusSpec{ListenAddress: "0.0.0.0:8000"},
			},
		},
		"admin tools and ui without images": {
			AdminTools: &v1beta1.TemporalAdminToolsSpec{Enabled: false},
			UI:         &v1beta1.TemporalUISpec{Enabled: true, Version: "2.31.2"},
		},
		"cert-manager mTLS": {
			MTLS: &v1beta1.MTLSSpec{Provider: v1beta1.CertManagerMTLSProvider},
		},
	}

	for name, spec := range tests {
		t.Run(name, func(t *testing.T) {
			stored := &v1beta1.TemporalCluster{
				ObjectMeta: metav1.ObjectMeta{Name: "prod", Namespace: "demo"},
				Spec:       spec,
			}
			want := stored.DeepCopy()
			require.NoError(t, (&webhooks.TemporalClusterWebhook{}).Default(context.Background(), want))

			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(stored).Build()
			r := &TemporalClusterReconciler{Base: Base{Client: c}}
			got, err := r.getCluster(context.Background(), types.NamespacedName{Name: "prod", Namespace: "demo"})
			require.NoError(t, err)
			require.Equal(t, want.Spec, got.Spec)
		})
	}
}

// A stored cluster that sets only the deprecated listenAddress keeps its port.
// Before Default did this conversion, only the webhook did, and the reconciler
// would have defaulted the port to 9090.
func TestGetClusterKeepsDeprecatedListenAddressPort(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, v1beta1.AddToScheme(scheme))

	stored := &v1beta1.TemporalCluster{
		ObjectMeta: metav1.ObjectMeta{Name: "prod", Namespace: "demo"},
		Spec: v1beta1.TemporalClusterSpec{
			Metrics: &v1beta1.MetricsSpec{
				Enabled:    true,
				Prometheus: &v1beta1.PrometheusSpec{ListenAddress: "0.0.0.0:8000"},
			},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(stored).Build()
	r := &TemporalClusterReconciler{Base: Base{Client: c}}

	got, err := r.getCluster(context.Background(), types.NamespacedName{Name: "prod", Namespace: "demo"})
	require.NoError(t, err)
	require.Equal(t, int32(8000), *got.Spec.Metrics.Prometheus.ListenPort)
	require.Empty(t, got.Spec.Metrics.Prometheus.ListenAddress)
}
