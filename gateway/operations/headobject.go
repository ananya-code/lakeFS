package operations

import (
	"fmt"

	"github.com/treeverse/lakefs/db"
	"github.com/treeverse/lakefs/gateway/errors"
	"github.com/treeverse/lakefs/gateway/serde"
	"github.com/treeverse/lakefs/index/model"
	"github.com/treeverse/lakefs/permissions"

	"golang.org/x/xerrors"
)

type HeadObject struct{}

func (controller *HeadObject) GetArn() string {
	return "arn:treeverse:repos:::{repo}"
}

func (controller *HeadObject) GetPermission() permissions.Permission {
	return permissions.ReadRepo
}

func (controller *HeadObject) Handle(o *PathOperation) {
	entry, err := o.Index.ReadEntry(o.Repo.GetRepoId(), o.Branch, o.Path)
	if xerrors.Is(err, db.ErrNotFound) {
		// TODO: create distinction between missing repo & missing key
		o.Log().
			WithField("path", o.Path).
			WithField("branch", o.Branch).
			WithField("repo", o.Repo.GetRepoId()).
			Warn("path not found")
		o.EncodeError(errors.Codes.ToAPIErr(errors.ErrNoSuchKey))
		return
	}
	if err != nil {
		o.Log().WithError(err).Error("failed querying path")
		o.EncodeError(errors.Codes.ToAPIErr(errors.ErrInternalError))
		return
	}
	if entry.GetType() != model.Entry_OBJECT {
		// only objects should return a successful response
		o.EncodeError(errors.Codes.ToAPIErr(errors.ErrNoSuchKey))
		return
	}
	o.SetHeader("Accept-Ranges", "bytes")
	o.SetHeader("Last-Modified", serde.HeaderTimestamp(entry.GetTimestamp()))
	o.SetHeader("ETag", serde.ETag(entry.GetChecksum()))
	o.SetHeader("Content-Length", fmt.Sprintf("%d", entry.GetSize()))
}