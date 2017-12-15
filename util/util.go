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

// IntSliceToString int slice combine to string with ","
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

// RemoveFromIntSlice remove elem from slice
func RemoveFromIntSlice(intSlice []int, elem int) []int {
	slen := len(intSlice)
	if slen == 0 {
		return intSlice
	}
	temp := make([]int, 0, slen)
	for _, item := range intSlice {
		if item != elem {
			temp = append(temp, item)
		}
	}
	return temp
}

// StringInSlice 字符是否在slice中存在
func StringInSlice(str string, strSlice []string) (found bool) {
	for _, s := range strSlice {
		if str == s {
			return true
		}
	}
	return false
}
