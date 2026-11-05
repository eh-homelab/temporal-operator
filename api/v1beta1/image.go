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

package v1beta1

import (
	"fmt"
	"strings"

	"github.com/alexandrevilain/temporal-operator/pkg/version"
)

// ImageTag returns the tag of an image reference ("repo:tag" or "repo:tag@sha256:..."),
// or "" if it has none. A registry port ("registry:5000/repo") is not a tag.
func ImageTag(ref string) string {
	ref, _, _ = strings.Cut(ref, "@")
	i := strings.LastIndex(ref, ":")
	if i < 0 || strings.Contains(ref[i+1:], "/") {
		return ""
	}
	return ref[i+1:]
}

// ImageWithTag returns ref as written when it already carries a tag or a digest,
// so a fully pinned reference is used verbatim, and ref:tag otherwise.
func ImageWithTag(ref, tag string) string {
	if ImageTag(ref) != "" || strings.Contains(ref, "@") {
		return ref
	}
	return ref + ":" + tag
}

// AdminToolsTagMatches reports whether an admin tools tag ("1.28", "1.28.4-tctl-...")
// is for Temporal major.minor want ("1.28").
func AdminToolsTagMatches(tag, want string) bool {
	if !strings.HasPrefix(tag, want) {
		return false
	}
	rest := tag[len(want):]
	return rest == "" || rest[0] == '.' || rest[0] == '-'
}

// SchemaAdminToolsTag returns the admin tools tag the schema jobs use for an untagged
// spec.admintools.image: spec.admintools.version when it is for the server's minor,
// and the default tag for the server version otherwise. The webhook stores the
// defaulted version, so after a server upgrade it can name the previous minor, whose
// tools would apply the wrong schema.
func (c *TemporalCluster) SchemaAdminToolsTag() string {
	want := fmt.Sprintf("%d.%d", c.Spec.Version.Major(), c.Spec.Version.Minor())
	if c.Spec.AdminTools != nil && AdminToolsTagMatches(c.Spec.AdminTools.Version, want) {
		return c.Spec.AdminTools.Version
	}
	return version.DefaultAdminToolTag(c.Spec.Version)
}
