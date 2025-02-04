package v1

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/beacon-chain/state/stateutil"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	statepb "github.com/prysmaticlabs/prysm/proto/prysm/v2/state"
	"github.com/prysmaticlabs/prysm/shared/hashutil"
)

func (f *FieldTrie) validateIndices(idxs []uint64) error {
	for _, idx := range idxs {
		if idx >= f.length {
			return errors.Errorf("invalid index for field %s: %d >= length %d", f.field.String(), idx, f.length)
		}
	}
	return nil
}

func validateElements(field fieldIndex, elements interface{}, length uint64) error {
	val := reflect.ValueOf(elements)
	if val.Len() > int(length) {
		return errors.Errorf("elements length is larger than expected for field %s: %d > %d", field.String(), val.Len(), length)
	}
	return nil
}

// this converts the corresponding field and the provided elements to the appropriate roots.
func fieldConverters(field fieldIndex, indices []uint64, elements interface{}, convertAll bool) ([][32]byte, error) {
	switch field {
	case blockRoots, stateRoots, randaoMixes:
		val, ok := elements.([][]byte)
		if !ok {
			return nil, errors.Errorf("Wanted type of %v but got %v",
				reflect.TypeOf([][]byte{}).Name(), reflect.TypeOf(elements).Name())
		}
		return stateutil.HandleByteArrays(val, indices, convertAll)
	case eth1DataVotes:
		val, ok := elements.([]*ethpb.Eth1Data)
		if !ok {
			return nil, errors.Errorf("Wanted type of %v but got %v",
				reflect.TypeOf([]*ethpb.Eth1Data{}).Name(), reflect.TypeOf(elements).Name())
		}
		return HandleEth1DataSlice(val, indices, convertAll)
	case validators:
		val, ok := elements.([]*ethpb.Validator)
		if !ok {
			return nil, errors.Errorf("Wanted type of %v but got %v",
				reflect.TypeOf([]*ethpb.Validator{}).Name(), reflect.TypeOf(elements).Name())
		}
		return stateutil.HandleValidatorSlice(val, indices, convertAll)
	case previousEpochAttestations, currentEpochAttestations:
		val, ok := elements.([]*statepb.PendingAttestation)
		if !ok {
			return nil, errors.Errorf("Wanted type of %v but got %v",
				reflect.TypeOf([]*statepb.PendingAttestation{}).Name(), reflect.TypeOf(elements).Name())
		}
		return handlePendingAttestation(val, indices, convertAll)
	default:
		return [][32]byte{}, errors.Errorf("got unsupported type of %v", reflect.TypeOf(elements).Name())
	}
}

// HandleEth1DataSlice processes a list of eth1data and indices into the appropriate roots.
func HandleEth1DataSlice(val []*ethpb.Eth1Data, indices []uint64, convertAll bool) ([][32]byte, error) {
	length := len(indices)
	if convertAll {
		length = len(val)
	}
	roots := make([][32]byte, 0, length)
	hasher := hashutil.CustomSHA256Hasher()
	rootCreator := func(input *ethpb.Eth1Data) error {
		newRoot, err := eth1Root(hasher, input)
		if err != nil {
			return err
		}
		roots = append(roots, newRoot)
		return nil
	}
	if convertAll {
		for i := range val {
			err := rootCreator(val[i])
			if err != nil {
				return nil, err
			}
		}
		return roots, nil
	}
	if len(val) > 0 {
		for _, idx := range indices {
			if idx > uint64(len(val))-1 {
				return nil, fmt.Errorf("index %d greater than number of items in eth1 data slice %d", idx, len(val))
			}
			err := rootCreator(val[idx])
			if err != nil {
				return nil, err
			}
		}
	}
	return roots, nil
}

func handlePendingAttestation(val []*statepb.PendingAttestation, indices []uint64, convertAll bool) ([][32]byte, error) {
	length := len(indices)
	if convertAll {
		length = len(val)
	}
	roots := make([][32]byte, 0, length)
	hasher := hashutil.CustomSHA256Hasher()
	rootCreator := func(input *statepb.PendingAttestation) error {
		newRoot, err := stateutil.PendingAttRootWithHasher(hasher, input)
		if err != nil {
			return err
		}
		roots = append(roots, newRoot)
		return nil
	}
	if convertAll {
		for i := range val {
			err := rootCreator(val[i])
			if err != nil {
				return nil, err
			}
		}
		return roots, nil
	}
	if len(val) > 0 {
		for _, idx := range indices {
			if idx > uint64(len(val))-1 {
				return nil, fmt.Errorf("index %d greater than number of pending attestations %d", idx, len(val))
			}
			err := rootCreator(val[idx])
			if err != nil {
				return nil, err
			}
		}
	}
	return roots, nil
}
