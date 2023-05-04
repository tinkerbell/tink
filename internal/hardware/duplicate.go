package hardware

import (
	"fmt"
	"strings"

	"github.com/tinkerbell/tink/api/v1alpha2"
)

type duplicates map[string]*hardwareList

func (d *duplicates) AppendTo(k string, hw ...v1alpha2.Hardware) {
	if _, ok := (*d)[k]; !ok {
		(*d)[k] = &hardwareList{}
	}
	(*d)[k].Append(hw...)
}

func (d duplicates) String() string {
	var buf []string
	for mac, dupes := range d {
		buf = append(buf, fmt.Sprintf("{%v: %v}", mac, dupes.String()))
	}
	return strings.Join(buf, "; ")
}

type hardwareList []v1alpha2.Hardware

func (d *hardwareList) Append(hw ...v1alpha2.Hardware) {
	*d = append(*d, hw...)
}

func (d hardwareList) String() string {
	var names []string
	for _, hw := range d {
		names = append(names, fmt.Sprintf("[Name: %v; Namespace: %v]", hw.Name, hw.Namespace))
	}
	return strings.Join(names, " ")
}
