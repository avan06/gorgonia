package main

import "text/template"

// level 2 aggregation (tensor.StdEng) templates

const prepVVRaw = `var reuse *Dense
	var safe, toReuse {{if eq .TypeClassCheck "Number"}}, incr{{else}}, same{{end}} bool
	if reuse, safe, toReuse, {{if eq .TypeClassCheck "Number"}}incr, _,{{else}}_, same,{{end}} err = prepBinaryTensor(a, b, {{.TypeClassCheck | lower}}Types, opts...); err != nil {
		return
	}

	if reuse != nil && !reuse.IsNativelyAccessible() {
		err = errors.Errorf(inaccessibleData, reuse)
		return
	}

	typ := a.Dtype().Type
	var dataA, dataB, dataReuse *storage.Header
	var ait, bit, iit Iterator
	var useIter bool
	if dataA, dataB, dataReuse, ait, bit, iit, useIter, err = prepDataVV(a, b, reuse); err != nil {
		err = errors.Wrapf(err, "StdEng.{{.Name}}")
		return
	}
`

const prepMixedRaw = `var reuse *Dense
	var safe, toReuse {{if eq .TypeClassCheck "Number"}}, incr{{else}}, same{{end}} bool
	if reuse, safe, toReuse, {{if eq .TypeClassCheck "Number"}}incr, _,{{else}}_, same,{{end}} err = prepUnaryTensor(t, {{.TypeClassCheck | lower}}Types, opts...); err != nil {
		return
	}

	a := t
	typ := t.Dtype().Type
	var ait, bit,  iit Iterator
	var dataA, dataB, dataReuse *storage.Header
	var useIter bool

	if leftTensor {
		if dataA, dataB, dataReuse, ait, iit, useIter, err = prepDataVS(t, s, reuse); err != nil {
			err = errors.Wrapf(err, opFail, "StdEng.{{.Name}}")
			return
		}
	} else {
		if dataA, dataB, dataReuse, bit, iit, useIter, err = prepDataSV(s, t, reuse); err != nil {
			err = errors.Wrapf(err, opFail, "StdEng.{{.Name}}")
			return
		}	
	}

`

const prepUnaryRaw = `var reuse *Dense
	var safe, toReuse, incr bool
	if reuse, safe, toReuse, incr, _, err = prepUnaryTensor(a, {{.TypeClassCheck | lower}}Types, opts...); err != nil {
		return
	}

	typ := a.Dtype().Type
	var ait, rit Iterator
	var dataA, dataReuse *storage.Header
	var useIter bool

	if dataA, dataReuse, ait, rit, useIter, err = prepDataUnary(a, reuse); err != nil{
		err = errors.Wrapf(err, opFail, "StdEng.{{.Name}}")
		return
	}
	`

const agg2BodyRaw = `if useIter {
		switch {
		case incr:
			err = e.E.{{.Name}}IterIncr(typ, dataA, dataB, dataReuse, ait, bit, iit)
			retVal = reuse
		{{if .VV -}}
		case toReuse:
			storage.CopyIter(typ,dataReuse, dataA, iit, ait)
			ait.Reset()
			iit.Reset()
			err = e.E.{{.Name}}Iter(typ, dataReuse, dataB, iit, bit)
			retVal = reuse
		{{else -}}
		case toReuse && leftTensor:
			storage.CopyIter(typ, dataReuse, dataA, iit, ait)
			ait.Reset()
			iit.Reset()
			err = e.E.{{.Name}}Iter(typ, dataReuse, dataB, iit, bit)
			retVal = reuse
		case toReuse && !leftTensor:
			storage.CopyIter(typ, dataReuse, dataB, iit, bit)
			ait.Reset()
			iit.Reset()
			err = e.E.{{.Name}}Iter(typ, dataA, dataReuse, ait, iit)
			retVal = reuse
		{{end -}}
		case !safe:
			err = e.E.{{.Name}}Iter(typ, dataA, dataB, ait, bit)
			retVal = a
		default:
			ret := a.Clone().(headerer)
		{{if .VV -}}
			err = e.E.{{.Name}}Iter(typ, ret.hdr(), dataB, ait, bit)
		{{else -}}
			if leftTensor {
				err = e.E.{{.Name}}Iter(typ, ret.hdr(), dataB, ait, bit)
			}else {
				err = e.E.{{.Name}}Iter(typ, dataA, ret.hdr(), ait, bit)
			}
		{{end -}}
			retVal = ret.(Tensor)
		}
		return
	}
	switch {
	case incr:
		err = e.E.{{.Name}}Incr(typ, dataA, dataB, dataReuse)
		retVal = reuse
	{{if .VV -}}
	case toReuse:
		storage.Copy(typ,dataReuse, dataA)
		err = e.E.{{.Name}}(typ, dataReuse, dataB)
		retVal = reuse
	{{else -}}
	case toReuse && leftTensor:
		storage.Copy(typ, dataReuse, dataA)
		err = e.E.{{.Name}}(typ, dataReuse, dataB)
		retVal = reuse
	case toReuse && !leftTensor:
		storage.Copy(typ, dataReuse, dataB)
		err = e.E.{{.Name}}(typ, dataA, dataReuse)
		retVal = reuse
	{{end -}}
	case !safe:
		err = e.E.{{.Name}}(typ, dataA, dataB)
		retVal = a
	default:
		ret := a.Clone().(headerer)
		{{if .VV -}}
		err = e.E.{{.Name}}(typ, ret.hdr(), dataB)
		{{else -}}
		if leftTensor {
			err = e.E.{{.Name}}(typ, ret.hdr(), dataB)
		}else {
			err = e.E.{{.Name}}(typ, dataA, ret.hdr())	
		}
		{{end -}}
		retVal = ret.(Tensor)
	}
	return
`

