package myanalyzer

import (
	"errors"
	"github.com/gostaticanalysis/analysisutil"
	"github.com/gostaticanalysis/ident"
	"go/types"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/inspect"
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

	// for debug
	//for _, implementsStruct := range tgtStructs {
	//	fmt.Println(implementsStruct)
	//}

	//m := pass.ResultOf[ident.Analyzer].(ident.Map)
	//for id := range m{
	//	for _, tgt := range tgtStructs {
	//		if types.Identical(id.Type(), tgt.Type()){
	//			//pass.Reportf(id.Pos(), "%s", id.Type().String())
	//			//fmt.Printf("%s: %s", obj.Pos(), obj.Type().String())
	//		}
	//	}
	//}

	//inspect_ := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	//
	//nodeFilter := []ast.Node{
	//	new(ast.CallExpr),
	//}
	//
	//inspect_.Preorder(nodeFilter, func(n ast.Node) {
	//	switch n := n.(type) {
	//	case *ast.CallExpr:
	//		ast.Print(token.NewFileSet(), n)
	//	}
	//})

	//s := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	//for _, f := range s.SrcFuncs {
	//	fmt.Println(f)
	//	for _, b := range f.Blocks {
	//		fmt.Printf("\tBlock %d\n", b.Index)
	//		for _, instr := range b.Instrs {
	//			fmt.Printf("\t\t%[1]T\t%[1]v(%[1]p)\n", instr)
	//			for _, v := range instr.Operands(nil) {
	//				if v != nil { fmt.Printf("\t\t\t%[1]T\t%[1]v(%[1]p)\n", *v) }
	//			}
	//		}
	//	}
	//}

	return nil, nil
}
