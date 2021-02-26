package slasher

import (
	"reflect"
	"testing"

	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	slashertypes "github.com/prysmaticlabs/prysm/beacon-chain/slasher/types"
	"github.com/prysmaticlabs/prysm/shared/params"
	"github.com/prysmaticlabs/prysm/shared/testutil/require"
	logTest "github.com/sirupsen/logrus/hooks/test"
)

func TestService_groupByValidatorChunkIndex(t *testing.T) {
	tests := []struct {
		name   string
		params *Parameters
		atts   []*slashertypes.CompactAttestation
		want   map[uint64][]*slashertypes.CompactAttestation
	}{
		{
			name:   "No attestations returns empty map",
			params: DefaultParams(),
			atts:   make([]*slashertypes.CompactAttestation, 0),
			want:   make(map[uint64][]*slashertypes.CompactAttestation),
		},
		{
			name: "Groups multiple attestations belonging to single validator chunk",
			params: &Parameters{
				validatorChunkSize: 2,
			},
			atts: []*slashertypes.CompactAttestation{
				{
					AttestingIndices: []uint64{0, 1},
				},
				{
					AttestingIndices: []uint64{0, 1},
				},
			},
			want: map[uint64][]*slashertypes.CompactAttestation{
				0: {
					{
						AttestingIndices: []uint64{0, 1},
					},
					{
						AttestingIndices: []uint64{0, 1},
					},
				},
			},
		},
		{
			name: "Groups single attestation belonging to multiple validator chunk",
			params: &Parameters{
				validatorChunkSize: 2,
			},
			atts: []*slashertypes.CompactAttestation{
				{
					AttestingIndices: []uint64{0, 2, 4},
				},
			},
			want: map[uint64][]*slashertypes.CompactAttestation{
				0: {
					{
						AttestingIndices: []uint64{0, 2, 4},
					},
				},
				1: {
					{
						AttestingIndices: []uint64{0, 2, 4},
					},
				},
				2: {
					{
						AttestingIndices: []uint64{0, 2, 4},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				params: tt.params,
			}
			if got := s.groupByValidatorChunkIndex(tt.atts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("groupByValidatorChunkIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_groupByChunkIndex(t *testing.T) {
	tests := []struct {
		name   string
		params *Parameters
		atts   []*slashertypes.CompactAttestation
		want   map[uint64][]*slashertypes.CompactAttestation
	}{
		{
			name:   "No attestations returns empty map",
			params: DefaultParams(),
			atts:   make([]*slashertypes.CompactAttestation, 0),
			want:   make(map[uint64][]*slashertypes.CompactAttestation),
		},
		{
			name: "Groups multiple attestations belonging to single chunk",
			params: &Parameters{
				chunkSize:     2,
				historyLength: 3,
			},
			atts: []*slashertypes.CompactAttestation{
				{
					Source: 0,
				},
				{
					Source: 1,
				},
			},
			want: map[uint64][]*slashertypes.CompactAttestation{
				0: {
					{
						Source: 0,
					},
					{
						Source: 1,
					},
				},
			},
		},
		{
			name: "Groups multiple attestations belonging to multiple chunks",
			params: &Parameters{
				chunkSize:     2,
				historyLength: 3,
			},
			atts: []*slashertypes.CompactAttestation{
				{
					Source: 0,
				},
				{
					Source: 1,
				},
				{
					Source: 2,
				},
			},
			want: map[uint64][]*slashertypes.CompactAttestation{
				0: {
					{
						Source: 0,
					},
					{
						Source: 1,
					},
				},
				1: {
					{
						Source: 2,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				params: tt.params,
			}
			if got := s.groupByChunkIndex(tt.atts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("groupByChunkIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_logSlashingEvent(t *testing.T) {
	tests := []struct {
		name     string
		slashing *slashertypes.Slashing
		want     string
		noLog    bool
	}{
		{
			name:     "Surrounding vote",
			slashing: &slashertypes.Slashing{Kind: slashertypes.SurroundingVote},
			want:     "Attester surrounding vote",
		},
		{
			name:     "Surrounded vote",
			slashing: &slashertypes.Slashing{Kind: slashertypes.SurroundedVote},
			want:     "Attester surrounded vote",
		},
		{
			name:     "Double vote",
			slashing: &slashertypes.Slashing{Kind: slashertypes.DoubleVote},
			want:     "Attester double vote",
		},
		{
			name:     "Not slashable",
			slashing: &slashertypes.Slashing{Kind: slashertypes.NotSlashable},
			noLog:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook := logTest.NewGlobal()
			logSlashingEvent(tt.slashing)
			if tt.noLog {
				require.LogsDoNotContain(t, hook, "slashing")
			} else {
				require.LogsContain(t, hook, tt.want)
			}
		})
	}
}

func Test_validateAttestationIntegrity(t *testing.T) {
	tests := []struct {
		name string
		att  *ethpb.IndexedAttestation
		want bool
	}{
		{
			name: "Nil attestation returns false",
			att:  nil,
			want: false,
		},
		{
			name: "Nil attestation data returns false",
			att:  &ethpb.IndexedAttestation{},
			want: false,
		},
		{
			name: "Nil attestation source and target returns false",
			att: &ethpb.IndexedAttestation{
				Data: &ethpb.AttestationData{},
			},
			want: false,
		},
		{
			name: "Nil attestation source and good target returns false",
			att: &ethpb.IndexedAttestation{
				Data: &ethpb.AttestationData{
					Target: &ethpb.Checkpoint{},
				},
			},
			want: false,
		},
		{
			name: "Nil attestation target and good source returns false",
			att: &ethpb.IndexedAttestation{
				Data: &ethpb.AttestationData{
					Source: &ethpb.Checkpoint{},
				},
			},
			want: false,
		},
		{
			name: "Source > target returns false",
			att: &ethpb.IndexedAttestation{
				Data: &ethpb.AttestationData{
					Source: &ethpb.Checkpoint{
						Epoch: 1,
					},
					Target: &ethpb.Checkpoint{
						Epoch: 0,
					},
				},
			},
			want: false,
		},
		{
			name: "Source == target returns false",
			att: &ethpb.IndexedAttestation{
				Data: &ethpb.AttestationData{
					Source: &ethpb.Checkpoint{
						Epoch: 1,
					},
					Target: &ethpb.Checkpoint{
						Epoch: 1,
					},
				},
			},
			want: false,
		},
		{
			name: "Source < target returns true",
			att: &ethpb.IndexedAttestation{
				Data: &ethpb.AttestationData{
					Source: &ethpb.Checkpoint{
						Epoch: 1,
					},
					Target: &ethpb.Checkpoint{
						Epoch: 2,
					},
				},
			},
			want: true,
		},
		{
			name: "Source 0 target 0 returns true (genesis epoch attestations)",
			att: &ethpb.IndexedAttestation{
				Data: &ethpb.AttestationData{
					Source: &ethpb.Checkpoint{
						Epoch: 0,
					},
					Target: &ethpb.Checkpoint{
						Epoch: 0,
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateAttestationIntegrity(tt.att); got != tt.want {
				t.Errorf("validateAttestationIntegrity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isDoubleProposal(t *testing.T) {
	type args struct {
		incomingSigningRoot [32]byte
		existingSigningRoot [32]byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Existing signing root empty returns false",
			args: args{
				incomingSigningRoot: [32]byte{1},
				existingSigningRoot: params.BeaconConfig().ZeroHash,
			},
			want: false,
		},
		{
			name: "Existing signing root non-empty and equal to incoming returns false",
			args: args{
				incomingSigningRoot: [32]byte{1},
				existingSigningRoot: [32]byte{1},
			},
			want: false,
		},
		{
			name: "Existing signing root non-empty and not-equal to incoming returns true",
			args: args{
				incomingSigningRoot: [32]byte{1},
				existingSigningRoot: [32]byte{2},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDoubleProposal(tt.args.incomingSigningRoot, tt.args.existingSigningRoot); got != tt.want {
				t.Errorf("isDoubleProposal() = %v, want %v", got, tt.want)
			}
		})
	}
}