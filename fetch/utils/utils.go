package utils

import (
	"path/filepath"
	"strconv"
	"strings"
)

// AppendIfMissing ...
func AppendIfMissing(slice []string, i string) []string {
	for _, ele := range slice {
		if ele == i {
			return slice
		}
	}
	return append(slice, i)
}

// Worker ...
func Worker(jobs <-chan string, results chan<- []StreamFlow, idxs *[]uint64) {
	for f := range jobs {
		var product string
		if strings.Contains(f, "retrospective") {
			product = "Retrospective"
		} else {
			product = strings.Split(filepath.Base(f), ".")[2]
		}
		result, err := GetNetCDFData(f, &product, idxs)
		// fmt.Println("Processing: ", f)
		if err != nil {
			// fmt.Println("Read Error in Go Channel for:", f)
			var errorStruct []StreamFlow
			errorStruct = append(errorStruct, StreamFlow{string(f), 0.0, "Error", uint64(0)})
			results <- errorStruct
		} else {
			results <- result
		}

	}
}

// StringsToUint64s ...
func StringsToUint64s(strs []string) []uint64 {
	uints := make([]uint64, 0)
	for _, s := range strs {
		i64, err := strconv.ParseInt(s, 10, 64)
		CheckError(err)
		uints = append(uints, uint64(i64))
	}
	return uints
}
