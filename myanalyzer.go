package myanalyzer

import (
	"errors"
	"github.com/gostaticanalysis/analysisutil"
	"github.com/gostaticanalysis/ident"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "myanalyzer is ..."

// Analyzer is ...
var Analyzer = &analysis.Analyzer{
	Name: "myanalyzer",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		ident.Analyzer,
		inspect.Analyzer,
		buildssa.Analyzer,
	},
}

func run(pass *analysis.Pass) (interface{}, error) {
	// identify json.Marshaler Interface
	marshalerType, ok := analysisutil.TypeOf(pass, "encoding/json", "Marshaler").Underlying().(*types.Interface)
	if !ok {
		return nil, errors.New("failed to identify json.Marshaler Interface")
	}

	// identify json.Marshal Interface
	jsonMarshalMethod := analysisutil.ObjectOf(pass, "encoding/json", "Marshal")
	if jsonMarshalMethod == nil {
		return nil, errors.New("failed to identify json.Marshal Function")
	}

	// search target struct in this analyzer
	tgtStructs := make([]*types.TypeName, 0)
	for _, name := range pass.Pkg.Scope().Names() {
		obj, ok := pass.Pkg.Scope().Lookup(name).(*types.TypeName)
		if ok && obj != nil {
			// json.Marshaler Interfaceをpointer receiverで実装しているstruct
			if !types.Implements(obj.Type(), marshalerType) && types.Implements(types.NewPointer(obj.Type()), marshalerType) {
				tgtStructs = append(tgtStructs, obj)
			}
		}
	}

	// json.Marshalに上記structが値渡しされている箇所を検出 // TODO 内部的にjson.Marshalを呼んでいるものを検出

	inspect_ := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect_.Preorder([]ast.Node{new(ast.CallExpr)}, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.CallExpr:
			for _, arg := range n.Args {
				tv, ok := pass.TypesInfo.Types[arg]
				if !ok {
					return
				}

				isTarget := false
				for _, tgtStruct := range tgtStructs {
					if types.Identical(tv.Type, tgtStruct.Type()) {
						isTarget = true
					}
				}

				if !isTarget {
					continue
				}

				caller, ok := n.Fun.(*ast.SelectorExpr)
				if !ok {
					return
				}

				if pass.TypesInfo.ObjectOf(caller.Sel).Pkg() != jsonMarshalMethod.Pkg() {
					continue
				}

				if caller.Sel.Name != "Marshal" { // TODO ハードコーディングやめたい
					continue
				}

				pass.Reportf(n.Pos(), "NG")
			}
		}
	})

	return nil, nil
}
