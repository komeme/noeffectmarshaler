package noeffectmarshaler

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
	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/static"
	"golang.org/x/tools/go/ssa"
)

const doc = "noeffectmarshaler is ..."

// Analyzer is ...
var Analyzer = &analysis.Analyzer{
	Name: "noeffectmarshaler",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		ident.Analyzer,
		inspect.Analyzer,
		buildssa.Analyzer,
	},
	FactTypes: []analysis.Fact{new(dummy)},
}

type dummy struct{}

func (f *dummy) AFact() {}

func run(pass *analysis.Pass) (interface{}, error) {
	// identify json.Marshaler Interface
	marshalerType := analysisutil.TypeOf(pass, "encoding/json", "Marshaler")
	if marshalerType == nil {
		return nil, nil
	}
	marshalerIface, ok := marshalerType.Underlying().(*types.Interface)
	if !ok {
		return nil, errors.New("failed to identify json.Marshaler Interface")
	}

	// identify json.Marshal Object
	jsonMarshalObj := analysisutil.ObjectOf(pass, "encoding/json", "Marshal").(*types.Func)
	if jsonMarshalObj == nil {
		return nil, errors.New("failed to identify json.Marshal Function")
	}

	// search target struct in this analyzer
	implementors := pointerReceivingImplementors(pass, marshalerIface)

	// create call graph
	s := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	graph := static.CallGraph(s.Pkg.Prog)
	callers := Callers(graph.Nodes[targetFunctions(graph, "encoding/json", "Marshal")]) // json.Marshalを内部的に呼んでいく関数群

	// json.Marshalに上記structが値渡しされている箇所を検出
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
				for tgtStruct := range implementors {
					if types.Identical(tv.Type, tgtStruct.Type()) {
						isTarget = true
					}
				}
				if !isTarget {
					continue
				}

				switch fn := n.Fun.(type) {
				case *ast.SelectorExpr:
					funObj := pass.TypesInfo.ObjectOf(fn.Sel)
					for caller, _ := range callers {
						if funObj == caller.Func.Object() {
							pass.Reportf(n.Pos(), "NG")
							break
						}
					}
				case *ast.Ident:
					funObj := pass.TypesInfo.ObjectOf(fn)
					for caller, _ := range callers {
						if funObj == caller.Func.Object() {
							pass.Reportf(n.Pos(), "NG")
							break
						}
					}
				}
			}
		}
	})

	return nil, nil
}

func pointerReceivingImplementors(pass *analysis.Pass, iface *types.Interface) map[*types.TypeName]bool {
	result := make(map[*types.TypeName]bool)
	for _, name := range pass.Pkg.Scope().Names() {
		obj, ok := pass.Pkg.Scope().Lookup(name).(*types.TypeName)
		if ok && obj != nil {
			if !types.Implements(obj.Type(), iface) && types.Implements(types.NewPointer(obj.Type()), iface) {
				result[obj] = true
			}
		}
	}
	return result
}

// TODO もっといい探し方
func targetFunctions(graph *callgraph.Graph, pkgPath string, name string) *ssa.Function {
	var tgt *ssa.Function
	for function, _ := range graph.Nodes {
		if function == nil || function.Pkg == nil {
			continue
		}
		if function.Package().Pkg.Path() == pkgPath && function.Name() == name {
			tgt = function
			break
		}
	}

	if tgt == nil {
		return nil
	}

	return tgt
}

func Callers(target *callgraph.Node) map[*callgraph.Node]bool {
	callers := make(map[*callgraph.Node]bool)
	callers[target] = true

	for _, edge := range target.In {
		if _, ok := callers[edge.Caller]; ok {
			continue
		}
		for caller := range Callers(edge.Caller) {
			callers[caller] = true
		}
	}

	return callers
}