const agg2CmpBodyRaw = `
	if !same && !toReuse{
		reuse = NewDense(Bool, a.Shape().Clone(), WithEngine(e))
		dataReuse = reuse.array.hdr()
		iit = IteratorFromDense(reuse)
	}

	if useIter {
		switch {
			case !toReuse && same:
				err = e.E.{{.Name}}SameIter(typ, dataA, dataB, ait, bit)
				retVal = a
			case toReuse && same:
				storage.CopyIter(typ,dataReuse,dataA, iit, ait)
				ait.Reset()
				iit.Reset()
				err = e.E.{{.Name}}SameIter(typ, dataReuse, dataB, iit, bit)
				retVal = reuse
			case toReuse && !same:
				err = e.E.{{.Name}}Iter(typ, dataA, dataB, dataReuse, ait, bit, iit)
				retVal = reuse
			case !safe:
				err = e.E.{{.Name}}SameIter(typ, dataA, dataB, ait, bit)
				retVal = a
			default:
				err = e.E.{{.Name}}Iter(typ, dataA, dataB, dataReuse, ait, bit, iit)
				retVal = reuse
		}
		return
	}
	switch {
			case !toReuse && same:
				err = e.E.{{.Name}}Same(typ, dataA, dataB)
				retVal = a
			case toReuse && same:
				storage.Copy(typ,dataReuse,dataA)
				err = e.E.{{.Name}}Same(typ, dataReuse, dataB)
				retVal = reuse
			case toReuse && !same:
				err = e.E.{{.Name}}(typ, dataA, dataB, dataReuse)
				retVal = reuse
			case !safe:
				err = e.E.{{.Name}}Same(typ, dataA, dataB)
				retVal = a
			default:
				err = e.E.{{.Name}}(typ, dataA, dataB, dataReuse)
				retVal = reuse
	}
	return
`

const agg2UnaryBodyRaw = `
	if useIter{
		switch {
		case incr:
			cloned := a.Clone().(Tensor)
			var dataCloned *storage.Header
			if dataCloned, _, _, _,_, err = prepDataUnary(cloned, nil); err != nil{
				err = errors.Wrap(err, "Unable to clone a")
				return
			}
			if err = e.E.{{.Name}}Iter(typ, dataCloned, ait); err != nil {
				err = errors.Wrap(err, "Unable to perform {{.Name}}")
				return
			}
			err = e.E.AddIter(typ, dataReuse, dataCloned, rit, ait)
			retVal = reuse
		case toReuse:
			storage.CopyIter(typ, dataReuse, dataA, rit, ait)
			rit.Reset()
			err = e.E.{{.Name}}Iter(typ, dataReuse, rit)
			retVal = reuse
		case !safe:
			err = e.E.{{.Name}}Iter(typ, dataA, ait)
			retVal = a
		default: // safe by default
			cloned := a.Clone().(Tensor)
			var dataCloned *storage.Header
			if dataCloned, _, _, _, _, err = prepDataUnary(cloned, nil); err != nil{
				err = errors.Wrapf(err,"Unable to clone")
				return
			}
			err = e.E.{{.Name}}Iter(typ, dataCloned, ait)
			retVal = cloned
		}
		return
	}else {
		switch {
		case incr:
			cloned := a.Clone().(Tensor)
			var dataCloned *storage.Header
			if dataCloned, _, _, _, _, err = prepDataUnary(cloned, nil); err != nil{
				err = errors.Wrap(err, "Unable to clone")
				return
			}
			if err = e.E.{{.Name}}(typ, dataCloned); err != nil {
				err = errors.Wrap(err, "Unable to perform {{.Name}}")
				return
			}
			err = e.E.Add(typ, dataReuse, dataCloned)
			retVal = reuse
		case toReuse:
			storage.Copy(typ, dataReuse, dataA)
			err = e.E.{{.Name}}(typ, dataReuse)
			retVal = reuse
		case !safe:
			err = e.E.{{.Name}}(typ, dataA)
			retVal = a
		default: // safe by default
			cloned := a.Clone().(Tensor)
			var dataCloned *storage.Header
			if dataCloned, _, _,_, _,err = prepDataUnary(cloned, nil); err != nil{
				err = errors.Wrapf(err, "Unable to clone")
				return
			}
			err = e.E.{{.Name}}(typ, dataCloned)
			retVal = cloned
		}
		return
	}
`

var (
	prepVV        *template.Template
	prepMixed     *template.Template
	prepUnary     *template.Template
	agg2Body      *template.Template
	agg2CmpBody   *template.Template
	agg2UnaryBody *template.Template
)

func init() {
	prepVV = template.Must(template.New("prepVV").Funcs(funcs).Parse(prepVVRaw))
	prepMixed = template.Must(template.New("prepMixed").Funcs(funcs).Parse(prepMixedRaw))
	prepUnary = template.Must(template.New("prepUnary").Funcs(funcs).Parse(prepUnaryRaw))
	agg2Body = template.Must(template.New("agg2body").Funcs(funcs).Parse(agg2BodyRaw))
	agg2CmpBody = template.Must(template.New("agg2CmpBody").Funcs(funcs).Parse(agg2CmpBodyRaw))
	agg2UnaryBody = template.Must(template.New("agg2UnaryBody").Funcs(funcs).Parse(agg2UnaryBodyRaw))
}