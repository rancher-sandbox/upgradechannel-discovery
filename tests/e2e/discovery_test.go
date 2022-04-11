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
	"fmt"
	"time"

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
			)

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
	})
})
