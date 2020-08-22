package util

import (
	"context"
	"github.com/google/go-github/github"
	"github.com/innobead/kubefire/pkg/data"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

type GithubInfoer struct {
	client *github.Client
}

func NewGithubInfoer(client *github.Client) *GithubInfoer {
	return &GithubInfoer{client: client}
}

func (g *GithubInfoer) GetVersionsAfterVersion(afterVersion data.Version, repoOwner string, repo string, minorVersionCount int) ([]*data.Version, error) {
	var versions []*data.Version

	opt := github.ListOptions{}
	lastMajorVersion := afterVersion.Major
done:
	for {
		releases, resp, err := g.client.Repositories.ListReleases(context.Background(), repoOwner, repo, &opt)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		for _, release := range releases {
			if lastMajorVersion != afterVersion.Major {
				if !strings.HasPrefix(release.GetTagName(), afterVersion.MajorString()+".") {
					afterVersion.Minor = data.ParseVersion(release.GetTagName()).Minor
				}
			}

			if !strings.HasPrefix(release.GetTagName(), afterVersion.MajorMinorString()+".") {
				continue
			}

			v := data.ParseVersion(release.GetTagName())
			if v == nil {
				continue
			}
			versions = append(versions, v)

			minorVersionCount--
			if minorVersionCount == 0 {
				break done
			}

			if afterVersion.Minor.ToInt()-1 < 0 {
				// FIXME how to deal with the minor version when major version changes like from v2 to v1
				afterVersion.Major = data.SubVersionType(strconv.Itoa(afterVersion.Major.ToInt() - 1))
			} else {
				afterVersion.Minor = data.SubVersionType(strconv.Itoa(afterVersion.Minor.ToInt() - 1))
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return versions, nil
}

func (g *GithubInfoer) GetLatestVersion(repoOwner string, repo string) (*data.Version, error) {
	release, _, err := g.client.Repositories.GetLatestRelease(context.Background(), repoOwner, repo)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return data.ParseVersion(release.GetTagName()), nil
}
