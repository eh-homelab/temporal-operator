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
	"testing"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"k8s.io/utils/ptr"
)

func TestConfigmapBuilderClusterMetadata(t *testing.T) {
	type rendered struct {
		ClusterMetadata struct {
			EnableGlobalNamespace    bool   `yaml:"enableGlobalNamespace"`
			FailoverVersionIncrement int64  `yaml:"failoverVersionIncrement"`
			MasterClusterName        string `yaml:"masterClusterName"`
			CurrentClusterName       string `yaml:"currentClusterName"`
			ClusterInformation       map[string]struct {
				InitialFailoverVersion int64 `yaml:"initialFailoverVersion"`
			} `yaml:"clusterInformation"`
		} `yaml:"clusterMetadata"`
		DCRedirectionPolicy struct {
			Policy string `yaml:"policy"`
		} `yaml:"dcRedirectionPolicy"`
	}

	tests := map[string]struct {
		spec              *v1beta1.ClusterMetadataSpec
		wantGlobal        bool
		wantFailover      int64
		wantDCRedirection string
	}{
		"empty block keeps the defaults": {
			spec:         &v1beta1.ClusterMetadataSpec{},
			wantFailover: 1,
		},
		"global namespaces with the default version": {
			spec:         &v1beta1.ClusterMetadataSpec{EnableGlobalNamespace: true},
			wantGlobal:   true,
			wantFailover: 1,
		},
		"all fields": {
			spec: &v1beta1.ClusterMetadataSpec{
				EnableGlobalNamespace:  true,
				InitialFailoverVersion: ptr.To[int64](2),
				DCRedirectionPolicy:    "selected-apis-forwarding",
			},
			wantGlobal:        true,
			wantFailover:      2,
			wantDCRedirection: "selected-apis-forwarding",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := newCluster()
			c.Spec.ClusterMetadata = tt.spec

			var got rendered
			require.NoError(t, yaml.Unmarshal([]byte(render(t, c)), &got))

			md := got.ClusterMetadata
			require.Equal(t, tt.wantGlobal, md.EnableGlobalNamespace)
			require.Equal(t, v1beta1.FailoverVersionIncrement, md.FailoverVersionIncrement)
			require.Equal(t, c.Name, md.MasterClusterName)
			require.Equal(t, c.Name, md.CurrentClusterName)
			require.Len(t, md.ClusterInformation, 1)
			require.Equal(t, tt.wantFailover, md.ClusterInformation[c.Name].InitialFailoverVersion)
			require.Equal(t, tt.wantDCRedirection, got.DCRedirectionPolicy.Policy)
		})
	}
}
