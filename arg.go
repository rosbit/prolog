package pl

import (
	"github.com/mndrix/golog/term"
	"github.com/mndrix/golog/read"
	"reflect"
	"fmt"
	"math/big"
)

type PlArg interface {
	ToTerm() term.Term
}

var (
	_ PlBool  = PlBool(false)
	_ PlInt   = PlInt(0)
	_ PlFloat = PlFloat(0.0)
	_ PlString= PlString("")
	_ PlVar   = PlVar("")
	_ PlStrTerm = PlStrTerm("[]")
	_ *PlList = &PlList{}
	_ *PlRecord = &PlRecord{}
)

func makePlTerm(v interface{}) (term.Term, error) {
	if v == nil {
		return term.NewAtom("[]"), nil
	}

	switch vv := v.(type) {
	case int, int8, int16, int32, int64,
	     uint,uint8,uint16,uint32,uint64:
		return makeInt(v), nil
	case string:
		return PlString(vv).ToTerm(), nil
	case bool:
		return PlBool(vv).ToTerm(), nil
	case float64:
		return PlFloat(vv).ToTerm(), nil
	case float32:
		return PlFloat(float64(vv)).ToTerm(), nil
	case PlVar:
		return vv.ToTerm(), nil
	case PlStrTerm:
		return vv.ToTerm(), nil
	case Record:
		r, err := newRecord(vv)
		if err != nil {
			return term.NewAtom("[]"), err
		}
		return r.ToTerm(), nil
	default:
	}

	vv := reflect.ValueOf(v)
	switch vv.Kind() {
	case reflect.Slice:
		t := vv.Type()
		if t.Elem().Kind() == reflect.Uint8 {
			return PlString(string(v.([]byte))).ToTerm(), nil
		}
		fallthrough
	case reflect.Array:
		if plL, err := newPlList(v); err != nil {
			return term.NewAtom("[]"), err
		} else {
			return plL.ToTerm(), nil
		}
	case reflect.Ptr:
		switch vv.Elem().Kind() {
		case reflect.Array, reflect.Struct:
			if plL, err := newPlList(v); err != nil {
				return term.NewAtom("[]"), err
			} else {
				return plL.ToTerm(), nil
			}
		}
		return makePlTerm(vv.Elem().Interface())
	/*
	case reflect.Map:
	case reflect.Struct:
	case reflect.Func:
	*/
	default:
		return nil, fmt.Errorf("unsupported type %v", vv.Kind())
	}
}

func makeInt(i interface{}) term.Term {
	switch i.(type) {
	case int,int8,int16,int32,int64:
		return PlInt(reflect.ValueOf(i).Int()).ToTerm()
	case uint8,uint16,uint32:
		return PlInt(int64(reflect.ValueOf(i).Uint())).ToTerm()
	case uint,uint64:
		return PlFloat(float64(reflect.ValueOf(i).Uint())).ToTerm()
	default:
		return term.NewInt64(0)
	}
}

func fromPlTerm(plVal term.Term) (goVal interface{}, err error) {
	plType := plVal.Type()
	switch plType {
	case term.AtomType:
		s := plVal.String()
		l := len(s)
		if l <= 2 {
			goVal = s
			return
		}
		if s[0] == '\'' && s[l-1] == '\'' {
			goVal = s[1:l-1]
			return
		}
		goVal = s
		return
	case term.IntegerType:
		i, _ := plVal.(*term.Integer)
		goVal = ((*big.Int)(i)).Int64()
		return
	case term.FloatType:
		goVal = plVal.(*term.Float).Value()
		return
	case term.CompoundType:
		c, ok := plVal.(*term.Compound)
		if ok {
			if c.Func == "." {
				args := c.Args
				res, e := fromPlCons(args)
				if e != nil {
					err = e
					return
				}
				goVal = res
				return
			}
			goVal = c.String()
			return
		}
		goVal = plVal.String()
		return
	// case term.VariableType:
	// case term.ErrorType:
	default:
		err = fmt.Errorf("unsupported type %d", plType)
		return
	}
}

func fromPlCons(cons []term.Term) (res []interface{}, err error) {
	head, e := fromPlTerm(cons[0])
	if e != nil {
		err = e
		return
	}
	res = append(res, head)
	for cons[1] != nil {
		if tail, ok := cons[1].(*term.Compound); ok {
			if tail.Func == "." {
				cons = tail.Args
				head, e = fromPlTerm(cons[0])
				if e != nil {
					err = e
					break
				}
				res = append(res, head)
			}
		} else {
			break
		}
	}
	return
}

// var
type PlVar string
func (v PlVar) ToTerm() term.Term {
	return term.NewVar(string(v))
}

// bool
type PlBool bool
func (b PlBool) ToTerm() term.Term {
	if b {
		return term.NewAtom("true")
	}
	return term.NewAtom("false")
}

// int
type PlInt int64
func (i PlInt) ToTerm() term.Term {
	return term.NewInt64(int64(i))
}

// float
type PlFloat float64
func (f PlFloat) ToTerm() term.Term {
	return term.NewFloat64(float64(f))
}

// string
type PlString string
func (s PlString) ToTerm() term.Term {
	return term.NewAtom(string(s))
}

// string -> term
type PlStrTerm string
func (s PlStrTerm) ToTerm() term.Term {
	str := string(s)
	if len(str) == 0 {
		return term.NewAtom("")
	}

	strT := fmt.Sprintf("%s.", str)
	if t, err := read.Term(strT); err == nil {
		return t
	}

	strT = fmt.Sprintf("%s.", term.QuoteFunctor(str))
	t, err := read.Term(strT)
	if err != nil {
		panic(err)
	}
	return t
}

// list
type PlList struct {
	pa term.Term
}
func newPlList(a interface{}) (plL *PlList, err error) {
	if a == nil {
		plL = &PlList{term.NewAtom("[]")}
		return
	}
	v := reflect.ValueOf(a)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		l := v.Len()
		pa := make([]term.Term, l)
		for i:=0; i<l; i++ {
			e := v.Index(i).Interface()
			ev, err := makePlTerm(e)
			if err != nil {
				pa[i] = term.NewAtom("[]")
			} else {
				pa[i] = ev
			}
		}
		plL = &PlList{term.NewTermList(pa)}
		return
	/*
	case reflect.Map:
	case reflect.Struct:
	*/
	case reflect.Ptr:
		return newPlList(v.Elem().Interface())
	default:
		err = fmt.Errorf("slice or list expected")
		return
	}
}
func (a *PlList) ToTerm() term.Term {
	return term.Term(a.pa)
}

// record
type PlRecord struct {
	r Record
}
func newRecord(rec Record) (*PlRecord, error) {
	if len(rec.TableName()) == 0 {
		return nil, fmt.Errorf("table name expected for record")
	}
	return &PlRecord{r: rec}, nil
}
func (r *PlRecord) ToTerm() term.Term {
	fields := r.r.FieldValues()
	ts := make([]term.Term, len(fields))
	for i, f := range fields {
		ts[i], _ = makePlTerm(f)
	}
	return term.NewCallable(r.r.TableName(), ts...)
}
