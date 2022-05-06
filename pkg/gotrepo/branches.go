package gotrepo

import (
	"bytes"
	"context"

	"github.com/brendoncarroll/go-state"
	"github.com/gotvc/got/pkg/branches"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

// CreateBranch creates a branch using the default spec.
func (r *Repo) CreateBranch(ctx context.Context, name string, params branches.Params) (*Branch, error) {
	return r.space.Create(ctx, name, params)
}

// CreateBranchWithSpec creates a branch using spec
func (r *Repo) CreateBranchWithSpec(name string, spec BranchSpec) (*Branch, error) {
	return r.specDir.CreateWithSpec(name, spec)
}

// DeleteBranch deletes a branch
func (r *Repo) DeleteBranch(ctx context.Context, name string) error {
	return r.space.Delete(ctx, name)
}

// GetBranch returns the branch with the specified name
func (r *Repo) GetBranch(ctx context.Context, name string) (*Branch, error) {
	if name == "" {
		_, branch, err := r.GetActiveBranch(ctx)
		return branch, err
	}
	return r.space.Get(ctx, name)
}

// ForEachBranch calls fn once for each branch, or until an error is returned from fn
func (r *Repo) ForEachBranch(ctx context.Context, fn func(string) error) error {
	return r.space.ForEach(ctx, state.Span[string]{}, fn)
}

// SetActiveBranch sets the active branch to name
func (r *Repo) SetActiveBranch(ctx context.Context, name string) error {
	branch, err := r.GetBranch(ctx, name)
	if err != nil {
		return err
	}
	isEmpty, err := r.stage.IsEmpty(ctx)
	if err != nil {
		return err
	}
	if !isEmpty {
		current, err := r.GetBranch(ctx, "")
		if err != nil {
			return err
		}
		if !bytes.Equal(branch.Salt, current.Salt) {
			return errors.Errorf("staging must be empty to change to a branch with a different salt")
		}
	}
	return setActiveBranch(r.db, name)
}

// GetActiveBranch returns the name of the active branch, and the branch
func (r *Repo) GetActiveBranch(ctx context.Context) (string, *Branch, error) {
	name, err := getActiveBranch(r.db)
	if err != nil {
		return "", nil, err
	}
	if name == "" {
		name = nameMaster
	}
	branch, err := r.GetSpace().Get(ctx, name)
	if err != nil {
		return "", nil, err
	}
	return name, branch, nil
}

// SetBranchHead
func (r *Repo) SetBranchHead(ctx context.Context, name string, snap Snap) error {
	branch, err := r.GetBranch(ctx, name)
	if err != nil {
		return err
	}
	return branches.SetHead(ctx, *branch, r.stagingTriple(), snap)
}

func (r *Repo) GetBranchHead(ctx context.Context, name string) (*Snap, error) {
	branch, err := r.GetBranch(ctx, name)
	if err != nil {
		return nil, err
	}
	return branches.GetHead(ctx, *branch)
}

// Fork creates a new branch called next and sets its head to match base's
func (r *Repo) Fork(ctx context.Context, base, next string) error {
	baseBranch, err := r.GetBranch(ctx, base)
	if err != nil {
		return err
	}
	nextBranch, err := r.CreateBranch(ctx, next, branches.Params{
		Salt: baseBranch.Salt,
	})
	if err != nil {
		return err
	}
	if err := branches.SyncVolumes(ctx, nextBranch.Volume, baseBranch.Volume, false); err != nil {
		return err
	}
	return r.SetActiveBranch(ctx, next)
}

func (r *Repo) History(ctx context.Context, name string, fn func(ref Ref, s Snap) error) error {
	branch, err := r.GetBranch(ctx, name)
	if err != nil {
		return err
	}
	return branches.History(ctx, *branch, r.getVCOp(branch), fn)
}

func getActiveBranch(db *bolt.DB) (string, error) {
	var name string
	if err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketDefault))
		if b == nil {
			return nil
		}
		v := b.Get([]byte(keyActive))
		if len(v) > 0 {
			name = string(v)
		}
		return nil
	}); err != nil {
		return "", err
	}
	return name, nil
}

func setActiveBranch(db *bolt.DB, name string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketDefault))
		if err != nil {
			return err
		}
		return b.Put([]byte(keyActive), []byte(name))
	})
}
