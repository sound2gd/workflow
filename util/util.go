package util

import (
	"strconv"
	"strings"
)

const (
	// DatetimeHMS "2006-01-02 15:04:05"
	DatetimeHMS = "2006-01-02 15:04:05"

	// DatetimeHM "2006-01-02 15:04"
	DatetimeHM = "2006-01-02 15:04"

	// DatetimeH "2006-01-02 15"
	DatetimeH = "2006-01-02 15"
)

// IntSliceToString int slice cobine to string with ","
func IntSliceToString(intSlice []int) string {
	str := ""
	if len(intSlice) == 0 {
		return str
	}
	for index, item := range intSlice {
		if index == 0 {
			str = strconv.Itoa(item)
		} else {
			str = str + "," + strconv.Itoa(item)
		}
	}

	str = strings.Trim(str, ",")
	return str
}

// RemoveFromIntSlice 传递的是引用,会改变原始的slice数据. ???
func RemoveFromIntSlice(intSlice []int, elem int) []int {
	found := false
	if len(intSlice) == 0 {
		return intSlice
	}
	i := 0
	for index, item := range intSlice {
		if item == elem {
			i = index
			found = true
			break
		}
	}
	if found {
		intSlice = append(intSlice[:i], intSlice[i+1:]...)
	}
	return intSlice
}
