package pkg

import (
	"encoding/json"

	"github.com/tinkerbell/tink/protos/hardware"
)

// HardwareWrapper is a wrapper for the Hardware type generated from the protos
// to allow for custom marshalling/unmarshalling of the Hardware type
//
// Why we need this:
// Since the metadata field is a json formatted string, when it's being marshaled
// along with the rest of the Hardware object to json format, the double quotes
// inside the metadata string would be automatically escaped, making it hard to read/consume.
//
// The point of the custom marshal/unmarshal is to manually convert the metadata field
// between a string (for creating a hardware object) and a map (for printing) as a workaround
// for this problem.
type HardwareWrapper struct {
	*hardware.Hardware
}

// MarshalJSON marshals the HardwareWrapper object, with the metadata field represented as a map
// as the final result
//
// Since the Hardware object's metadata field is of type string, to convert the field into a map
// we need to solve somewhat of a water jug problem.
//
// 1. Create an empty map A
// 2. Marshal Hardware object to then be unmarshaled into map A
// 3. Create another empty map B to unmarshal metadata string into
// 4. Set map B as map A's metadata field
// 5. Marshal map A
func (h HardwareWrapper) MarshalJSON() ([]byte, error) {
	tmp := make(map[string]interface{})     // map (A) to hold metadata as a map (as opposed to string)
	hwByte, err := json.Marshal(h.Hardware) // marshal hardware object
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(hwByte, &tmp) // fill out the tmp map
	if err != nil {
		return nil, err
	}
	if h.Metadata != "" {
		metadata := make(map[string]interface{})            // map (B) to unmarshal metadata string into
		err = json.Unmarshal([]byte(h.Metadata), &metadata) // fill metadata map
		if err != nil {
			return nil, err
		}
		tmp["metadata"] = metadata // set hw metadata field to the metadata map
	}
	tmpByte, err := json.Marshal(tmp) // marshal hw map into []byte
	if err != nil {
		return nil, err
	}
	return tmpByte, err
}

// UnmarshalJSON unmarshals the HardwareWrapper object, with the metadata field represented as a string
// as the final result
//
// Since the incoming []byte b represents the metadata as a map, we can't directly unmarshal it into
// a Hardware object, cue water jug problem.
//
// 1. Create empty map
// 2. Unmarshal []byte b into map
// 3. Marshal metadata field
// 4. Set metadata string as map's metadata field
// 5. Marshal map to then be unmarshaled into Hardware object
func (h *HardwareWrapper) UnmarshalJSON(b []byte) error {
	type hwWrapper HardwareWrapper      // intermediary type to avoid infinite recursion
	tmp := make(map[string]interface{}) // map to hold metadata as a string (as well as all the hardware data)
	err := json.Unmarshal(b, &tmp)      // unmarshal []byte b into map
	if err != nil {
		return err
	}
	if _, ok := tmp["metadata"]; ok { // check if there's metadata
		metadata, err := json.Marshal(tmp["metadata"]) // marshal metadata field
		if err != nil {
			return err
		}
		tmp["metadata"] = string(metadata) // set tmp metadata field to the metadata string
	}
	tmpByte, err := json.Marshal(tmp) // marshal map tmp
	if err != nil {
		return err
	}
	var hw hwWrapper
	err = json.Unmarshal(tmpByte, &hw) // unmarshal tmpByte into hardware object
	if err != nil {
		return err
	}
	*h = HardwareWrapper(hw)
	return nil
}
