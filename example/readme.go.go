package main

import "github.com/d4l3k/go-pry/pry"

func main() {
	a := 1
	pry.Apply(map[string]interface{}{ "main": main, "a": a, "pry": pry.Package{Name: "pry", Functions: map[string]interface{}{"Package": pry.Package{}, "Pry": pry.Pry, "Apply": pry.Apply, "Append": pry.Append, "Make": pry.Make, "Highlight": pry.Highlight, "HighlightWords": pry.HighlightWords, "InterpretExpr": pry.InterpretExpr, "StringToType": pry.StringToType, "ValuesToInterfaces": pry.ValuesToInterfaces, "Scope": pry.Scope{}, "InterpretString": pry.InterpretString, "ComputeBinaryOp": pry.ComputeBinaryOp, "ComputeUnaryOp": pry.ComputeUnaryOp, }}, })

}
