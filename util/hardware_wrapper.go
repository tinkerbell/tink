package util

import (
	"encoding/json"

	"github.com/tinkerbell/tink/protos/hardware"
)

// HardwareWrapper is a wrapper for the Hardware type generated from the protos
// this is to allow for custom marshalling/unmarshalling of the hardware type
type HardwareWrapper struct {
	*hardware.Hardware
}

func (h HardwareWrapper) MarshalJSON() ([]byte, error) {
	type hwWrapper HardwareWrapper                     // intermediary type to avoid infinite recursion
	hw := make(map[string]interface{})                 // map to hold metadata as a map (as opposed to string)
	hwByte, err := json.Marshal(hwWrapper{h.Hardware}) // marshal hardware h into []byte
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(hwByte, &hw) // fill out the hw map
	if err != nil {
		return nil, err
	}
	if h.Metadata != "" {
		metadata := make(map[string]interface{})            // map to unmarshal metadata string into
		err = json.Unmarshal([]byte(h.Metadata), &metadata) // fill metadata map
		if err != nil {
			return nil, err
		}
		hw["metadata"] = metadata // set hw metadata field to metadata map
	}
	b, err := json.Marshal(hw) // marshal hw map into []byte
	if err != nil {
		return nil, err
	}
	return b, err
}

func (h *HardwareWrapper) UnmarshalJSON(b []byte) error {
	type hwWrapper HardwareWrapper // intermediary type to avoid infinite recursion
	hw := make(map[string]interface{})
	err := json.Unmarshal(b, &hw)
	if err != nil {
		return err
	}
	if _, ok := hw["metadata"]; ok {
		metadata, err := json.Marshal(hw["metadata"])
		if err != nil {
			return err
		}
		hw["metadata"] = string(metadata)
	}
	tmp, err := json.Marshal(hw)
	if err != nil {
		return err
	}
	var w hwWrapper
	err = json.Unmarshal(tmp, &w)
	if err != nil {
		return err
	}
	*h = HardwareWrapper(w)
	return nil
}
