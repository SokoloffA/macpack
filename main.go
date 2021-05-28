package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SokoloffA/macpack/pipeline"
	"github.com/jessevdk/go-flags"
)

const specFileExt = "yaml"

var opts struct {
	Args struct {
		SpecFile string
	} `positional-args:"yes"`

	Verbose bool   `short:"v" long:"verbose" description:"Print debug information during command processing"`
	WorkDir string `short:"w" long:"workdir" default:"./.tmp" description:"Specifies a destination directory where temporary files are to be created"`
	OutDir  string `short:"o" long:"outdir" default:"." description:"Specifies a destination directory where bundle is to be created"`
}

func findSpec(arg string) (string, error) {
	if arg != "" {
		if _, err := os.Stat(arg); os.IsNotExist(err) {
			return "", fmt.Errorf("file %s not found", arg)
		}

		return arg, nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	files, err := filepath.Glob(fmt.Sprintf("*.%s", specFileExt))
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", fmt.Errorf("can't find .%s file in the %s directory", specFileExt, dir)
	}

	if len(files) > 1 {
		return "", fmt.Errorf("the %s directory contains several .%s files", dir, specFileExt)
	}

	return files[0], nil
}

func main() {
	args := flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash|flags.HelpFlag)
	// p.SubcommandsOptional = true

	if _, err := args.Parse(); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	specFile, err := findSpec(opts.Args.SpecFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	workDir := opts.WorkDir
	if !filepath.IsAbs(workDir) {
		if workDir, err = filepath.Abs(workDir); err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	}

	outDir := opts.OutDir
	if !filepath.IsAbs(outDir) {
		if outDir, err = filepath.Abs(outDir); err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	}

	pipeline := pipeline.New()
	pipeline.WorkDir = workDir
	pipeline.OutDir = outDir

	if err = pipeline.Load(specFile); err != nil {
		fatal(err)
	}

	if err = pipeline.Process(); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "\033[0;31m%s\033[0m\n", "Error!")
	fmt.Fprintf(os.Stderr, "\033[0;31m%s\033[0m\n", err)
	os.Exit(3)
}
