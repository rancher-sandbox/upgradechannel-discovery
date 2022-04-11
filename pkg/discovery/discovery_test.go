/*
Copyright © 2022 SUSE LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package discovery_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	provv1 "github.com/rancher-sandbox/rancheros-operator/pkg/apis/rancheros.cattle.io/v1"
	. "github.com/rancher-sandbox/upgradechannel-discovery/pkg/discovery"

	github "github.com/rancher-sandbox/upgradechannel-discovery/pkg/discovery/type/github"
)

var _ = Describe("discovery", func() {

	Context("discovery", func() {
		It("detect releases and constructs valid OSVersions", func() {
			rf, err := github.NewReleaseFinder(github.WithRepository("rancher-sandbox/os2"))
			Expect(err).ToNot(HaveOccurred())

			b, err := Versions(rf)
			Expect(err).ToNot(HaveOccurred())

			res := []*provv1.ManagedOSVersion{}
			err = json.Unmarshal(b, &res)
			Expect(err).ToNot(HaveOccurred())

			Expect(len(res) > 0).To(BeTrue())

			Expect(res[0].ObjectMeta.Name).ToNot(BeEmpty())
			Expect(res[0].Spec.Metadata.Data).To(HaveKey("upgradeImage"))
		})
	})
})
