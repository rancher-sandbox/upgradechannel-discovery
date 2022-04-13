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

package git_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/rancher-sandbox/upgradechannel-discovery/pkg/discovery/type/git"
)

var _ = Describe("git discovery", func() {
	Context("discovery", func() {
		It("fails if there aren't enough information", func() {
			rf, err := NewReleaseFinder()
			Expect(err).ToNot(HaveOccurred())

			_, err = rf.Discovery()
			Expect(err).To(HaveOccurred())
		})

		It("includes all versions", func() {
			rf, err := NewReleaseFinder(WithRepository("https://github.com/rancher-sandbox/upgradechannel-discovery-test-repo"))
			Expect(err).ToNot(HaveOccurred())

			res, err := rf.Discovery()
			Expect(err).ToNot(HaveOccurred())

			Expect(len(res)).To(Equal(3))

			names := []string{}

			for _, r := range res {
				names = append(names, r.ObjectMeta.Name)
			}
			for _, release := range res {
				Expect(release.Name).ToNot(ContainSubstring("rc1"))
			}
			Expect(names).To(ContainElements("v0.1.0-alpha22", "v0.1.0-beta1", "v0.1.0-alpha23"))
		})

		It("includes only version in subpaths", func() {
			rf, err := NewReleaseFinder(
				WithRepository("https://github.com/rancher-sandbox/upgradechannel-discovery-test-repo"),
				WithSubpath("sub"))
			Expect(err).ToNot(HaveOccurred())

			res, err := rf.Discovery()
			Expect(err).ToNot(HaveOccurred())

			Expect(len(res)).To(Equal(1))

			Expect(res[0].ObjectMeta.Name).ToNot(BeEmpty())
			Expect(res[0].ObjectMeta.Name).To(Equal("v0.1.0-beta1"))

			Expect(res[0].Spec.Metadata.Data).To(HaveKey("upgradeImage"))
			for _, release := range res {
				Expect(release.Name).ToNot(ContainSubstring("rc1"))
			}
		})

		It("includes all versions in the branch", func() {
			rf, err := NewReleaseFinder(
				WithBranch("test-branch"),
				WithRepository("https://github.com/rancher-sandbox/upgradechannel-discovery-test-repo"))
			Expect(err).ToNot(HaveOccurred())

			res, err := rf.Discovery()
			Expect(err).ToNot(HaveOccurred())

			Expect(len(res)).To(Equal(4))

			names := []string{}

			for _, r := range res {
				names = append(names, r.ObjectMeta.Name)
			}
			for _, release := range res {
				Expect(release.Name).ToNot(ContainSubstring("rc1"))
			}
			Expect(names).To(ContainElements("v0.1.0-alpha77", "v0.1.0-alpha22", "v0.1.0-beta1", "v0.1.0-alpha23"))
		})
	})
})
