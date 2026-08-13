package cli

import (
	"flag"
	"fmt"
	"io"

	"github.com/entroforge/go-system-builder/internal/scenario"
)

// runScenario exposes the module-level fact-driven scenario package without
// adding a new Runtime stage or evidence kind. The scenario engine owns the
// package contract; this layer only maps CLI arguments and reports.
func runScenario(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "scenario requires <generate|validate>")
		return 2
	}
	switch args[0] {
	case "generate":
		return runScenarioGenerate(args[1:], stdout, stderr)
	case "validate":
		return runScenarioValidate(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "scenario: unknown subcommand %q; expected generate or validate\n", args[0])
		return 2
	}
}

func runScenarioGenerate(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("scenario generate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	bindUsage(flags, "scenario generate")
	module := flags.String("module", "", "module name")
	root := flags.String("root", ".", "repository root")
	all := flags.Bool("all", false, "not supported for generate; use --module")
	requireSpecs := flags.Bool("require-specs", false, "only supported for validate")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if len(flags.Args()) != 0 {
		fmt.Fprintln(stderr, "scenario generate: unexpected positional arguments")
		return 2
	}
	if *all {
		fmt.Fprintln(stderr, "scenario generate: --all is not supported; provide exactly one --module")
		return 2
	}
	if *requireSpecs {
		fmt.Fprintln(stderr, "scenario generate: --require-specs is only supported for validate")
		return 2
	}
	if *module == "" {
		fmt.Fprintln(stderr, "scenario generate requires --module")
		return 2
	}
	report, err := scenario.GenerateModule(*root, *module)
	if err != nil {
		fmt.Fprintln(stderr, formatFailure("scenario generate", err))
		return 1
	}
	return encodeJSON(stdout, report)
}

func runScenarioValidate(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("scenario validate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	bindUsage(flags, "scenario validate")
	module := flags.String("module", "", "module name")
	all := flags.Bool("all", false, "validate all module packages")
	root := flags.String("root", ".", "repository root")
	requireSpecs := flags.Bool("require-specs", false, "require module-scoped CASE and PATH references in Playwright specs")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if len(flags.Args()) != 0 {
		fmt.Fprintln(stderr, "scenario validate: unexpected positional arguments")
		return 2
	}
	if (*module == "") != *all {
		fmt.Fprintln(stderr, "scenario validate requires exactly one of --module or --all; they are mutually exclusive")
		return 2
	}
	options := scenario.ValidateOptions{RequireSpecs: *requireSpecs}
	if *all {
		reports, err := scenario.ValidateAll(*root, options)
		if err != nil {
			fmt.Fprintln(stderr, formatFailure("scenario validate", err))
			return 1
		}
		return encodeJSON(stdout, reports)
	}
	report, err := scenario.ValidateModule(*root, *module, options)
	if err != nil {
		fmt.Fprintln(stderr, formatFailure("scenario validate", err))
		return 1
	}
	return encodeJSON(stdout, report)
}
