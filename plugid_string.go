// Code generated by "stringer -type=plugID"; DO NOT EDIT.

package main

import "fmt"

const _plugID_name = "plugAllplugOneplugTwo"

var _plugID_index = [...]uint8{0, 7, 14, 21}

func (i plugID) String() string {
	if i < 0 || i >= plugID(len(_plugID_index)-1) {
		return fmt.Sprintf("plugID(%d)", i)
	}
	return _plugID_name[_plugID_index[i]:_plugID_index[i+1]]
}