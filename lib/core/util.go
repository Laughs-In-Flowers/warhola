package core

import "unicode"

type AlphaS []string

func (a AlphaS) Len() int      { return len(a) }
func (a AlphaS) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a AlphaS) Less(i, j int) bool {
	iRunes := []rune(a[i])
	jRunes := []rune(a[j])

	max := len(iRunes)
	if max > len(jRunes) {
		max = len(jRunes)
	}

	for idx := 0; idx < max; idx++ {
		ir := iRunes[idx]
		jr := jRunes[idx]

		lir := unicode.ToLower(ir)
		ljr := unicode.ToLower(jr)

		if lir != ljr {
			return lir < ljr
		}

		// the lowercase runes are the same, so compare the original
		if ir != jr {
			return ir < jr
		}
	}
	return false
}

//func clampUint8(x float64) uint8 {
//	v := int64(x + 0.5)
//	if v > 255 {
//		return 255
//	}
//	if v > 0 {
//		return uint8(v)
//	}
//	return 0
//}

//func stringToRect(s string) image.Rectangle {
//	req := strings.Split(s, ",")
//	lreq := len(req)
//	if lreq != 4 {
//		return image.ZR
//	}
//	var n []int
//	for _, v := range req {
//		iv, err := strconv.ParseInt(v, 10, 64)
//		if err != nil {
//			return image.ZR
//		}
//		n = append(n, int(iv))
//	}
//	return image.Rect(n[0], n[1], n[2], n[3])
//}
