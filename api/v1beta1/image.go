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

import "strings"

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
