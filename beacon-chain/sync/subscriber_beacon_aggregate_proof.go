package sync

import (
	"context"
	"errors"
	"fmt"

	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"google.golang.org/protobuf/proto"
)

// beaconAggregateProofSubscriber forwards the incoming validated aggregated attestation and proof to the
// attestation pool for processing.
func (s *Service) beaconAggregateProofSubscriber(_ context.Context, msg proto.Message) error {
	a, ok := msg.(*ethpb.SignedAggregateAttestationAndProof)
	if !ok {
		return fmt.Errorf("message was not type *eth.SignedAggregateAttestationAndProof, type=%T", msg)
	}

	if a.Message.Aggregate == nil || a.Message.Aggregate.Data == nil {
		return errors.New("nil aggregate")
	}

	// An unaggregated attestation can make it here. It’s valid, the aggregator it just itself, although it means poor performance for the subnet.
	if !helpers.IsAggregated(a.Message.Aggregate) {
		return s.cfg.AttPool.SaveUnaggregatedAttestation(a.Message.Aggregate)
	}

	return s.cfg.AttPool.SaveAggregatedAttestation(a.Message.Aggregate)
}
