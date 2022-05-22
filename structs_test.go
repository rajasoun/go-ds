package structs

import (
	"reflect"
	"testing"
	"time"
)

func TestMapNonStruct(t *testing.T) {
	foo := []string{"foo"}

	defer func() {
		err := recover()
		if err == nil {
			t.Error("Passing a non struct into Map should panic")
		}
	}()

	// this should panic. We are going to recover and and test it
	_ = Map(foo)
}

func TestMap(t *testing.T) {
	var T = struct {
		A string
		B int
		C bool
	}{
		A: "a-value",
		B: 2,
		C: true,
	}

	a := Map(T)

	if typ := reflect.TypeOf(a).Kind(); typ != reflect.Map {
		t.Errorf("Map should return a map type, got: %v", typ)
	}

	// we have three fields
	if len(a) != 3 {
		t.Errorf("Map should return a map of len 3, got: %d", len(a))
	}

	inMap := func(val interface{}) bool {
		for _, v := range a {
			if reflect.DeepEqual(v, val) {
				return true
			}
		}

		return false
	}

	for _, val := range []interface{}{"a-value", 2, true} {
		if !inMap(val) {
			t.Errorf("Map should have the value %v", val)
		}
	}

}

func TestMap_Tag(t *testing.T) {
	var T = struct {
		A string `structs:"x"`
		B int    `structs:"y"`
		C bool   `structs:"z"`
	}{
		A: "a-value",
		B: 2,
		C: true,
	}

	a := Map(T)

	inMap := func(key interface{}) bool {
		for k := range a {
			if reflect.DeepEqual(k, key) {
				return true
			}
		}
		return false
	}

	for _, key := range []string{"x", "y", "z"} {
		if !inMap(key) {
			t.Errorf("Map should have the key %v", key)
		}
	}

}

func TestMap_CustomTag(t *testing.T) {
	var T = struct {
		A string `json:"x"`
		B int    `json:"y"`
		C bool   `json:"z"`
		D struct {
			E string `json:"jkl"`
		} `json:"nested"`
	}{
		A: "a-value",
		B: 2,
		C: true,
	}
	T.D.E = "e-value"

	s := New(T)
	s.TagName = "json"

	a := s.Map()

	inMap := func(key interface{}) bool {
		for k := range a {
			if reflect.DeepEqual(k, key) {
				return true
			}
		}
		return false
	}

	for _, key := range []string{"x", "y", "z"} {
		if !inMap(key) {
			t.Errorf("Map should have the key %v", key)
		}
	}

	nested, ok := a["nested"].(map[string]interface{})
	if !ok {
		t.Fatalf("Map should contain the D field that is tagged as 'nested'")
	}

	e, ok := nested["jkl"].(string)
	if !ok {
		t.Fatalf("Map should contain the D.E field that is tagged as 'jkl'")
	}

	if e != "e-value" {
		t.Errorf("D.E field should be equal to 'e-value', got: '%v'", e)
	}

}

func TestMap_OmitEmpty(t *testing.T) {
	type A struct {
		Name  string
		Value string    `structs:",omitempty"`
		Time  time.Time `structs:",omitempty"`
	}
	a := A{}

	m := Map(a)

	_, ok := m["Value"].(map[string]interface{})
	if ok {
		t.Error("Map should not contain the Value field that is tagged as omitempty")
	}

	_, ok = m["Time"].(map[string]interface{})
	if ok {
		t.Error("Map should not contain the Time field that is tagged as omitempty")
	}
}

func TestMap_OmitNested(t *testing.T) {
	type A struct {
		Name  string
		Value string
		Time  time.Time `structs:",omitnested"`
	}
	a := A{Time: time.Now()}

	type B struct {
		Desc string
		A    A
	}
	b := &B{A: a}

	m := Map(b)

	in, ok := m["A"].(map[string]interface{})
	if !ok {
		t.Error("Map nested structs is not available in the map")
	}

	// should not happen
	if _, ok := in["Time"].(map[string]interface{}); ok {
		t.Error("Map nested struct should omit recursiving parsing of Time")
	}

	if _, ok := in["Time"].(time.Time); !ok {
		t.Error("Map nested struct should stop parsing of Time at is current value")
	}
}

func TestMap_Nested(t *testing.T) {
	type A struct {
		Name string
	}
	a := &A{Name: "example"}

	type B struct {
		A *A
	}
	b := &B{A: a}

	m := Map(b)

	if typ := reflect.TypeOf(m).Kind(); typ != reflect.Map {
		t.Errorf("Map should return a map type, got: %v", typ)
	}

	in, ok := m["A"].(map[string]interface{})
	if !ok {
		t.Error("Map nested structs is not available in the map")
	}

	if name := in["Name"].(string); name != "example" {
		t.Errorf("Map nested struct's name field should give example, got: %s", name)
	}
}

