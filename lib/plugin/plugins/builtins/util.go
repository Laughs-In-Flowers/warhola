package builtins

import (
	"image"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

func clamp(value, min, max float64) float64 {
	if value > max {
		return max
	}
	if value < min {
		return min
	}
	return value
}

func clampUint8(x float64) uint8 {
	v := int64(x + 0.5)
	if v > 255 {
		return 255
	}
	if v > 0 {
		return uint8(v)
	}
	return 0
}

func stringToRect(s string) image.Rectangle {
	req := strings.Split(s, ",")
	lreq := len(req)
	if lreq != 4 {
		return image.ZR
	}
	var n []int
	for _, v := range req {
		iv, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			spew.Dump(err)
			return image.ZR
		}
		n = append(n, int(iv))
	}
	return image.Rect(n[0], n[1], n[2], n[3])
}
