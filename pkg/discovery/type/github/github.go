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
	baseImage         string
	githubToken       string
	repository        string
	ctx               context.Context
}

type githubSetting func(g *githubOptions) error

func WithRepository(s string) githubSetting {
	return func(g *githubOptions) error {
		g.repository = s
		return nil
	}
}
func WithContext(ctx context.Context) githubSetting {
	return func(g *githubOptions) error {
		g.ctx = ctx
		return nil
	}
}
func WithToken(s string) githubSetting {
	return func(g *githubOptions) error {
		g.githubToken = s
		return nil
	}
}
func WithBaseImage(s string) githubSetting {
	return func(g *githubOptions) error {
		g.baseImage = s
		return nil
	}
}
func WithVersionNamePrefix(s string) githubSetting {
	return func(g *githubOptions) error {
		g.versionNamePrefix = s
		return nil
	}
}
func WithVersionNameSuffix(s string) githubSetting {
	return func(g *githubOptions) error {
		g.versionNameSuffix = s
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

func NewReleaseFinder(opts ...githubSetting) (*releaseFinder, error) {
	o := &githubOptions{}

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

func (f *releaseFinder) Discovery() (res []*provv1.ManagedOSVersion, err error) {
	rels, err := f.findAll(f.opts.repository)
	for _, r := range rels {
		res = append(res, &provv1.ManagedOSVersion{
			ObjectMeta: v1.ObjectMeta{
				Name: fmt.Sprintf("%s%s%s", f.opts.versionNamePrefix, *r.TagName, f.opts.versionNameSuffix),
			},
			Spec: provv1.ManagedOSVersionSpec{
				Type:    "container",
				Version: *r.TagName,
				Metadata: &v1alpha1.GenericMap{
					Data: map[string]interface{}{
						"upgradeImage": fmt.Sprintf("%s:%s", f.opts.baseImage, *r.TagName),
					},
				},
			},
		})
	}
	return
}
