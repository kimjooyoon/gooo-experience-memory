package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/gooo-experience-memory/internal/experience"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return 2
	}
	switch args[0] {
	case "compile":
		return compile(args[1:], stdout, stderr)
	case "generate":
		return generate(args[1:], stdout, stderr)
	case "evaluate":
		return evaluate(args[1:], stdout, stderr)
	case "version":
		fmt.Fprintln(stdout, "gooo-experience-memory/v1")
		return 0
	default:
		usage(stderr)
		return 2
	}
}

func compile(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("compile", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sourcePath := flags.String("source", "examples/experience-memory/main.gooo", "Gooo source")
	contractPath := flags.String("contract", "contracts/experience-memory-denominator-v1.json", "fixed denominator")
	outputPath := flags.String("output", "", "absolute caller-owned semantic IR output")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		return 2
	}
	if *outputPath == "" || !filepath.IsAbs(*outputPath) {
		fmt.Fprintln(stderr, "compile requires an absolute --output path")
		return 2
	}
	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	contract, contractDigest, err := experience.LoadDenominator(*contractPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	ir, err := experience.CompileSource(*sourcePath, source, contract, contractDigest)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := experience.WriteJSON(*outputPath, ir); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return writeStdout(stdout, ir)
}

func generate(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("generate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	irPath := flags.String("ir", "semantic-ir.json", "semantic IR")
	outputPath := flags.String("output", "", "absolute caller-owned generated Go output")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		return 2
	}
	if *outputPath == "" || !filepath.IsAbs(*outputPath) {
		fmt.Fprintln(stderr, "generate requires an absolute --output path")
		return 2
	}
	var ir experience.SemanticIR
	if err := experience.LoadJSON(*irPath, &ir); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := experience.GenerateGo(ir, *outputPath); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	result := map[string]string{"generated_go": *outputPath, "generated_go_digest": ""}
	if data, err := os.ReadFile(*outputPath); err == nil {
		result["generated_go_digest"] = experience.DigestBytes(data)
	}
	return writeStdout(stdout, result)
}

func evaluate(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("evaluate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sourcePath := flags.String("source", "examples/experience-memory/main.gooo", "Gooo source")
	contractPath := flags.String("contract", "contracts/experience-memory-denominator-v1.json", "fixed denominator")
	irPath := flags.String("ir", "semantic-ir.json", "semantic IR")
	generatedPath := flags.String("generated", "semantic.gooo.go", "generated Go")
	fixturePath := flags.String("fixture", "fixtures/fixed-fixture.json", "fixed fixture")
	memoryPath := flags.String("memory", "fixtures/memory.ndjson", "append-only memory")
	receiptPath := flags.String("receipt", "fixtures/outcome-receipt.json", "immutable outcome receipt")
	casesPath := flags.String("cases", "fixtures/cases", "canonical cases")
	runtimePath := flags.String("runtime", "", "CI runtime observations")
	outputDir := flags.String("output-dir", "", "caller-owned output directory")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 || *outputDir == "" || !filepath.IsAbs(*outputDir) {
		fmt.Fprintln(stderr, "evaluate requires an absolute --output-dir")
		return 2
	}
	meta, err := experience.LoadMeta(*sourcePath, *contractPath, *irPath, *generatedPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fixture, err := experience.LoadFixture(*fixturePath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	records, _, err := experience.LoadMemory(*memoryPath, experience.FixtureDigest(fixture))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	var receipt *experience.OutcomeReceipt
	if *receiptPath != "" {
		loaded, err := experience.LoadReceipt(*receiptPath)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		receipt = &loaded
	}
	cases, err := experience.LoadCases(*casesPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	runtimeInput := experience.RuntimeInput{Authority: experience.Authority{}}
	if *runtimePath != "" {
		if err := experience.LoadJSON(*runtimePath, &runtimeInput); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
	}
	report, err := experience.Evaluate(meta, fixture, records, receipt, cases, runtimeInput)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := experience.WriteReport(*outputDir, report); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return writeStdout(stdout, report)
}

func writeStdout(writer io.Writer, value any) int {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return 1
	}
	return 0
}
func usage(writer io.Writer) {
	fmt.Fprintln(writer, "usage: gooo-experience-memory compile|generate|evaluate [flags]")
	fmt.Fprintln(writer, "compile --source PATH --contract PATH --output ABSOLUTE_PATH")
	fmt.Fprintln(writer, "generate --ir PATH --output ABSOLUTE_PATH")
	fmt.Fprintln(writer, "evaluate --source PATH --contract PATH --ir PATH --generated PATH --fixture PATH --memory PATH --receipt PATH --cases PATH --output-dir ABSOLUTE_PATH")
}
