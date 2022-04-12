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

package github_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/rancher-sandbox/upgradechannel-discovery/pkg/discovery/type/github"
)

var _ = Describe("github discovery", func() {

	Context("discovery", func() {
		It("fails if there aren't enough information", func() {
			rf, err := NewReleaseFinder()
			Expect(err).ToNot(HaveOccurred())

			_, err = rf.Discovery()
			Expect(err).To(HaveOccurred())
		})
		It("detect releases", func() {
			rf, err := NewReleaseFinder(WithRepository("rancher-sandbox/os2"))
			Expect(err).ToNot(HaveOccurred())

			res, err := rf.Discovery()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(res) > 0).To(BeTrue())

			Expect(res[0].ObjectMeta.Name).ToNot(BeEmpty())
			Expect(res[0].Spec.Metadata.Data).To(HaveKey("upgradeImage"))
		})

		It("manipulates releases results", func() {
			rf, err := NewReleaseFinder(
				WithRepository("rancher-sandbox/os2"),
				WithVersionPrefix("foo"),
				WithVersionSuffix("bar"),
				WithVersionNamePrefix("zap"),
				WithVersionNameSuffix("zof"),
			)
			Expect(err).ToNot(HaveOccurred())

			res, err := rf.Discovery()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(res) > 0).To(BeTrue())

			Expect(res[0].ObjectMeta.Name).ToNot(BeEmpty())
			Expect(res[0].Spec.Metadata.Data).To(HaveKey("upgradeImage"))

			Expect(res[0].ObjectMeta.Name).To(And(
				MatchRegexp("^zap.*"),
				MatchRegexp(".*zof$"),
			))
			Expect(res[0].Spec.Version).To(And(
				MatchRegexp("^foo.*"),
				MatchRegexp(".*bar$"),
			))
			Expect(res[0].Spec.Metadata.Data["upgradeImage"]).To(And(
				MatchRegexp(".*:foo.*bar$"),
			))
		})
	})
})
