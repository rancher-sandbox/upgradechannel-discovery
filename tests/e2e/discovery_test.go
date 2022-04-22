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

package e2e_test

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/go-github/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kubectl "github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/rancheros-operator/tests/catalog"
)

var _ = Describe("Discovery e2e tests", func() {
	k := kubectl.New()
	Context("discovery", func() {
		It("Populates new ManagedOSVersion", func() {
			mr := catalog.NewManagedOSVersionChannel(
				"testchannel",
				"custom",
				map[string]interface{}{
					"image": testImage,
					"envs": []map[string]string{
						{
							"name":  "REPOSITORY",
							"value": "rancher-sandbox/os2",
						},
						{
							"name":  "OUTPUT_FILE",
							"value": "/output/data",
						},
						{
							"name":  "IMAGE_PREFIX",
							"value": "test/test",
						},
					},
					"command":    []string{"/usr/bin/upgradechannel-discovery"},
					"mountPath":  "/output",      // This defaults to /data
					"outputFile": "/output/data", // This defaults to /data/output
					"args":       []string{"github"},
				},
				nil,
			)
			defer k.Delete("managedosversionchannel", "-n", "fleet-default", "testchannel")

			Eventually(func() error {
				return k.ApplyYAML("fleet-default", "testchannel", mr)
			}, 2*time.Minute, 2*time.Second).ShouldNot(HaveOccurred())

			Eventually(func() string {
				r, err := kubectl.GetData("fleet-default", "ManagedOSVersion", "v0.1.0-alpha22", `jsonpath={.spec.metadata.upgradeImage}`)
				if err != nil {
					fmt.Println(err)
				}
				return string(r)
			}, 6*time.Minute, 2*time.Second).Should(
				Equal("test/test:v0.1.0-alpha22"),
			)
		})

		It("Populates new ManagedOSVersion with defaults from github releases", func() {
			mr := catalog.NewManagedOSVersionChannel(
				"testchannel",
				"custom",
				map[string]interface{}{
					"image": testImage,
					"envs": []map[string]string{
						{
							"name":  "REPOSITORY",
							"value": "rancher-sandbox/os2",
						},
						{
							"name":  "IMAGE_PREFIX",
							"value": "test/test2",
						},
					},
					"command": []string{"/usr/bin/upgradechannel-discovery"},
					"args":    []string{"github"},
				},
				nil,
			)
			defer k.Delete("managedosversionchannel", "-n", "fleet-default", "testchannel")

			Eventually(func() error {
				return k.ApplyYAML("fleet-default", "testchannel", mr)
			}, 2*time.Minute, 2*time.Second).ShouldNot(HaveOccurred())

			Eventually(func() string {
				r, err := kubectl.GetData("fleet-default", "ManagedOSVersion", "v0.1.0-alpha22", `jsonpath={.spec.metadata.upgradeImage}`)
				if err != nil {
					fmt.Println(err)
				}
				return string(r)
			}, 6*time.Minute, 2*time.Second).Should(
				Equal("test/test2:v0.1.0-alpha22"),
			)

			githubData, err := kubectl.GetData("fleet-default", "ManagedOSVersion", "v0.1.0-alpha22", `jsonpath={.spec.metadata.github_data}`)
			if err != nil {
				fmt.Println(err)
			}

			releases := &github.RepositoryRelease{}

			err = json.Unmarshal([]byte(githubData), &releases)
			Expect(err).ToNot(HaveOccurred(), string(githubData))

			Expect(*releases.URL).To(ContainSubstring("https"))
		})

		It("Populates new ManagedOSVersion from a git repository", func() {
			mr := catalog.NewManagedOSVersionChannel(
				"testchannel3",
				"custom",
				map[string]interface{}{
					"image": testImage,
					"envs": []map[string]string{
						{
							"name":  "REPOSITORY",
							"value": "https://github.com/rancher-sandbox/upgradechannel-discovery-test-repo",
						},
					},
					"command": []string{"/usr/bin/upgradechannel-discovery"},
					"args":    []string{"git"},
				},
				nil,
			)
			defer k.Delete("managedosversionchannel", "-n", "fleet-default", "testchannel3")

			Eventually(func() error {
				return k.ApplyYAML("fleet-default", "testchannel3", mr)
			}, 2*time.Minute, 2*time.Second).ShouldNot(HaveOccurred())

			Eventually(func() string {
				r, err := kubectl.GetData("fleet-default", "ManagedOSVersion", "v0.1.0-beta1", `jsonpath={.spec.metadata.upgradeImage}`)
				if err != nil {
					fmt.Println(err)
				}
				return string(r)
			}, 6*time.Minute, 2*time.Second).Should(
				Equal("foo/bar:v0.1.0-beta1"),
			)
		})
	})
})
