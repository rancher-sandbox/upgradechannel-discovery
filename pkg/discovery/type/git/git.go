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

package git

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	provv1 "github.com/rancher-sandbox/rancheros-operator/pkg/apis/rancheros.cattle.io/v1"
	"github.com/sirupsen/logrus"
)

type gitOptions struct {
	repository string
	subdir     string
	branch     string
}

type gitSetting func(g *gitOptions) error

// WithRepository sets a git repository to scan releases against
func WithRepository(s string) gitSetting { //nolint:golint,revive
	return func(g *gitOptions) error {
		g.repository = s
		return nil
	}
}

// WithSubpath sets a repository subpath used to retrieve versions from
func WithSubpath(s string) gitSetting { //nolint:golint,revive
	return func(g *gitOptions) error {
		g.subdir = s
		return nil
	}
}

// WithBranch sets a repository branch used to retrieve versions from
func WithBranch(s string) gitSetting { //nolint:golint,revive
	return func(g *gitOptions) error {
		g.branch = s
		return nil
	}
}

func (g *gitOptions) apply(opts ...gitSetting) error {
	for _, o := range opts {
		if err := o(g); err != nil {
			return err
		}
	}
	return nil
}

// NewReleaseFinder returns a new git release finder discovery with the required settings
func NewReleaseFinder(opts ...gitSetting) (*releaseFinder, error) { //nolint:golint,revive
	o := &gitOptions{}

	err := o.apply(opts...)
	if err != nil {
		return nil, err
	}

	return &releaseFinder{
		opts: *o,
	}, nil
}

type releaseFinder struct {
	opts gitOptions
}

// Discovery retrieves ManagedOSVersion from git repositories
func (f *releaseFinder) Discovery() (res []*provv1.ManagedOSVersion, err error) {

	opts := &git.CloneOptions{
		URL:   f.opts.repository,
		Depth: 1,
	}

	if f.opts.branch != "" {
		opts.ReferenceName = plumbing.NewBranchReferenceName(f.opts.branch)
	}

	temp, err := os.MkdirTemp("", "rf")
	if err != nil {
		return
	}

	defer os.RemoveAll(temp)
	logrus.Infof("Cloning %s", f.opts.repository)

	_, err = git.PlainClone(temp, false, opts)
	if err != nil {
		return
	}
	logrus.Infof("Cloning of '%s' in '%s' done", f.opts.repository, temp)

	err = filepath.Walk(filepath.Join(temp, f.opts.subdir),
		func(path string, info os.FileInfo, err error) error {

			if err != nil {
				return nil
			}
			if !strings.HasSuffix(path, "json") {
				return nil
			}

			logrus.Infof("'%s' found", path)

			v := &provv1.ManagedOSVersion{}
			dat, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			err = json.Unmarshal(dat, v)
			if err == nil {
				res = append(res, v)
			}
			return nil

		})
	if err != nil {
		return
	}

	return
}
