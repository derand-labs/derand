package biguint

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/lookup/logderivlookup"
)

type LookupTable interface {
	Insert(bs ...BigUint)
	Lookup(idx frontend.Variable) BigUint
}

type lookupTable struct {
	api        frontend.API
	arraysize  int
	limbTables []logderivlookup.Table
}

func NewLookupTable(api frontend.API) LookupTable {
	return &lookupTable{api: api, arraysize: 0, limbTables: make([]logderivlookup.Table, 0)}
}

func (table *lookupTable) Insert(bs ...BigUint) {
	for _, b := range bs {
		if len(b.Limbs) > len(table.limbTables) {
			for range len(b.Limbs) - len(table.limbTables) {
				table.limbTables = append(table.limbTables, logderivlookup.New(table.api))
				for range table.arraysize {
					table.limbTables[len(table.limbTables)-1].Insert(0)
				}
			}
		}

		for i, limb := range b.Limbs {
			table.limbTables[i].Insert(limb)
		}

		table.arraysize += 1
	}
}

func (table lookupTable) Lookup(idx frontend.Variable) BigUint {
	limbs := []frontend.Variable{}
	for i := range table.limbTables {
		limbs = append(limbs, table.limbTables[i].Lookup(idx)[0])
	}

	return BigUint{Limbs: limbs}
}
