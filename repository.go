package caddy_webhook

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"go.uber.org/zap"
)

const (
	DefaultRemote = "origin"
	DefaultBranch = "main"
)

type Repo struct {
	URL    string
	Path   string
	Branch string
	Depth  int
	Secret string

	repo    *git.Repository
	refName plumbing.ReferenceName
}

func NewRepo(w *WebHook) *Repo {
	r := &Repo{
		URL:    w.Repository,
		Path:   w.Path,
		Branch: w.Branch,
		Depth:  w.depth,
		Secret: w.Secret,
	}

	return r
}

func (r *Repo) Setup(ctx context.Context, log *zap.Logger) error {
	var err error
	log.Info("setting up repository", zap.String("path", r.Path))

	err = r.setRef(ctx)
	if err != nil {
		return err
	}

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
			URL:           r.URL,
			RemoteName:    DefaultRemote,
			ReferenceName: r.refName,
			Depth:         r.Depth,
			Tags:          git.AllTags,
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
	if r.refName.IsBranch() {
		return r.pull(ctx)
	}

	return git.NoErrAlreadyUpToDate
}

func (r *Repo) pull(ctx context.Context) error {
	worktree, err := r.repo.Worktree()
	if err != nil {
		return err
	}

	if err := worktree.PullContext(ctx, &git.PullOptions{
		RemoteName:    DefaultRemote,
		ReferenceName: r.refName,
		Depth:         r.Depth,
	}); err != nil {
		return err
	}
	return nil
}

func (r *Repo) setRef(ctx context.Context) error {
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: DefaultRemote,
		URLs: []string{r.URL},
	})

	if err := remote.FetchContext(ctx, &git.FetchOptions{
		RemoteName: DefaultRemote,
		Depth:      r.Depth,
		Tags:       git.AllTags,
	}); err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	refs, err := remote.List(&git.ListOptions{})
	if err != nil {
		return err
	}

	if r.Branch == "" {
		r.refName = plumbing.NewBranchReferenceName(DefaultBranch)
	} else {
		branchRef := plumbing.NewBranchReferenceName(r.Branch)
		tagRef := plumbing.NewTagReferenceName(r.Branch)

		for _, ref := range refs {
			if ref.Name() == branchRef {
				r.refName = branchRef
				break
			}
			if ref.Name() == tagRef {
				r.refName = tagRef
				break
			}
		}

		if r.refName == plumbing.ReferenceName("") {
			return fmt.Errorf("reference with name '%s' not found", r.Branch)
		}
	}

	return nil
}
