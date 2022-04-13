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

package discovery

import (
	"encoding/json"

	"github.com/hashicorp/go-multierror"
	provv1 "github.com/rancher-sandbox/rancheros-operator/pkg/apis/rancheros.cattle.io/v1"
)

type Discoverer interface {
	Discovery() (res []*provv1.ManagedOSVersion, err error)
}

func Versions(d ...Discoverer) ([]byte, error) {
	var err error
	var versions []*provv1.ManagedOSVersion
	for _, dd := range d {
		res, e := dd.Discovery()
		if e != nil {
			err = multierror.Append(err, e)
		}
		versions = append(versions, res...)
	}

	b, e := json.Marshal(versions)
	if e != nil {
		err = multierror.Append(err, e)
	}

	return b, err
}
