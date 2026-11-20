package shard

import (
	"crypto/sha256"
	"io/fs"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/nspcc-dev/neofs-node/pkg/local_object_storage/blobstor/fstree"
	meta "github.com/nspcc-dev/neofs-node/pkg/local_object_storage/metabase"
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

// noSpaceMetabase fails Delete of more than fits objects the way bbolt does
// when its file cannot take the transaction.
type noSpaceMetabase struct {
	metabase
	fits  int
	sizes []int
}

func (x *noSpaceMetabase) Delete(cnr cid.ID, ids []oid.ID) ([]oid.ID, meta.CountersDiff, error) {
	x.sizes = append(x.sizes, len(ids))
	if len(ids) > x.fits {
		return nil, meta.CountersDiff{}, &fs.PathError{Op: "write", Path: "meta", Err: syscall.ENOSPC}
	}
	return x.metabase.Delete(cnr, ids)
}

func TestShard_DeleteNoSpaceInMetabase(t *testing.T) {
	dir := t.TempDir()
	fst := fstree.New(fstree.WithPath(filepath.Join(dir, "fstree")))
	sh := New(
		WithBlobstor(fst),
		WithMetaBaseOptions(
			meta.WithPath(filepath.Join(dir, "meta")),
			meta.WithEpochState(epochState{}),
		),
	)
	require.NoError(t, sh.Open())
	require.NoError(t, sh.Init())
	t.Cleanup(func() { _ = sh.Close() })

	cnr := cidtest.ID()
	put := func(n int) []oid.ID {
		ids := make([]oid.ID, n)
		for i := range ids {
			obj := object.New(cnr, usertest.ID())
			ids[i] = oidtest.ID()
			obj.SetID(ids[i])
			obj.SetPayload([]byte("payload"))
			obj.SetPayloadSize(7)
			obj.SetPayloadChecksum(checksum.NewSHA256(sha256.Sum256(obj.Payload())))
			require.NoError(t, sh.Put(obj, nil))
		}
		return ids
	}
	stored := func(ids []oid.ID) (inMeta, inBlob int) {
		for _, id := range ids {
			addr := oid.NewAddress(cnr, id)
			if ok, err := sh.metaBase.Exists(addr, true); err == nil && ok {
				inMeta++
			}
			if _, err := fst.Get(addr); err == nil {
				inBlob++
			} else {
				require.ErrorIs(t, err, apistatus.ErrObjectNotFound)
			}
		}
		return
	}

	t.Run("smaller transactions fit", func(t *testing.T) {
		ids := put(10)
		mb := &noSpaceMetabase{metabase: sh.metaBase, fits: 3}
		sh.metaBaseIface = mb

		require.NoError(t, sh.Delete(cnr, ids))

		inMeta, inBlob := stored(ids)
		require.Zero(t, inMeta)
		require.Zero(t, inBlob)
		require.Equal(t, []int{10, 5, 2, 3, 5, 2, 3}, mb.sizes)
	})

	t.Run("nothing fits", func(t *testing.T) {
		ids := put(4)
		mb := &noSpaceMetabase{metabase: sh.metaBase}
		sh.metaBaseIface = mb

		require.ErrorIs(t, sh.Delete(cnr, ids), syscall.ENOSPC)

		inMeta, inBlob := stored(ids)
		require.Equal(t, len(ids), inMeta)
		require.Equal(t, len(ids), inBlob)
		require.Equal(t, []int{4, 2, 1, 1, 2, 1, 1}, mb.sizes)
	})

}
