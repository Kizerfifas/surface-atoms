package scheme

import (
	"fmt"
	"math"

	"github.com/Knetic/govaluate"
)

// MarinovR is the gas constant as used in Marinov/Kim–Boudart style formulas (8.31·T in denominator).
const MarinovR = 8.31

// exprFunctions are available in rate, probability, and lambda_expr (govaluate).
// Operators built into govaluate: + - * / % ^ (power), comparisons, && ||.
var exprFunctions = map[string]govaluate.ExpressionFunction{
	"exp":   unaryFloat(math.Exp, "exp"),
	"log":   unaryFloat(math.Log, "log"),
	"ln":    unaryFloat(math.Log, "ln"),
	"log10": unaryFloat(math.Log10, "log10"),
	"exp10": unaryFloat(func(x float64) float64 { return math.Pow(10, x) }, "exp10"),
	"sqrt":  unaryFloat(math.Sqrt, "sqrt"),
	"abs":   unaryFloat(math.Abs, "abs"),
	"pow":   binaryFloat(math.Pow, "pow"),
	"min":   binaryFloat(math.Min, "min"),
	"max":   binaryFloat(math.Max, "max"),
	"pi": nullaryFloat(math.Pi, "pi"),
}

// FunctionNames returns supported function names for documentation/UI.
func FunctionNames() []string {
	return []string{
		"exp", "log", "ln", "log10", "exp10", "sqrt", "abs", "pow", "min", "max", "pi",
	}
}

func unaryFloat(fn func(float64) float64, name string) govaluate.ExpressionFunction {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("%s expects 1 argument", name)
		}
		f, ok := toFloat(args[0])
		if !ok {
			return nil, fmt.Errorf("%s: argument not a number", name)
		}
		return fn(f), nil
	}
}

func binaryFloat(fn func(float64, float64) float64, name string) govaluate.ExpressionFunction {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("%s expects 2 arguments", name)
		}
		a, ok := toFloat(args[0])
		if !ok {
			return nil, fmt.Errorf("%s: first argument not a number", name)
		}
		b, ok := toFloat(args[1])
		if !ok {
			return nil, fmt.Errorf("%s: second argument not a number", name)
		}
		return fn(a, b), nil
	}
}

func nullaryFloat(val float64, name string) govaluate.ExpressionFunction {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("%s expects 0 arguments", name)
		}
		return val, nil
	}
}
