package lang

import (
	"github.com/pt-main/tycl/shared"
)

// MergeInto copies every entry of src over dst.
//
// Objects present on both sides are merged recursively, so an override file
// only needs to declare the keys it actually changes. Scalars, arrays and
// typed nulls are replaced wholesale.
func MergeInto(dst, src *shared.Config) {
	if dst == nil || src == nil || dst == src {
		return
	}

	for key, v := range src.IntV {
		dst.IntV[key] = v
	}
	for key, v := range src.FloatV {
		dst.FloatV[key] = v
	}
	for key, v := range src.BoolV {
		dst.BoolV[key] = v
	}
	for key, v := range src.StringV {
		dst.StringV[key] = v
	}
	for key, v := range src.NullV {
		dst.NullV[key] = v
	}
	for key, v := range src.IntArrV {
		dst.IntArrV[key] = v
	}
	for key, v := range src.FloatArrV {
		dst.FloatArrV[key] = v
	}
	for key, v := range src.BoolArrV {
		dst.BoolArrV[key] = v
	}
	for key, v := range src.StringArrV {
		dst.StringArrV[key] = v
	}
	for key, v := range src.InnerArrV {
		dst.InnerArrV[key] = v
	}
	for key, v := range src.InnerV {
		existing, ok := dst.InnerV[key]
		if !ok {
			dst.InnerV[key] = v
			continue
		}
		MergeInto(existing, v)
	}
}
