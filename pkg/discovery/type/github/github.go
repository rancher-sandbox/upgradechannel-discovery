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

package github

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"strings"

	"github.com/google/go-github/github"
	provv1 "github.com/rancher-sandbox/rancheros-operator/pkg/apis/rancheros.cattle.io/v1"
	"github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
	"golang.org/x/oauth2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type githubOptions struct {
	versionNamePrefix string
	versionNameSuffix string
	versionSuffix     string
	versionPrefix     string
	baseImage         string
	githubToken       string
	repository        string
	ctx               context.Context
}

type githubSetting func(g *githubOptions) error

// WithRepository sets a Github repository to scan releases against
func WithRepository(s string) githubSetting { //nolint:golint,revive
	return func(g *githubOptions) error {
		g.repository = s
		return nil
	}
}

// WithContext sets a context for the discovery action
func WithContext(ctx context.Context) githubSetting { //nolint:golint,revive
	return func(g *githubOptions) error {
		g.ctx = ctx
		return nil
	}
}

// WithToken sets a github token to use for auth requests.
func WithToken(s string) githubSetting { //nolint:golint,revive
	return func(g *githubOptions) error {
		g.githubToken = s
		return nil
	}
}

// WithBaseImage Sets a base image to prefix the upgradeImage version with.
func WithBaseImage(s string) githubSetting { //nolint:golint,revive
	return func(g *githubOptions) error {
		g.baseImage = s
		return nil
	}
}

// WithVersionNamePrefix adds a prefix to the created ManagedOSVersion resource
func WithVersionNamePrefix(s string) githubSetting { //nolint:golint,revive
	return func(g *githubOptions) error {
		g.versionNamePrefix = s
		return nil
	}
}

// WithVersionNameSuffix appends a suffix to the created ManagedOSVersion resource
func WithVersionNameSuffix(s string) githubSetting { //nolint:golint,revive
	return func(g *githubOptions) error {
		g.versionNameSuffix = s
		return nil
	}
}

// WithVersionSuffix appends a suffix to the retrieved version
func WithVersionSuffix(s string) githubSetting { //nolint:golint,revive
	return func(g *githubOptions) error {
		g.versionSuffix = s
		return nil
	}
}

// WithVersionPrefix adds a prefix to the retrieved version
func WithVersionPrefix(s string) githubSetting { //nolint:golint,revive
	return func(g *githubOptions) error {
		g.versionPrefix = s
		return nil
	}
}

func (g *githubOptions) apply(opts ...githubSetting) error {
	for _, o := range opts {
		if err := o(g); err != nil {
			return err
		}
	}
	return nil
}

type releaseFinder struct {
	api  *github.Client
	opts githubOptions
}

func newHTTPClient(ctx context.Context, token string) *http.Client {
	if token == "" {
		return &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	return oauth2.NewClient(ctx, src)
}

// NewReleaseFinder returns a new Github release finder discovery with the required settings
func NewReleaseFinder(opts ...githubSetting) (*releaseFinder, error) { //nolint:golint,revive
	o := &githubOptions{
		ctx: context.Background(),
	}

	err := o.apply(opts...)
	if err != nil {
		return nil, err
	}

	hc := newHTTPClient(o.ctx, o.githubToken)
	cli := github.NewClient(hc)

	return &releaseFinder{
		api:  cli,
		opts: *o,
	}, nil
}

func (f *releaseFinder) findAll(slug string) ([]*github.RepositoryRelease, error) {
	repo := strings.Split(slug, "/")
	if len(repo) != 2 || repo[0] == "" || repo[1] == "" {
		return nil, fmt.Errorf("Invalid slug format. It should be 'owner/name': %s", slug)
	}

	rels, res, err := f.api.Repositories.ListReleases(f.opts.ctx, repo[0], repo[1], nil)
	if err != nil {
		log.Println("API returned an error response:", err)
		if res != nil && res.StatusCode == 404 {
			// 404 means repository not found or release not found. It's not an error here.
			err = nil
			log.Println("API returned 404. Repository or release not found")
		}
		return nil, err
	}

	return rels, nil
}

// Discovery retrieves ManagedOSVersion from github releases
func (f *releaseFinder) Discovery() (res []*provv1.ManagedOSVersion, err error) {
	rels, err := f.findAll(f.opts.repository)
	for _, r := range rels {
		v := strings.Join([]string{f.opts.versionPrefix, *r.TagName, f.opts.versionSuffix}, "")

		res = append(res, &provv1.ManagedOSVersion{
			ObjectMeta: v1.ObjectMeta{
				Name: strings.Join([]string{f.opts.versionNamePrefix, v, f.opts.versionNameSuffix}, ""),
			},
			Spec: provv1.ManagedOSVersionSpec{
				Type:    "container",
				Version: *r.TagName,
				Metadata: &v1alpha1.GenericMap{
					Data: map[string]interface{}{
						"upgradeImage": fmt.Sprintf("%s:%s", f.opts.baseImage, v),
					},
				},
			},
		})
	}
	return
}
