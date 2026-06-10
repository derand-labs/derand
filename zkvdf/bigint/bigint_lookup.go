package bigint

import (
	"zkvdf/biguint"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/lookup/logderivlookup"
)

type LookupTable interface {
	Insert(bs ...BigInt)
	Lookup(idx frontend.Variable) BigInt
}

type lookupTable struct {
	signTable logderivlookup.Table
	magTable  biguint.LookupTable
}

func NewLookupTable(api frontend.API) LookupTable {
	return &lookupTable{
		signTable: logderivlookup.New(api),
		magTable:  biguint.NewLookupTable(api),
	}
}

func (table *lookupTable) Insert(bs ...BigInt) {
	for _, b := range bs {
		table.signTable.Insert(b.Sign)
		table.magTable.Insert(b.Mag)
	}
}

func (table lookupTable) Lookup(idx frontend.Variable) BigInt {
	return BigInt{Sign: table.signTable.Lookup(idx)[0], Mag: table.magTable.Lookup(idx)}
}
