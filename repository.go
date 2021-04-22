package caddy_webhook

import (
	"context"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"go.uber.org/zap"
)

const (
	DefaultRemote = "origin"
)

type Repo struct {
	URL    string
	Path   string
	Branch string
	Depth  int

	repo *git.Repository
}

func (r *Repo) Setup(ctx context.Context, log *zap.Logger) error {
	var err error
	log.Info("setting up repository", zap.String("path", r.Path))

	r.repo, err = git.PlainOpen(r.Path)
	if err == nil {
		err = r.repo.DeleteRemote(DefaultRemote)
		if err != nil && err != git.ErrRemoteNotFound {
			return err
		}

		_, err = r.repo.CreateRemote(&config.RemoteConfig{
			Name: DefaultRemote,
			URLs: []string{r.URL},
		})
		if err != nil {
			return err
		}

		err = r.repo.FetchContext(ctx, &git.FetchOptions{
			RemoteName: DefaultRemote,
			Depth:      r.Depth,
			Tags:       git.AllTags,
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return err
		}

	} else if err == git.ErrRepositoryNotExists {
		r.repo, err = git.PlainCloneContext(ctx, r.Path, false, &git.CloneOptions{
			URL:        r.URL,
			RemoteName: DefaultRemote,
			Depth:      r.Depth,
		})
		if err != nil {
			return err
		}
	} else {
		return err
	}

	return nil
}

func (r *Repo) Update(ctx context.Context) error {
	worktree, err := r.repo.Worktree()
	if err != nil {
		return err
	}

	if err := worktree.PullContext(ctx, &git.PullOptions{
		RemoteName: DefaultRemote,
		Depth:      r.Depth,
	}); err != nil {
		return err
	}
	return nil
}
