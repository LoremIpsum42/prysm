package beacon

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	eth2types "github.com/prysmaticlabs/eth2-types"
	"github.com/prysmaticlabs/go-bitfield"
	chainMock "github.com/prysmaticlabs/prysm/beacon-chain/blockchain/testing"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/beacon-chain/operations/attestations"
	"github.com/prysmaticlabs/prysm/beacon-chain/operations/slashings"
	"github.com/prysmaticlabs/prysm/beacon-chain/operations/voluntaryexits"
	p2pMock "github.com/prysmaticlabs/prysm/beacon-chain/p2p/testing"
	ethpb "github.com/prysmaticlabs/prysm/proto/eth/v1"
	"github.com/prysmaticlabs/prysm/proto/migration"
	eth "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	statepb "github.com/prysmaticlabs/prysm/proto/prysm/v2/state"
	"github.com/prysmaticlabs/prysm/shared/bls"
	"github.com/prysmaticlabs/prysm/shared/bytesutil"
	"github.com/prysmaticlabs/prysm/shared/grpcutils"
	"github.com/prysmaticlabs/prysm/shared/params"
	"github.com/prysmaticlabs/prysm/shared/testutil"
	"github.com/prysmaticlabs/prysm/shared/testutil/assert"
	"github.com/prysmaticlabs/prysm/shared/testutil/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestListPoolAttestations(t *testing.T) {
	state, err := testutil.NewBeaconState()
	require.NoError(t, err)
	att1 := &eth.Attestation{
		AggregationBits: []byte{1, 10},
		Data: &eth.AttestationData{
			Slot:            1,
			CommitteeIndex:  1,
			BeaconBlockRoot: bytesutil.PadTo([]byte("blockroot1"), 32),
			Source: &eth.Checkpoint{
				Epoch: 1,
				Root:  bytesutil.PadTo([]byte("sourceroot1"), 32),
			},
			Target: &eth.Checkpoint{
				Epoch: 10,
				Root:  bytesutil.PadTo([]byte("targetroot1"), 32),
			},
		},
		Signature: bytesutil.PadTo([]byte("signature1"), 96),
	}
	att2 := &eth.Attestation{
		AggregationBits: []byte{4, 40},
		Data: &eth.AttestationData{
			Slot:            4,
			CommitteeIndex:  4,
			BeaconBlockRoot: bytesutil.PadTo([]byte("blockroot4"), 32),
			Source: &eth.Checkpoint{
				Epoch: 4,
				Root:  bytesutil.PadTo([]byte("sourceroot4"), 32),
			},
			Target: &eth.Checkpoint{
				Epoch: 40,
				Root:  bytesutil.PadTo([]byte("targetroot4"), 32),
			},
		},
		Signature: bytesutil.PadTo([]byte("signature4"), 96),
	}
	att3 := &eth.Attestation{
		AggregationBits: []byte{2, 20},
		Data: &eth.AttestationData{
			Slot:            2,
			CommitteeIndex:  2,
			BeaconBlockRoot: bytesutil.PadTo([]byte("blockroot2"), 32),
			Source: &eth.Checkpoint{
				Epoch: 2,
				Root:  bytesutil.PadTo([]byte("sourceroot2"), 32),
			},
			Target: &eth.Checkpoint{
				Epoch: 20,
				Root:  bytesutil.PadTo([]byte("targetroot2"), 32),
			},
		},
		Signature: bytesutil.PadTo([]byte("signature2"), 96),
	}
	att4 := &eth.Attestation{
		AggregationBits: bitfield.NewBitlist(8),
		Data: &eth.AttestationData{
			Slot:            4,
			CommitteeIndex:  4,
			BeaconBlockRoot: bytesutil.PadTo([]byte("blockroot2"), 32),
			Source: &eth.Checkpoint{
				Epoch: 2,
				Root:  bytesutil.PadTo([]byte("sourceroot2"), 32),
			},
			Target: &eth.Checkpoint{
				Epoch: 20,
				Root:  bytesutil.PadTo([]byte("targetroot2"), 32),
			},
		},
		Signature: bytesutil.PadTo([]byte("signature2"), 96),
	}
	att5 := &eth.Attestation{
		AggregationBits: bitfield.NewBitlist(8),
		Data: &eth.AttestationData{
			Slot:            2,
			CommitteeIndex:  4,
			BeaconBlockRoot: bytesutil.PadTo([]byte("blockroot1"), 32),
			Source: &eth.Checkpoint{
				Epoch: 2,
				Root:  bytesutil.PadTo([]byte("sourceroot2"), 32),
			},
			Target: &eth.Checkpoint{
				Epoch: 20,
				Root:  bytesutil.PadTo([]byte("targetroot2"), 32),
			},
		},
		Signature: bytesutil.PadTo([]byte("signature1"), 96),
	}
	att6 := &eth.Attestation{
		AggregationBits: bitfield.NewBitlist(8),
		Data: &eth.AttestationData{
			Slot:            2,
			CommitteeIndex:  4,
			BeaconBlockRoot: bytesutil.PadTo([]byte("blockroot2"), 32),
			Source: &eth.Checkpoint{
				Epoch: 2,
				Root:  bytesutil.PadTo([]byte("sourceroot2"), 32),
			},
			Target: &eth.Checkpoint{
				Epoch: 20,
				Root:  bytesutil.PadTo([]byte("targetroot2"), 32),
			},
		},
		Signature: bytesutil.PadTo([]byte("signature2"), 96),
	}
	s := &Server{
		ChainInfoFetcher: &chainMock.ChainService{State: state},
		AttestationsPool: attestations.NewPool(),
	}
	require.NoError(t, s.AttestationsPool.SaveAggregatedAttestations([]*eth.Attestation{att1, att2, att3}))
	require.NoError(t, s.AttestationsPool.SaveUnaggregatedAttestations([]*eth.Attestation{att4, att5, att6}))

	t.Run("empty request", func(t *testing.T) {
		req := &ethpb.AttestationsPoolRequest{}
		resp, err := s.ListPoolAttestations(context.Background(), req)
		require.NoError(t, err)
		require.Equal(t, 6, len(resp.Data))
	})

	t.Run("slot request", func(t *testing.T) {
		slot := eth2types.Slot(2)
		req := &ethpb.AttestationsPoolRequest{
			Slot: &slot,
		}
		resp, err := s.ListPoolAttestations(context.Background(), req)
		require.NoError(t, err)
		require.Equal(t, 3, len(resp.Data))
		for _, datum := range resp.Data {
			assert.DeepEqual(t, datum.Data.Slot, slot)
		}
	})

	t.Run("index request", func(t *testing.T) {
		index := eth2types.CommitteeIndex(4)
		req := &ethpb.AttestationsPoolRequest{
			CommitteeIndex: &index,
		}
		resp, err := s.ListPoolAttestations(context.Background(), req)
		require.NoError(t, err)
		require.Equal(t, 4, len(resp.Data))
		for _, datum := range resp.Data {
			assert.DeepEqual(t, datum.Data.Index, index)
		}
	})

	t.Run("both slot + index request", func(t *testing.T) {
		slot := eth2types.Slot(2)
		index := eth2types.CommitteeIndex(4)
		req := &ethpb.AttestationsPoolRequest{
			Slot:           &slot,
			CommitteeIndex: &index,
		}
		resp, err := s.ListPoolAttestations(context.Background(), req)
		require.NoError(t, err)
		require.Equal(t, 2, len(resp.Data))
		for _, datum := range resp.Data {
			assert.DeepEqual(t, datum.Data.Index, index)
			assert.DeepEqual(t, datum.Data.Slot, slot)
		}
	})
}

