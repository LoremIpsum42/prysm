package testing

import (
	"fmt"
	"os"
	"strconv"

	statepb "github.com/prysmaticlabs/prysm/proto/prysm/v2/state"
	"github.com/prysmaticlabs/prysm/shared/testutil"
)

const fileBase = "0-11-0/mainnet/beaconstate"
const fileBaseENV = "BEACONSTATES_PATH"

// BeaconFuzzState returns a beacon state by ID using the beacon-fuzz corpora.
func BeaconFuzzState(ID uint64) (*statepb.BeaconState, error) {
	base := fileBase
	// Using an environment variable allows a host image to specify the path when only the binary
	// executable was uploaded (without the runfiles). i.e. fuzzit's platform.
	if p, ok := os.LookupEnv(fileBaseENV); ok {
		base = p
	}
	ok, err := testutil.BazelDirectoryNonEmpty(base)
	if err != nil {
		panic(err)
	}
	if !ok {
		panic(fmt.Sprintf("Beacon states directory (%s) does not exist or has no files.", base))
	}
	b, err := testutil.BazelFileBytes(base, strconv.Itoa(int(ID)))
	if err != nil {
		return nil, err
	}
	st := &statepb.BeaconState{}
	if err := st.UnmarshalSSZ(b); err != nil {
		return nil, err
	}
	return st, nil
}
