package kv

import (
	"bytes"
	"context"

	types "github.com/prysmaticlabs/eth2-types"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/shared/bytesutil"
	bolt "go.etcd.io/bbolt"
)

var migrationArchivedIndex0Key = []byte("archive_index_0")

func migrateArchivedIndex(tx *bolt.Tx) error {
	mb := tx.Bucket(migrationsBucket)
	if b := mb.Get(migrationArchivedIndex0Key); bytes.Equal(b, migrationCompleted) {
		return nil // Migration already completed.
	}

	bkt := tx.Bucket(archivedRootBucket)
	if bkt == nil {
		return nil
	}
	// Remove "last archived index" key before iterating over all keys.
	if err := bkt.Delete(lastArchivedIndexKey); err != nil {
		return err
	}

	var highest types.Slot
	c := bkt.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		// Look up actual slot from block
		b := tx.Bucket(blocksBucket).Get(v)
		// Skip this key if there is no block for whatever reason.
		if b == nil {
			continue
		}
		blk := &ethpb.SignedBeaconBlock{}
		if err := decode(context.TODO(), b, blk); err != nil {
			return err
		}
		if err := tx.Bucket(stateSlotIndicesBucket).Put(bytesutil.SlotToBytesBigEndian(blk.Block.Slot), v); err != nil {
			return err
		}
		if blk.Block.Slot > highest {
			highest = blk.Block.Slot
		}
	}

	// Delete deprecated buckets.
	for _, bkt := range [][]byte{slotsHasObjectBucket, archivedRootBucket} {
		if tx.Bucket(bkt) != nil {
			if err := tx.DeleteBucket(bkt); err != nil {
				return err
			}
		}
	}

	// Mark migration complete.
	return mb.Put(migrationArchivedIndex0Key, migrationCompleted)
}
