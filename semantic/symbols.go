package semantic

import (
	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/types"
)

type Symbol interface {
	Name() string
	Type() types.Type
}

type BasicSymbol struct {
	name string
	typ  types.Type
}

func (bs *BasicSymbol) Name() string {
	return bs.name
}

func (bs *BasicSymbol) Type() types.Type {
	return bs.typ
}

type DeclareSymbol struct {
	name string
	typ  types.Type
	mut  bool
}

func (ds *DeclareSymbol) Name() string {
	return ds.name
}

func (ds *DeclareSymbol) Type() types.Type {
	return ds.typ
}

func (ds *DeclareSymbol) Mutable() bool {
	return ds.mut
}

type FunctionSymbol struct {
	name        string
	typ         types.Type
	params      []*ast.Parameter
	returnTypes []types.Type
}

func (fs *FunctionSymbol) Name() string {
	return fs.name
}

func (fs *FunctionSymbol) Type() types.Type {
	return fs.typ
}

func (fs *FunctionSymbol) Params() []*ast.Parameter {
	return fs.params
}

func (fs *FunctionSymbol) ReturnTypes() []types.Type {
	return fs.returnTypes
}
