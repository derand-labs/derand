package classgroup

import (
	"zkvdf/bigint"

	"github.com/consensys/gnark/frontend"
)

type LookupTable interface {
	Insert(bs ...Form)
	Lookup(idx frontend.Variable) Form
}

type lookupTable struct {
	aTable bigint.LookupTable
	bTable bigint.LookupTable
	cTable bigint.LookupTable
}

func NewLookupTable(api frontend.API) LookupTable {
	return &lookupTable{
		aTable: bigint.NewLookupTable(api),
		bTable: bigint.NewLookupTable(api),
		cTable: bigint.NewLookupTable(api),
	}
}

func (table *lookupTable) Insert(fs ...Form) {
	for _, f := range fs {
		table.aTable.Insert(f.A)
		table.bTable.Insert(f.B)
		table.cTable.Insert(f.C)
	}
}

func (table lookupTable) Lookup(idx frontend.Variable) Form {
	return Form{
		A: table.aTable.Lookup(idx),
		B: table.bTable.Lookup(idx),
		C: table.cTable.Lookup(idx),
	}
}
