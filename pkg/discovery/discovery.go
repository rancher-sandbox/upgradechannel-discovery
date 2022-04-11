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

	return json.Marshal(versions)
}