func TestMap_NestedMapWithStructValues(t *testing.T) {
	type A struct {
		Name string
	}

	type B struct {
		A map[string]*A
	}

	a := &A{Name: "example"}

	b := &B{
		A: map[string]*A{
			"example_key": a,
		},
	}

	m := Map(b)

	if typ := reflect.TypeOf(m).Kind(); typ != reflect.Map {
		t.Errorf("Map should return a map type, got: %v", typ)
	}

	in, ok := m["A"].(map[string]interface{})
	if !ok {
		t.Errorf("Nested type of map should be of type map[string]interface{}, have %T", m["A"])
	}

	example := in["example_key"].(map[string]interface{})
	if name := example["Name"].(string); name != "example" {
		t.Errorf("Map nested struct's name field should give example, got: %s", name)
	}
}

func TestMap_NestedMapWithStringValues(t *testing.T) {
	type B struct {
		Foo map[string]string
	}

	type A struct {
		B *B
	}

	b := &B{
		Foo: map[string]string{
			"example_key": "example",
		},
	}

	a := &A{B: b}

	m := Map(a)

	if typ := reflect.TypeOf(m).Kind(); typ != reflect.Map {
		t.Errorf("Map should return a map type, got: %v", typ)
	}

	in, ok := m["B"].(map[string]interface{})
	if !ok {
		t.Errorf("Nested type of map should be of type map[string]interface{}, have %T", m["B"])
	}

	foo := in["Foo"].(map[string]string)
	if name := foo["example_key"]; name != "example" {
		t.Errorf("Map nested struct's name field should give example, got: %s", name)
	}
}

func TestMap_NestedMapWithSliceStructValues(t *testing.T) {
	type address struct {
		Country string `structs:"country"`
	}

	type B struct {
		Foo map[string][]address
	}

	type A struct {
		B *B
	}

	b := &B{
		Foo: map[string][]address{
			"example_key": {
				{Country: "Turkey"},
			},
		},
	}

	a := &A{B: b}
	m := Map(a)

	if typ := reflect.TypeOf(m).Kind(); typ != reflect.Map {
		t.Errorf("Map should return a map type, got: %v", typ)
	}

	in, ok := m["B"].(map[string]interface{})
	if !ok {
		t.Errorf("Nested type of map should be of type map[string]interface{}, have %T", m["B"])
	}

	foo := in["Foo"].(map[string]interface{})

	addresses := foo["example_key"].([]interface{})

	addr, ok := addresses[0].(map[string]interface{})
	if !ok {
		t.Errorf("Nested type of map should be of type map[string]interface{}, have %T", m["B"])
	}

	if _, exists := addr["country"]; !exists {
		t.Errorf("Expecting country, but found Country")
	}
}

func TestMap_NestedSliceWithIntValues(t *testing.T) {
	type person struct {
		Name  string `structs:"name"`
		Ports []int  `structs:"ports"`
	}

	p := person{
		Name:  "test",
		Ports: []int{80},
	}
	m := Map(p)

	ports, ok := m["ports"].([]int)
	if !ok {
		t.Errorf("Nested type of map should be of type []int, have %T", m["ports"])
	}

	if ports[0] != 80 {
		t.Errorf("Map nested struct's ports field should give 80, got: %v", ports)
	}
}

func TestMap_Flatnested(t *testing.T) {
	type A struct {
		Name string
	}
	a := A{Name: "example"}

	type B struct {
		A `structs:",flatten"`
		C int
	}
	b := &B{C: 123}
	b.A = a

	m := Map(b)

	_, ok := m["A"].(map[string]interface{})
	if ok {
		t.Error("Embedded A struct with tag flatten has to be flat in the map")
	}

	expectedMap := map[string]interface{}{"Name": "example", "C": 123}
	if !reflect.DeepEqual(m, expectedMap) {
		t.Errorf("The exprected map %+v does't correspond to %+v", expectedMap, m)
	}

}

func TestParseTag_Name(t *testing.T) {
	tags := []struct {
		tag string
		has bool
	}{
		{"", false},
		{"name", true},
		{"name,opt", true},
		{"name , opt, opt2", false}, // has a single whitespace
		{", opt, opt2", false},
	}

	for _, tag := range tags {
		name, _ := parseTag(tag.tag)

		if (name != "name") && tag.has {
			t.Errorf("Parse tag should return name: %#v", tag)
		}
	}
}

func TestMap_TimeField(t *testing.T) {
	type A struct {
		CreatedAt time.Time
	}

	a := &A{CreatedAt: time.Now().UTC()}
	m := Map(a)

	_, ok := m["CreatedAt"].(time.Time)
	if !ok {
		t.Error("Time field must be final")
	}
}

func TestParseTag_Opts(t *testing.T) {
	tags := []struct {
		opts string
		has  bool
	}{
		{"name", false},
		{"name,opt", true},
		{"name , opt, opt2", false}, // has a single whitespace
		{",opt, opt2", true},
		{", opt3, opt4", false},
	}

	// search for "opt"
	for _, tag := range tags {
		_, opts := parseTag(tag.opts)

		if opts.Has("opt") != tag.has {
			t.Errorf("Tag opts should have opt: %#v", tag)
		}
	}
}
