package poet

import "bytes"

// FuncSpec represents information needed to write a function
type FuncSpec struct {
	Name             string
	Comment          string
	Parameters       []IdentifierParameter
	ResultParameters []IdentifierParameter
	Statements       []statement
}

var _ CodeBlock = (*FuncSpec)(nil)

// NewFuncSpec returns a FuncSpec with the given name
func NewFuncSpec(name string) *FuncSpec {
	return &FuncSpec{
		Name:             name,
		Parameters:       []IdentifierParameter{},
		ResultParameters: []IdentifierParameter{},
		Statements:       []statement{},
	}
}

// String returns a string representation of the function
func (f *FuncSpec) String() string {
	writer := newCodeWriter()

	if f.Comment != "" {
		writer.WriteCode("// " + f.Comment + "\n")
	}

	writer.WriteStatement(f.createSignature())

	for _, st := range f.Statements {
		writer.WriteStatement(st)
	}

	writer.WriteStatement(statement{
		BeforeIndent: -1,
		Format:       "}",
	})

	return writer.String()
}

// createSignature generates the function's signature as a statement, starting from "func" and ending with
// the opening curly brace.
func (f *FuncSpec) createSignature() statement {
	formatStr := bytes.Buffer{}
	signature, args := f.Signature()

	formatStr.WriteString("func ")
	formatStr.WriteString(signature)
	formatStr.WriteString(" {")

	return statement{
		AfterIndent: 1,
		Format:      formatStr.String(),
		Arguments:   args,
	}
}

// Signature returns a format string and slice of arguments for the function's signature, not
// including the starting "func" or opening curly brace
func (f *FuncSpec) Signature() (_ string, arguments []interface{}) {
	formatStr := bytes.Buffer{}

	formatStr.WriteString(f.Name)
	formatStr.WriteString("(")

	for i, param := range f.Parameters {
		formatStr.WriteString("$L $T")
		if param.Variadic {
			formatStr.WriteString("...")
		}

		arguments = append(arguments, param.Name, param.Type)

		if i != len(f.Parameters)-1 {
			formatStr.WriteString(", ")
		}
	}

	formatStr.WriteString(")")

	if len(f.ResultParameters) == 1 && f.ResultParameters[0].Name == "" {
		formatStr.WriteString(" $T")
		arguments = append(arguments, f.ResultParameters[0].Type)
	} else if len(f.ResultParameters) >= 1 {

		formatStr.WriteString(" (")
		for i, resultParameter := range f.ResultParameters {
			if resultParameter.Name != "" {
				formatStr.WriteString("$L ")
				arguments = append(arguments, resultParameter.Name)
			}

			formatStr.WriteString("$T")
			arguments = append(arguments, resultParameter.Type)

			if i != len(f.ResultParameters)-1 {
				formatStr.WriteString(", ")
			}
		}
		formatStr.WriteString(")")
	}

	return formatStr.String(), arguments
}

// GetImports returns a slice of imports that this function needs, including
// parameters, result parameters, and statements within the function
func (f *FuncSpec) GetImports() []Import {
	packages := []Import{}

	for _, st := range f.Statements {
		for _, arg := range st.Arguments {
			if asTypeRef, ok := arg.(TypeReference); ok {
				packages = append(packages, asTypeRef.GetImports()...)
			}
		}
	}

	for _, param := range f.Parameters {
		packages = append(packages, param.Type.GetImports()...)
	}

	for _, param := range f.ResultParameters {
		packages = append(packages, param.Type.GetImports()...)
	}

	return packages
}

// Statement is a convenient method to append a statement to the function
func (f *FuncSpec) Statement(format string, args ...interface{}) *FuncSpec {
	f.Statements = append(f.Statements, statement{
		Format:    format,
		Arguments: args,
	})

	return f
}

// BlockStart is a convenient method to append a statement that marks the start of a
// block of code.
func (f *FuncSpec) BlockStart(format string, args ...interface{}) *FuncSpec {
	f.Statements = append(f.Statements, statement{
		Format:      format + " {",
		Arguments:   args,
		AfterIndent: 1,
	})

	return f
}

// BlockEnd is a convenient method to append a statement that marks the end of a
// block of code.
func (f *FuncSpec) BlockEnd() *FuncSpec {
	f.Statements = append(f.Statements, statement{
		Format:       "}",
		BeforeIndent: -1,
	})

	return f
}

// Parameter is a convenient method to append a parameter to the function
func (f *FuncSpec) Parameter(name string, spec TypeReference) *FuncSpec {
	f.Parameters = append(f.Parameters, IdentifierParameter{
		Identifier: Identifier{
			Name: name,
			Type: spec,
		},
	})

	return f
}

// VariadicParameter is a convenient method to append a parameter to the function
func (f *FuncSpec) VariadicParameter(name string, spec TypeReference) *FuncSpec {
	f.Parameters = append(f.Parameters, IdentifierParameter{
		Identifier: Identifier{
			Name: name,
			Type: spec,
		},
		Variadic: true,
	})

	return f
}

// ResultParameter is a convenient method to append a result parameter to the function
func (f *FuncSpec) ResultParameter(name string, spec TypeReference) *FuncSpec {
	f.ResultParameters = append(f.ResultParameters, IdentifierParameter{
		Identifier: Identifier{
			Name: name,
			Type: spec,
		},
	})

	return f
}

// FunctionComment adds a comment to the function
func (f *FuncSpec) FunctionComment(comment string) *FuncSpec {
	f.Comment = comment

	return f
}
