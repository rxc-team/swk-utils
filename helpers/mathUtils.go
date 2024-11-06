package helpers

import (
	"math"
)

const (
	// BtoMB bytes to 1megabyte
	BtoMB = 1048576.00
)

// IntToFloat int to float convert
func IntToFloat(num int64) float64 {
	return float64(num * 100.00 / 100.00)
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

// ToFixed float user define precision option
func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(float64(num)*output)) / output
}

// CONVERSIONS

// BytesToMegabyte convert bytes to megabytes
func BytesToMegabyte(bytes int64, precision int) float64 {
	return ToFixed(IntToFloat(bytes)/BtoMB, precision)
}

// GroupBigSlices group big slices of array to chunks
func GroupBigSlices(b int, a ...interface{}) [][]interface{} {
	var c [][]interface{}
	for i := 0; i < len(a); i += b {
		if i+b > len(a) {
			c = append(c, a[i:])
		} else {
			c = append(c, a[i:i+b])
		}
	}
	return c
}

// mapObjToInterface convert map of slice to slices of interface for bulk
func mapObjToInterface(objMap ...map[string]interface{}) []interface{} {
	var x []interface{}
	for _, item := range objMap {
		x = append(x, item)
	}
	return x
}