func TestListPoolAttesterSlashings(t *testing.T) {
	state, err := testutil.NewBeaconState()
	require.NoError(t, err)
	slashing1 := &eth.AttesterSlashing{
		Attestation_1: &eth.IndexedAttestation{
			AttestingIndices: []uint64{1, 10},
			Data: &eth.AttestationData{
				Slot:            1,
				CommitteeIndex:  1,
				BeaconBlockRoot: bytesutil.PadTo([]byte("blockroot1"), 32),
				Source: &eth.Checkpoint{
					Epoch: 1,
					Root:  bytesutil.PadTo([]byte("sourceroot1"), 32),
				},
				Target: &eth.Checkpoint{
					Epoch: 10,
					Root:  bytesutil.PadTo([]byte("targetroot1"), 32),
				},
			},
			Signature: bytesutil.PadTo([]byte("signature1"), 96),
		},
		Attestation_2: &eth.IndexedAttestation{
			AttestingIndices: []uint64{2, 20},
			Data: &eth.AttestationData{
				Slot:            2,
				CommitteeIndex:  2,
				BeaconBlockRoot: bytesutil.PadTo([]byte("blockroot2"), 32),
				Source: &eth.Checkpoint{
					Epoch: 2,
					Root:  bytesutil.PadTo([]byte("sourceroot2"), 32),
				},
				Target: &eth.Checkpoint{
					Epoch: 20,
					Root:  bytesutil.PadTo([]byte("targetroot2"), 32),
				},
			},
			Signature: bytesutil.PadTo([]byte("signature2"), 96),
		},
	}
	slashing2 := &eth.AttesterSlashing{
		Attestation_1: &eth.IndexedAttestation{
			AttestingIndices: []uint64{3, 30},
			Data: &eth.AttestationData{
				Slot:            3,
				CommitteeIndex:  3,
				BeaconBlockRoot: bytesutil.PadTo([]byte("blockroot3"), 32),
				Source: &eth.Checkpoint{
					Epoch: 3,
					Root:  bytesutil.PadTo([]byte("sourceroot3"), 32),
				},
				Target: &eth.Checkpoint{
					Epoch: 30,
					Root:  bytesutil.PadTo([]byte("targetroot3"), 32),
				},
			},
			Signature: bytesutil.PadTo([]byte("signature3"), 96),
		},
		Attestation_2: &eth.IndexedAttestation{
			AttestingIndices: []uint64{4, 40},
			Data: &eth.AttestationData{
				Slot:            4,
				CommitteeIndex:  4,
				BeaconBlockRoot: bytesutil.PadTo([]byte("blockroot4"), 32),
				Source: &eth.Checkpoint{
					Epoch: 4,
					Root:  bytesutil.PadTo([]byte("sourceroot4"), 32),
				},
				Target: &eth.Checkpoint{
					Epoch: 40,
					Root:  bytesutil.PadTo([]byte("targetroot4"), 32),
				},
			},
			Signature: bytesutil.PadTo([]byte("signature4"), 96),
		},
	}

	s := &Server{
		ChainInfoFetcher: &chainMock.ChainService{State: state},
		SlashingsPool:    &slashings.PoolMock{PendingAttSlashings: []*eth.AttesterSlashing{slashing1, slashing2}},
	}

	resp, err := s.ListPoolAttesterSlashings(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
	require.Equal(t, 2, len(resp.Data))
	assert.DeepEqual(t, migration.V1Alpha1AttSlashingToV1(slashing1), resp.Data[0])
	assert.DeepEqual(t, migration.V1Alpha1AttSlashingToV1(slashing2), resp.Data[1])
}

func TestListPoolProposerSlashings(t *testing.T) {
	state, err := testutil.NewBeaconState()
	require.NoError(t, err)
	slashing1 := &eth.ProposerSlashing{
		Header_1: &eth.SignedBeaconBlockHeader{
			Header: &eth.BeaconBlockHeader{
				Slot:          1,
				ProposerIndex: 1,
				ParentRoot:    bytesutil.PadTo([]byte("parentroot1"), 32),
				StateRoot:     bytesutil.PadTo([]byte("stateroot1"), 32),
				BodyRoot:      bytesutil.PadTo([]byte("bodyroot1"), 32),
			},
			Signature: bytesutil.PadTo([]byte("signature1"), 96),
		},
		Header_2: &eth.SignedBeaconBlockHeader{
			Header: &eth.BeaconBlockHeader{
				Slot:          2,
				ProposerIndex: 2,
				ParentRoot:    bytesutil.PadTo([]byte("parentroot2"), 32),
				StateRoot:     bytesutil.PadTo([]byte("stateroot2"), 32),
				BodyRoot:      bytesutil.PadTo([]byte("bodyroot2"), 32),
			},
			Signature: bytesutil.PadTo([]byte("signature2"), 96),
		},
	}
	slashing2 := &eth.ProposerSlashing{
		Header_1: &eth.SignedBeaconBlockHeader{
			Header: &eth.BeaconBlockHeader{
				Slot:          3,
				ProposerIndex: 3,
				ParentRoot:    bytesutil.PadTo([]byte("parentroot3"), 32),
				StateRoot:     bytesutil.PadTo([]byte("stateroot3"), 32),
				BodyRoot:      bytesutil.PadTo([]byte("bodyroot3"), 32),
			},
			Signature: bytesutil.PadTo([]byte("signature3"), 96),
		},
		Header_2: &eth.SignedBeaconBlockHeader{
			Header: &eth.BeaconBlockHeader{
				Slot:          4,
				ProposerIndex: 4,
				ParentRoot:    bytesutil.PadTo([]byte("parentroot4"), 32),
				StateRoot:     bytesutil.PadTo([]byte("stateroot4"), 32),
				BodyRoot:      bytesutil.PadTo([]byte("bodyroot4"), 32),
			},
			Signature: bytesutil.PadTo([]byte("signature4"), 96),
		},
	}

	s := &Server{
		ChainInfoFetcher: &chainMock.ChainService{State: state},
		SlashingsPool:    &slashings.PoolMock{PendingPropSlashings: []*eth.ProposerSlashing{slashing1, slashing2}},
	}

	resp, err := s.ListPoolProposerSlashings(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
	require.Equal(t, 2, len(resp.Data))
	assert.DeepEqual(t, migration.V1Alpha1ProposerSlashingToV1(slashing1), resp.Data[0])
	assert.DeepEqual(t, migration.V1Alpha1ProposerSlashingToV1(slashing2), resp.Data[1])
}

func TestListPoolVoluntaryExits(t *testing.T) {
	state, err := testutil.NewBeaconState()
	require.NoError(t, err)
	exit1 := &eth.SignedVoluntaryExit{
		Exit: &eth.VoluntaryExit{
			Epoch:          1,
			ValidatorIndex: 1,
		},
		Signature: bytesutil.PadTo([]byte("signature1"), 96),
	}
	exit2 := &eth.SignedVoluntaryExit{
		Exit: &eth.VoluntaryExit{
			Epoch:          2,
			ValidatorIndex: 2,
		},
		Signature: bytesutil.PadTo([]byte("signature2"), 96),
	}

	s := &Server{
		ChainInfoFetcher:   &chainMock.ChainService{State: state},
		VoluntaryExitsPool: &voluntaryexits.PoolMock{Exits: []*eth.SignedVoluntaryExit{exit1, exit2}},
	}

	resp, err := s.ListPoolVoluntaryExits(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
	require.Equal(t, 2, len(resp.Data))
	assert.DeepEqual(t, migration.V1Alpha1ExitToV1(exit1), resp.Data[0])
	assert.DeepEqual(t, migration.V1Alpha1ExitToV1(exit2), resp.Data[1])
}

func TestSubmitAttesterSlashing_Ok(t *testing.T) {
	ctx := context.Background()

	_, keys, err := testutil.DeterministicDepositsAndKeys(1)
	require.NoError(t, err)
	validator := &eth.Validator{
		PublicKey: keys[0].PublicKey().Marshal(),
	}
	state, err := testutil.NewBeaconState(func(state *statepb.BeaconState) error {
		state.Validators = []*eth.Validator{validator}
		return nil
	})
	require.NoError(t, err)

	slashing := &ethpb.AttesterSlashing{
		Attestation_1: &ethpb.IndexedAttestation{
			AttestingIndices: []uint64{0},
			Data: &ethpb.AttestationData{
				Slot:            1,
				Index:           1,
				BeaconBlockRoot: bytesutil.PadTo([]byte("blockroot1"), 32),
				Source: &ethpb.Checkpoint{
					Epoch: 1,
					Root:  bytesutil.PadTo([]byte("sourceroot1"), 32),
				},
				Target: &ethpb.Checkpoint{
					Epoch: 10,
					Root:  bytesutil.PadTo([]byte("targetroot1"), 32),
				},
			},
			Signature: make([]byte, 96),
		},
		Attestation_2: &ethpb.IndexedAttestation{
			AttestingIndices: []uint64{0},
			Data: &ethpb.AttestationData{
				Slot:            1,
				Index:           1,
				BeaconBlockRoot: bytesutil.PadTo([]byte("blockroot2"), 32),
				Source: &ethpb.Checkpoint{
					Epoch: 1,
					Root:  bytesutil.PadTo([]byte("sourceroot2"), 32),
				},
				Target: &ethpb.Checkpoint{
					Epoch: 10,
					Root:  bytesutil.PadTo([]byte("targetroot2"), 32),
				},
			},
			Signature: make([]byte, 96),
		},
	}

	for _, att := range []*ethpb.IndexedAttestation{slashing.Attestation_1, slashing.Attestation_2} {
		sb, err := helpers.ComputeDomainAndSign(state, att.Data.Target.Epoch, att.Data, params.BeaconConfig().DomainBeaconAttester, keys[0])
		require.NoError(t, err)
		sig, err := bls.SignatureFromBytes(sb)
		require.NoError(t, err)
		att.Signature = sig.Marshal()
	}

	broadcaster := &p2pMock.MockBroadcaster{}
	s := &Server{
		ChainInfoFetcher: &chainMock.ChainService{State: state},
		SlashingsPool:    &slashings.PoolMock{},
		Broadcaster:      broadcaster,
	}

	_, err = s.SubmitAttesterSlashing(ctx, slashing)
	require.NoError(t, err)
	pendingSlashings := s.SlashingsPool.PendingAttesterSlashings(ctx, state, true)
	require.Equal(t, 1, len(pendingSlashings))
	assert.DeepEqual(t, migration.V1AttSlashingToV1Alpha1(slashing), pendingSlashings[0])
	assert.Equal(t, true, broadcaster.BroadcastCalled)
}

func TestSubmitAttesterSlashing_InvalidSlashing(t *testing.T) {
	ctx := context.Background()
	state, err := testutil.NewBeaconState()
	require.NoError(t, err)

	attestation := &ethpb.IndexedAttestation{
		AttestingIndices: []uint64{0},
		Data: &ethpb.AttestationData{
			Slot:            1,
			Index:           1,
			BeaconBlockRoot: bytesutil.PadTo([]byte("blockroot1"), 32),
			Source: &ethpb.Checkpoint{
				Epoch: 1,
				Root:  bytesutil.PadTo([]byte("sourceroot1"), 32),
			},
			Target: &ethpb.Checkpoint{
				Epoch: 10,
				Root:  bytesutil.PadTo([]byte("targetroot1"), 32),
			},
		},
		Signature: make([]byte, 96),
	}

	slashing := &ethpb.AttesterSlashing{
		Attestation_1: attestation,
		Attestation_2: attestation,
	}

	broadcaster := &p2pMock.MockBroadcaster{}
	s := &Server{
		ChainInfoFetcher: &chainMock.ChainService{State: state},
		SlashingsPool:    &slashings.PoolMock{},
		Broadcaster:      broadcaster,
	}

	_, err = s.SubmitAttesterSlashing(ctx, slashing)
	require.ErrorContains(t, "Invalid attester slashing", err)
	assert.Equal(t, false, broadcaster.BroadcastCalled)
}

func TestSubmitProposerSlashing_Ok(t *testing.T) {
	ctx := context.Background()

	_, keys, err := testutil.DeterministicDepositsAndKeys(1)
	require.NoError(t, err)
	validator := &eth.Validator{
		PublicKey:         keys[0].PublicKey().Marshal(),
		WithdrawableEpoch: eth2types.Epoch(1),
	}
	state, err := testutil.NewBeaconState(func(state *statepb.BeaconState) error {
		state.Validators = []*eth.Validator{validator}
		return nil
	})
	require.NoError(t, err)

	slashing := &ethpb.ProposerSlashing{
		SignedHeader_1: &ethpb.SignedBeaconBlockHeader{
			Message: &ethpb.BeaconBlockHeader{
				Slot:          1,
				ProposerIndex: 0,
				ParentRoot:    bytesutil.PadTo([]byte("parentroot1"), 32),
				StateRoot:     bytesutil.PadTo([]byte("stateroot1"), 32),
				BodyRoot:      bytesutil.PadTo([]byte("bodyroot1"), 32),
			},
			Signature: make([]byte, 96),
		},
		SignedHeader_2: &ethpb.SignedBeaconBlockHeader{
			Message: &ethpb.BeaconBlockHeader{
				Slot:          1,
				ProposerIndex: 0,
				ParentRoot:    bytesutil.PadTo([]byte("parentroot2"), 32),
				StateRoot:     bytesutil.PadTo([]byte("stateroot2"), 32),
				BodyRoot:      bytesutil.PadTo([]byte("bodyroot2"), 32),
			},
			Signature: make([]byte, 96),
		},
	}

	for _, h := range []*ethpb.SignedBeaconBlockHeader{slashing.SignedHeader_1, slashing.SignedHeader_2} {
		sb, err := helpers.ComputeDomainAndSign(
			state,
			helpers.SlotToEpoch(h.Message.Slot),
			h.Message,
			params.BeaconConfig().DomainBeaconProposer,
			keys[0],
		)
		require.NoError(t, err)
		sig, err := bls.SignatureFromBytes(sb)
		require.NoError(t, err)
		h.Signature = sig.Marshal()
	}

	broadcaster := &p2pMock.MockBroadcaster{}
	s := &Server{
		ChainInfoFetcher: &chainMock.ChainService{State: state},
		SlashingsPool:    &slashings.PoolMock{},
		Broadcaster:      broadcaster,
	}

	_, err = s.SubmitProposerSlashing(ctx, slashing)
	require.NoError(t, err)
	pendingSlashings := s.SlashingsPool.PendingProposerSlashings(ctx, state, true)
	require.Equal(t, 1, len(pendingSlashings))
	assert.DeepEqual(t, migration.V1ProposerSlashingToV1Alpha1(slashing), pendingSlashings[0])
	assert.Equal(t, true, broadcaster.BroadcastCalled)
}

func TestSubmitProposerSlashing_InvalidSlashing(t *testing.T) {
	ctx := context.Background()
	state, err := testutil.NewBeaconState()
	require.NoError(t, err)

	header := &ethpb.SignedBeaconBlockHeader{
		Message: &ethpb.BeaconBlockHeader{
			Slot:          1,
			ProposerIndex: 0,
			ParentRoot:    bytesutil.PadTo([]byte("parentroot1"), 32),
			StateRoot:     bytesutil.PadTo([]byte("stateroot1"), 32),
			BodyRoot:      bytesutil.PadTo([]byte("bodyroot1"), 32),
		},
		Signature: make([]byte, 96),
	}

	slashing := &ethpb.ProposerSlashing{
		SignedHeader_1: header,
		SignedHeader_2: header,
	}

	broadcaster := &p2pMock.MockBroadcaster{}
	s := &Server{
		ChainInfoFetcher: &chainMock.ChainService{State: state},
		SlashingsPool:    &slashings.PoolMock{},
		Broadcaster:      broadcaster,
	}

	_, err = s.SubmitProposerSlashing(ctx, slashing)
	require.ErrorContains(t, "Invalid proposer slashing", err)
	assert.Equal(t, false, broadcaster.BroadcastCalled)
}

func TestSubmitVoluntaryExit_Ok(t *testing.T) {
	ctx := context.Background()

	_, keys, err := testutil.DeterministicDepositsAndKeys(1)
	require.NoError(t, err)
	validator := &eth.Validator{
		ExitEpoch: params.BeaconConfig().FarFutureEpoch,
		PublicKey: keys[0].PublicKey().Marshal(),
	}
	state, err := testutil.NewBeaconState(func(state *statepb.BeaconState) error {
		state.Validators = []*eth.Validator{validator}
		// Satisfy activity time required before exiting.
		state.Slot = params.BeaconConfig().SlotsPerEpoch.Mul(uint64(params.BeaconConfig().ShardCommitteePeriod))
		return nil
	})
	require.NoError(t, err)

	exit := &ethpb.SignedVoluntaryExit{
		Message: &ethpb.VoluntaryExit{
			Epoch:          0,
			ValidatorIndex: 0,
		},
		Signature: make([]byte, 96),
	}

	sb, err := helpers.ComputeDomainAndSign(state, exit.Message.Epoch, exit.Message, params.BeaconConfig().DomainVoluntaryExit, keys[0])
	require.NoError(t, err)
	sig, err := bls.SignatureFromBytes(sb)
	require.NoError(t, err)
	exit.Signature = sig.Marshal()

	broadcaster := &p2pMock.MockBroadcaster{}
	s := &Server{
		ChainInfoFetcher:   &chainMock.ChainService{State: state},
		VoluntaryExitsPool: &voluntaryexits.PoolMock{},
		Broadcaster:        broadcaster,
	}

	_, err = s.SubmitVoluntaryExit(ctx, exit)
	require.NoError(t, err)
	pendingExits := s.VoluntaryExitsPool.PendingExits(state, state.Slot(), true)
	require.Equal(t, 1, len(pendingExits))
	assert.DeepEqual(t, migration.V1ExitToV1Alpha1(exit), pendingExits[0])
	assert.Equal(t, true, broadcaster.BroadcastCalled)
}

func TestSubmitVoluntaryExit_InvalidValidatorIndex(t *testing.T) {
	ctx := context.Background()

	_, keys, err := testutil.DeterministicDepositsAndKeys(1)
	require.NoError(t, err)
	validator := &eth.Validator{
		ExitEpoch: params.BeaconConfig().FarFutureEpoch,
		PublicKey: keys[0].PublicKey().Marshal(),
	}
	state, err := testutil.NewBeaconState(func(state *statepb.BeaconState) error {
		state.Validators = []*eth.Validator{validator}
		return nil
	})
	require.NoError(t, err)

	exit := &ethpb.SignedVoluntaryExit{
		Message: &ethpb.VoluntaryExit{
			Epoch:          0,
			ValidatorIndex: 99,
		},
		Signature: make([]byte, 96),
	}

	broadcaster := &p2pMock.MockBroadcaster{}
	s := &Server{
		ChainInfoFetcher:   &chainMock.ChainService{State: state},
		VoluntaryExitsPool: &voluntaryexits.PoolMock{},
		Broadcaster:        broadcaster,
	}

	_, err = s.SubmitVoluntaryExit(ctx, exit)
	require.ErrorContains(t, "Could not get exiting validator", err)
	assert.Equal(t, false, broadcaster.BroadcastCalled)
}

func TestSubmitVoluntaryExit_InvalidExit(t *testing.T) {
	ctx := context.Background()

	_, keys, err := testutil.DeterministicDepositsAndKeys(1)
	require.NoError(t, err)
	validator := &eth.Validator{
		ExitEpoch: params.BeaconConfig().FarFutureEpoch,
		PublicKey: keys[0].PublicKey().Marshal(),
	}
	state, err := testutil.NewBeaconState(func(state *statepb.BeaconState) error {
		state.Validators = []*eth.Validator{validator}
		return nil
	})
	require.NoError(t, err)

	exit := &ethpb.SignedVoluntaryExit{
		Message: &ethpb.VoluntaryExit{
			Epoch:          0,
			ValidatorIndex: 0,
		},
		Signature: make([]byte, 96),
	}

	broadcaster := &p2pMock.MockBroadcaster{}
	s := &Server{
		ChainInfoFetcher:   &chainMock.ChainService{State: state},
		VoluntaryExitsPool: &voluntaryexits.PoolMock{},
		Broadcaster:        broadcaster,
	}

	_, err = s.SubmitVoluntaryExit(ctx, exit)
	require.ErrorContains(t, "Invalid voluntary exit", err)
	assert.Equal(t, false, broadcaster.BroadcastCalled)
}

func TestServer_SubmitAttestations_Ok(t *testing.T) {
	ctx := context.Background()
	params.SetupTestConfigCleanup(t)
	c := params.BeaconConfig()
	// Required for correct committee size calculation.
	c.SlotsPerEpoch = 1
	params.OverrideBeaconConfig(c)

	_, keys, err := testutil.DeterministicDepositsAndKeys(1)
	require.NoError(t, err)
	validators := []*eth.Validator{
		{
			PublicKey: keys[0].PublicKey().Marshal(),
			ExitEpoch: params.BeaconConfig().FarFutureEpoch,
		},
	}
	state, err := testutil.NewBeaconState(func(state *statepb.BeaconState) error {
		state.Validators = validators
		state.Slot = 1
		state.PreviousJustifiedCheckpoint = &eth.Checkpoint{
			Epoch: 0,
			Root:  bytesutil.PadTo([]byte("sourceroot1"), 32),
		}
		return nil
	})
	require.NoError(t, err)
	b := bitfield.NewBitlist(1)
	b.SetBitAt(0, true)

	sourceCheckpoint := &ethpb.Checkpoint{
		Epoch: 0,
		Root:  bytesutil.PadTo([]byte("sourceroot1"), 32),
	}
	att1 := &ethpb.Attestation{
		AggregationBits: b,
		Data: &ethpb.AttestationData{
			Slot:            0,
			Index:           0,
			BeaconBlockRoot: bytesutil.PadTo([]byte("beaconblockroot1"), 32),
			Source:          sourceCheckpoint,
			Target: &ethpb.Checkpoint{
				Epoch: 0,
				Root:  bytesutil.PadTo([]byte("targetroot1"), 32),
			},
		},
		Signature: make([]byte, 96),
	}
	att2 := &ethpb.Attestation{
		AggregationBits: b,
		Data: &ethpb.AttestationData{
			Slot:            0,
			Index:           0,
			BeaconBlockRoot: bytesutil.PadTo([]byte("beaconblockroot2"), 32),
			Source:          sourceCheckpoint,
			Target: &ethpb.Checkpoint{
				Epoch: 0,
				Root:  bytesutil.PadTo([]byte("targetroot2"), 32),
			},
		},
		Signature: make([]byte, 96),
	}

	for _, att := range []*ethpb.Attestation{att1, att2} {
		sb, err := helpers.ComputeDomainAndSign(
			state,
			helpers.SlotToEpoch(att.Data.Slot),
			att.Data,
			params.BeaconConfig().DomainBeaconAttester,
			keys[0],
		)
		require.NoError(t, err)
		sig, err := bls.SignatureFromBytes(sb)
		require.NoError(t, err)
		att.Signature = sig.Marshal()
	}

	broadcaster := &p2pMock.MockBroadcaster{}
	s := &Server{
		ChainInfoFetcher: &chainMock.ChainService{State: state},
		AttestationsPool: &attestations.PoolMock{},
		Broadcaster:      broadcaster,
	}

	_, err = s.SubmitAttestations(ctx, &ethpb.SubmitAttestationsRequest{
		Data: []*ethpb.Attestation{att1, att2},
	})
	require.NoError(t, err)
	savedAtts := s.AttestationsPool.AggregatedAttestations()
	require.Equal(t, 2, len(savedAtts))
	expectedAtt1, err := att1.HashTreeRoot()
	require.NoError(t, err)
	expectedAtt2, err := att2.HashTreeRoot()
	require.NoError(t, err)
	actualAtt1, err := savedAtts[0].HashTreeRoot()
	require.NoError(t, err)
	actualAtt2, err := savedAtts[1].HashTreeRoot()
	require.NoError(t, err)
	for _, r := range [][32]byte{actualAtt1, actualAtt2} {
		assert.Equal(t, true, reflect.DeepEqual(expectedAtt1, r) || reflect.DeepEqual(expectedAtt2, r))
	}
	assert.Equal(t, true, broadcaster.BroadcastCalled)
	assert.Equal(t, 2, len(broadcaster.BroadcastMessages))
}

func TestServer_SubmitAttestations_ValidAttestationSubmitted(t *testing.T) {
	ctx := grpc.NewContextWithServerTransportStream(context.Background(), &runtime.ServerTransportStream{})

	params.SetupTestConfigCleanup(t)
	c := params.BeaconConfig()
	// Required for correct committee size calculation.
	c.SlotsPerEpoch = 1
	params.OverrideBeaconConfig(c)

	_, keys, err := testutil.DeterministicDepositsAndKeys(1)
	require.NoError(t, err)
	validators := []*eth.Validator{
		{
			PublicKey: keys[0].PublicKey().Marshal(),
			ExitEpoch: params.BeaconConfig().FarFutureEpoch,
		},
	}
	state, err := testutil.NewBeaconState(func(state *statepb.BeaconState) error {
		state.Validators = validators
		state.Slot = 1
		state.PreviousJustifiedCheckpoint = &eth.Checkpoint{
			Epoch: 0,
			Root:  bytesutil.PadTo([]byte("sourceroot1"), 32),
		}
		return nil
	})

	require.NoError(t, err)

	sourceCheckpoint := &ethpb.Checkpoint{
		Epoch: 0,
		Root:  bytesutil.PadTo([]byte("sourceroot1"), 32),
	}
	b := bitfield.NewBitlist(1)
	b.SetBitAt(0, true)
	attValid := &ethpb.Attestation{
		AggregationBits: b,
		Data: &ethpb.AttestationData{
			Slot:            0,
			Index:           0,
			BeaconBlockRoot: bytesutil.PadTo([]byte("beaconblockroot1"), 32),
			Source:          sourceCheckpoint,
			Target: &ethpb.Checkpoint{
				Epoch: 0,
				Root:  bytesutil.PadTo([]byte("targetroot1"), 32),
			},
		},
		Signature: make([]byte, 96),
	}
	attInvalidTarget := &ethpb.Attestation{
		AggregationBits: b,
		Data: &ethpb.AttestationData{
			Slot:            0,
			Index:           0,
			BeaconBlockRoot: bytesutil.PadTo([]byte("beaconblockroot2"), 32),
			Source: &ethpb.Checkpoint{
				Epoch: 0,
				Root:  bytesutil.PadTo([]byte("sourceroot2"), 32),
			},
			Target: &ethpb.Checkpoint{
				Epoch: 99,
				Root:  bytesutil.PadTo([]byte("targetroot2"), 32),
			},
		},
		Signature: make([]byte, 96),
	}
	attInvalidSignature := &ethpb.Attestation{
		AggregationBits: b,
		Data: &ethpb.AttestationData{
			Slot:            0,
			Index:           0,
			BeaconBlockRoot: bytesutil.PadTo([]byte("beaconblockroot2"), 32),
			Source:          sourceCheckpoint,
			Target: &ethpb.Checkpoint{
				Epoch: 0,
				Root:  bytesutil.PadTo([]byte("targetroot2"), 32),
			},
		},
		Signature: make([]byte, 96),
	}

	// Don't sign attInvalidSignature.
	for _, att := range []*ethpb.Attestation{attValid, attInvalidTarget} {
		sb, err := helpers.ComputeDomainAndSign(
			state,
			helpers.SlotToEpoch(att.Data.Slot),
			att.Data,
			params.BeaconConfig().DomainBeaconAttester,
			keys[0],
		)
		require.NoError(t, err)
		sig, err := bls.SignatureFromBytes(sb)
		require.NoError(t, err)
		att.Signature = sig.Marshal()
	}

	broadcaster := &p2pMock.MockBroadcaster{}
	s := &Server{
		ChainInfoFetcher: &chainMock.ChainService{State: state},
		AttestationsPool: &attestations.PoolMock{},
		Broadcaster:      broadcaster,
	}

	_, err = s.SubmitAttestations(ctx, &ethpb.SubmitAttestationsRequest{
		Data: []*ethpb.Attestation{attValid, attInvalidTarget, attInvalidSignature},
	})
	require.ErrorContains(t, "One or more attestations failed validation", err)
	savedAtts := s.AttestationsPool.AggregatedAttestations()
	require.Equal(t, 1, len(savedAtts))
	expectedAtt, err := attValid.HashTreeRoot()
	require.NoError(t, err)
	actualAtt, err := savedAtts[0].HashTreeRoot()
	require.NoError(t, err)
	assert.DeepEqual(t, expectedAtt, actualAtt)
	assert.Equal(t, true, broadcaster.BroadcastCalled)
	require.Equal(t, 1, len(broadcaster.BroadcastMessages))
	broadcastRoot, err := broadcaster.BroadcastMessages[0].(*eth.Attestation).HashTreeRoot()
	require.NoError(t, err)
	require.DeepEqual(t, expectedAtt, broadcastRoot)
}

func TestServer_SubmitAttestations_InvalidAttestationHeader(t *testing.T) {
	ctx := grpc.NewContextWithServerTransportStream(context.Background(), &runtime.ServerTransportStream{})

	params.SetupTestConfigCleanup(t)
	c := params.BeaconConfig()
	// Required for correct committee size calculation.
	c.SlotsPerEpoch = 1
	params.OverrideBeaconConfig(c)

	_, keys, err := testutil.DeterministicDepositsAndKeys(1)
	require.NoError(t, err)
	validators := []*eth.Validator{
		{
			PublicKey: keys[0].PublicKey().Marshal(),
			ExitEpoch: params.BeaconConfig().FarFutureEpoch,
		},
	}
	state, err := testutil.NewBeaconState(func(state *statepb.BeaconState) error {
		state.Validators = validators
		state.Slot = 1
		state.PreviousJustifiedCheckpoint = &eth.Checkpoint{
			Epoch: 0,
			Root:  bytesutil.PadTo([]byte("sourceroot1"), 32),
		}
		return nil
	})

	require.NoError(t, err)

	b := bitfield.NewBitlist(1)
	b.SetBitAt(0, true)
	att := &ethpb.Attestation{
		AggregationBits: b,
		Data: &ethpb.AttestationData{
			Slot:            0,
			Index:           0,
			BeaconBlockRoot: bytesutil.PadTo([]byte("beaconblockroot2"), 32),
			Source: &ethpb.Checkpoint{
				Epoch: 0,
				Root:  bytesutil.PadTo([]byte("sourceroot2"), 32),
			},
			Target: &ethpb.Checkpoint{
				Epoch: 99,
				Root:  bytesutil.PadTo([]byte("targetroot2"), 32),
			},
		},
		Signature: make([]byte, 96),
	}

	sb, err := helpers.ComputeDomainAndSign(
		state,
		helpers.SlotToEpoch(att.Data.Slot),
		att.Data,
		params.BeaconConfig().DomainBeaconAttester,
		keys[0],
	)
	require.NoError(t, err)
	sig, err := bls.SignatureFromBytes(sb)
	require.NoError(t, err)
	att.Signature = sig.Marshal()

	broadcaster := &p2pMock.MockBroadcaster{}
	s := &Server{
		ChainInfoFetcher: &chainMock.ChainService{State: state},
		AttestationsPool: &attestations.PoolMock{},
		Broadcaster:      broadcaster,
	}

	_, err = s.SubmitAttestations(ctx, &ethpb.SubmitAttestationsRequest{
		Data: []*ethpb.Attestation{att},
	})
	require.ErrorContains(t, "One or more attestations failed validation", err)
	sts, ok := grpc.ServerTransportStreamFromContext(ctx).(*runtime.ServerTransportStream)
	require.Equal(t, true, ok, "type assertion failed")
	md := sts.Header()
	v, ok := md[strings.ToLower(grpcutils.CustomErrorMetadataKey)]
	require.Equal(t, true, ok, "could not retrieve custom error metadata value")
	assert.DeepEqual(
		t,
		[]string{"{\"failures\":[{\"index\":0,\"message\":\"expected target epoch (99) to be the previous epoch (0) or the current epoch (1)\"}]}"},
		v,
	)
}
