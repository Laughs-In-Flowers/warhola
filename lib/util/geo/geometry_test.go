package geo

import (
	"reflect"
	"testing"
)

type check struct {
	k string
	v interface{}
}

type expect struct {
	t string
	r string
	c []check
}

func newExpect(t, r string, c ...check) expect {
	return expect{t, r, c}
}

var expected = []expect{
	newExpect("area", "500@", check{"Area", 500}),
	newExpect("aspect", "1.33:9.99", check{"AspectX", 1.33}),
	newExpect("aspect", "1.33:9.99", check{"AspectY", 9.99}),
	newExpect("gravity", "North", check{"Gravity", North}),
	newExpect("height", "height:900", check{"Y", 900}),
	newExpect("height", "y:900", check{"Y", 900}),
	newExpect("offset", "{+600/600}", check{"OffsetX", 600}),
	newExpect("offset", "{600/-600}", check{"OffsetY", -600}),
	newExpect("offset", "offset:+600,600", check{"OffsetX", 600}),
	newExpect("offset", "offset:600,-600", check{"OffsetY", -600}),
	newExpect("rect", "100,200,500,1000",
		check{"OffsetX", 100},
		check{"OffsetY", 200},
		check{"X", 500},
		check{"Y", 1000}),
	newExpect("rect", "rect:100,200,500,1000",
		check{"OffsetX", 100},
		check{"OffsetY", 200},
		check{"X", 500},
		check{"Y", 1000}),
	newExpect("scale", "300%", check{"ScaleX", 300.0}, check{"ScaleY", 300.0}),
	newExpect("sxs", "200%300", check{"ScaleX", 200.0}),
	newExpect("sxs", "200%300", check{"ScaleY", 300.0}),
	newExpect("wxh", "500x900", check{"X", 500}),
	newExpect("wxh", "500x900", check{"Y", 900}),
	newExpect("width", "width:500", check{"X", 500}),
	newExpect("width", "x:500", check{"X", 500}),
	newExpect("width+height", "x:500 y:900", check{"X", 500}, check{"Y", 900}),
	newExpect("width+height", "width:500 height:900", check{"X", 500}, check{"Y", 900}),
	newExpect("offset+height+width", "{222/223} width:500 height:900",
		check{"OffsetX", 222},
		check{"OffsetY", 223},
		check{"X", 500},
		check{"Y", 900},
	),
	newExpect("offset+height+width", "{222/223} x:500 y:900",
		check{"OffsetX", 222},
		check{"OffsetY", 223},
		check{"X", 500},
		check{"Y", 900},
	),
}

func TestGeometry(t *testing.T) {
	for _, v := range expected {
		ng := New(v.r)
		if len(ng.Error) != 0 {
			t.Errorf("error creating geometry for '%s': %v", v.t, ng.Error)
		}
		f := fill(ng)
		for _, vv := range v.c {
			a, b := f[vv.k], vv.v
			if a != b {
				t.Errorf("unequal: '%s - %s' expected %v == %v", v.t, v.r, a, b)
			}
		}
	}
}

func strctVal(g *Geometry) reflect.Value {
	v := reflect.ValueOf(g)

	// if pointer get the underlying element≤
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		panic("not struct")
	}

	return v
}

func structFields(g *Geometry) []reflect.StructField {
	t := strctVal(g).Type()

	var f []reflect.StructField

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		f = append(f, field)
	}

	return f
}

func fill(g *Geometry) map[string]interface{} {
	out := make(map[string]interface{})

	value := strctVal(g)

	fields := structFields(g)

	for _, field := range fields {
		name := field.Name
		val := value.FieldByName(name)
		finalVal := val.Interface()
		out[name] = finalVal
	}
	return out
}
