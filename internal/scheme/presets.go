package scheme

import (
	"fmt"
	"regexp"
	"strings"
)

// FunctionDef is a named formula preset (macro) expanded before govaluate parsing.
type FunctionDef struct {
	Params []string `yaml:"params"`
	Expr   string   `yaml:"expr"`
	Desc   string   `yaml:"desc,omitempty"`
}

// reservedNames are govaluate/math helpers — never treated as presets.
var reservedNames = map[string]bool{
	"exp": true, "log": true, "ln": true, "log10": true, "exp10": true,
	"sqrt": true, "abs": true, "pow": true, "min": true, "max": true, "pi": true,
}

// BuiltinFunctions are default presets (Marinov / report); YAML functions override by name.
func BuiltinFunctions() map[string]FunctionDef {
	return map[string]FunctionDef{
		"arrhenius": {
			Params: []string{"V", "E"},
			Expr:   "V * exp(-E / (R * T))",
			Desc:   "Аррениус: ν·exp(−E/(R·T))",
		},
		"per": {
			Params: []string{"E"},
			Expr:   "exp(-E / (R * T))",
			Desc:   "PER / E-R: exp(−E/(R·T))",
		},
		"plh": {
			Params: []string{"E"},
			Expr:   "exp(-E / (R * T))",
			Desc:   "PLH / L–H: exp(−E/(R·T))",
		},
		"adsorption_flux": {
			Params: []string{},
			Expr:   "atomFlux / (F_density + S_density)",
			Desc:   "Скорость адсорбции r1/r3: Φ/(F+S)",
		},
		"er_with_r3": {
			Params: []string{"E"},
			Expr:   "per(E) * r3",
			Desc:   "E-R: PER·r3 (вложенный пресет per)",
		},
		"bkl_sites_rate": {
			Params: []string{"sites", "rate"},
			Expr:   "sites * rate",
			Desc:   "λ BKL: число сайтов × скорость",
		},
		"bkl_atoms_rate": {
			Params: []string{"atoms", "rate"},
			Expr:   "atoms * rate",
			Desc:   "λ BKL: атомы на центре × скорость",
		},
	}
}

// FunctionRegistry merges builtins with scheme-specific functions (YAML wins on name clash).
func (s *Scheme) FunctionRegistry() map[string]FunctionDef {
	reg := BuiltinFunctions()
	if s == nil {
		return reg
	}
	for name, def := range s.Functions {
		reg[name] = def
	}
	return reg
}

// ExpandFormula expands preset calls like arrhenius(Vdes, Edes) into govaluate expressions.
// Expressions without preset calls are returned unchanged (manual input).
func ExpandFormula(expr string, registry map[string]FunctionDef) (string, error) {
	if strings.TrimSpace(expr) == "" {
		return "", fmt.Errorf("empty expression")
	}
	const maxIter = 64
	out := expr
	for iter := 0; iter < maxIter; iter++ {
		name, start, end, args, ok := findNextPresetCall(out, 0, registry)
		if !ok {
			return out, nil
		}
		def, ok := registry[name]
		if !ok {
			return "", fmt.Errorf("unknown preset %q", name)
		}
		if len(args) != len(def.Params) {
			return "", fmt.Errorf("preset %q: expected %d argument(s), got %d", name, len(def.Params), len(args))
		}
		body, err := substituteParams(def.Expr, def.Params, args, registry)
		if err != nil {
			return "", fmt.Errorf("preset %q: %w", name, err)
		}
		replacement := "(" + body + ")"
		out = out[:start] + replacement + out[end:]
	}
	return "", fmt.Errorf("preset expansion: too many nested expansions")
}

func substituteParams(body string, params, args []string, registry map[string]FunctionDef) (string, error) {
	out := body
	for i, p := range params {
		arg := strings.TrimSpace(args[i])
		if arg == "" {
			return "", fmt.Errorf("empty argument for parameter %q", p)
		}
		// Expand nested presets inside arguments first
		expandedArg, err := ExpandFormula(arg, registry)
		if err != nil {
			return "", err
		}
		arg = expandedArg
		repl := arg
		if needsParens(repl) {
			repl = "(" + repl + ")"
		}
		re := regexp.MustCompile(`\b` + regexp.QuoteMeta(p) + `\b`)
		if !re.MatchString(out) {
			return "", fmt.Errorf("parameter %q not found in preset body", p)
		}
		out = re.ReplaceAllString(out, repl)
	}
	// Expand any presets left in the body (e.g. per(E) inside er_with_r3)
	return ExpandFormula(out, registry)
}

func needsParens(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for _, c := range s {
		switch c {
		case '+', '-', '*', '/', '^', ' ', '\t':
			return true
		}
	}
	return false
}

func findNextPresetCall(expr string, from int, registry map[string]FunctionDef) (name string, start, end int, args []string, ok bool) {
	for i := from; i < len(expr); i++ {
		if !isIdentStart(expr[i]) {
			continue
		}
		j := i + 1
		for j < len(expr) && isIdentChar(expr[j]) {
			j++
		}
		candidate := expr[i:j]
		if reservedNames[candidate] {
			continue
		}
		if _, isPreset := registry[candidate]; !isPreset {
			continue
		}
		k := j
		for k < len(expr) && (expr[k] == ' ' || expr[k] == '\t') {
			k++
		}
		if k >= len(expr) || expr[k] != '(' {
			continue
		}
		closeIdx, found := matchingCloseParen(expr, k)
		if !found {
			continue
		}
		inner := expr[k+1 : closeIdx]
		parsedArgs, err := splitTopLevelArgs(inner)
		if err != nil {
			continue
		}
		return candidate, i, closeIdx + 1, parsedArgs, true
	}
	return "", 0, 0, nil, false
}

func matchingCloseParen(s string, openIdx int) (int, bool) {
	if openIdx >= len(s) || s[openIdx] != '(' {
		return 0, false
	}
	depth := 0
	for i := openIdx; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i, true
			}
		}
	}
	return 0, false
}

func splitTopLevelArgs(inner string) ([]string, error) {
	inner = strings.TrimSpace(inner)
	if inner == "" {
		return []string{}, nil
	}
	var args []string
	var b strings.Builder
	depth := 0
	for i := 0; i < len(inner); i++ {
		c := inner[i]
		switch c {
		case '(':
			depth++
			b.WriteByte(c)
		case ')':
			depth--
			b.WriteByte(c)
		case ',':
			if depth == 0 {
				args = append(args, b.String())
				b.Reset()
				continue
			}
			b.WriteByte(c)
		default:
			b.WriteByte(c)
		}
	}
	args = append(args, b.String())
	return args, nil
}

func isIdentStart(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func isIdentChar(c byte) bool {
	return isIdentStart(c) || (c >= '0' && c <= '9')
}
