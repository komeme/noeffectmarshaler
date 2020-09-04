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
	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/static"
	"golang.org/x/tools/go/ssa"
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
	FactTypes: []analysis.Fact{new(isWrapper)},
}

func run(pass *analysis.Pass) (interface{}, error) {
	// identify json.Marshaler Interface
	marshalerType := analysisutil.TypeOf(pass, "encoding/json", "Marshaler")
	if marshalerType == nil {
		return nil, nil
	}
	marshalerIf, ok := marshalerType.Underlying().(*types.Interface)
	if !ok {
		return nil, errors.New("failed to identify json.Marshaler Interface")
	}

	// identify json.Marshal Object
	jsonMarshalObj := analysisutil.ObjectOf(pass, "encoding/json", "Marshal").(*types.Func)
	if jsonMarshalObj == nil {
		return nil, errors.New("failed to identify json.Marshal Function")
	}

	// search target struct in this analyzer
	tgtStructs := make(map[*types.TypeName]bool)
	for _, name := range pass.Pkg.Scope().Names() {
		obj, ok := pass.Pkg.Scope().Lookup(name).(*types.TypeName)
		if ok && obj != nil {
			// json.Marshaler Interfaceをpointer receiverで実装しているstruct
			if !types.Implements(obj.Type(), marshalerIf) && types.Implements(types.NewPointer(obj.Type()), marshalerIf) {
				tgtStructs[obj] = true
			}
		}
	}

	// create call graph
	s := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	graph := static.CallGraph(s.Pkg.Prog)
	callers := Callers(graph.Nodes[targetFunctions(graph)]) // json.Marshalを内部的に呼んでいく関数ら

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
				for tgtStruct := range tgtStructs {
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

// TODO もっといい探し方
func targetFunctions(graph *callgraph.Graph) *ssa.Function {
	var tgt *ssa.Function
	for function, _ := range graph.Nodes {
		if function == nil || function.Pkg == nil {
			continue
		}
		if function.Package().Pkg.Path() == "encoding/json" && function.Name() == "Marshal" {
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

type isWrapper struct{}

func (f *isWrapper) AFact() {}
