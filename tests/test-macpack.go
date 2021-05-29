package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jessevdk/go-flags"
)

var (
	verbose = false
)

// type TestSpecFile struct {
// 	Args []string
// }

type File struct {
	path string
	info os.FileInfo
}

func scanDir(dir string) (map[string]File, error) {
	res := map[string]File{}

	l := len(dir) + 1
	err := filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if path == dir {
				return nil
			}

			res[path[l:]] = File{
				path: path,
				info: info,
			}
			return nil
		})

	return res, err
}

func compare(path string, expected, actual File) error {
	if expected.info.Mode().IsDir() && !actual.info.Mode().IsDir() {
		return fmt.Errorf("expected directory, but actual is regular file")
	}

	if !expected.info.Mode().IsDir() && actual.info.Mode().IsDir() {
		return fmt.Errorf("expected regular file, but actual is directory")
	}

	if expected.info.Mode().Perm() != actual.info.Mode().Perm() {
		return fmt.Errorf(
			"file has diferent permissions\n"+
				" - expected: %s \n"+
				" - actual:   %s ",
			expected.info.Mode().Perm(),
			actual.info.Mode().Perm())
	}

	if expected.info.IsDir() {
		return nil
	}

	aFile, err := os.Open(expected.path)
	if err != nil {
		return err
	}
	defer aFile.Close()

	aContent, err := ioutil.ReadAll(aFile)
	if err != nil {
		return err
	}

	eFile, err := os.Open(actual.path)
	if err != nil {
		return err
	}
	defer eFile.Close()

	eContent, err := ioutil.ReadAll(eFile)
	if err != nil {
		return err
	}

	if !bytes.Equal(aContent, eContent) {
		return fmt.Errorf("the contents of the files are different")
	}

	return nil
}

func check(actualDir, expectedDir string) error {
	aFiles, err := scanDir(actualDir)
	if err != nil {
		return err
	}

	eFiles, err := scanDir(expectedDir)
	if err != nil {
		return err
	}

	for path, eInfo := range eFiles {
		aInfo, ok := aFiles[path]
		if !ok {
			return fmt.Errorf("expected file is missing: %s", path)
		}
		delete(aFiles, path)

		if err := compare(path, eInfo, aInfo); err != nil {
			return fmt.Errorf("%s:\n - expected: %s\n - actual:   %s", err, eInfo.path, aInfo.path)
		}
	}

	if len(aFiles) > 0 {
		missed := []string{}
		for path, info := range aFiles {
			if info.info.IsDir() {
				path = path + "/"
			}

			missed = append(missed, "  "+path+"")
		}

		sort.Strings(missed)
		return fmt.Errorf("unexpected files and directories:\n%s", strings.Join(missed, "\n"))
	}

	return nil
}

func runExpectedScript(testDir, workDir string) error {

	scriptFile := testDir + "/expected.sh"
	expectedDir := workDir + "/expected"

	if err := os.MkdirAll(expectedDir, 0777); err != nil {
		return err
	}

	if _, err := os.Stat(scriptFile); os.IsNotExist(err) {
		return fmt.Errorf(`File "%s" not found`, scriptFile)
	}

	cmd := exec.Command("/bin/sh", scriptFile)
	cmd.Dir = expectedDir

	if verbose {
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func runTest(binPath, testDir, workDir string) error {

	//	name := filepath.Base(testDir)

	prevDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(prevDir)

	if err := os.Chdir(testDir); err != nil {
		return err
	}

	if err := os.RemoveAll(workDir); err != nil {
		return fmt.Errorf(`can't remove directory "%s":\n  %s`, workDir, err)
	}

	if err := os.MkdirAll(workDir, 0777); err != nil {
		return err
	}

	if err := runExpectedScript(testDir, workDir); err != nil {
		return err
	}
	// scriptFile := testDir + "/expected.sh"

	// data, err := ioutil.ReadFile(scriptFile)
	// if err != nil {
	// 	return err
	// }

	// spec := TestSpecFile{}

	// if err = yaml.Unmarshal(data, &spec); err != nil {
	// 	return err
	// }

	args := []string{
		"-w", workDir,
		"-o", workDir,
	}

	//args = append(args, spec.Args...)

	curDir, err := os.Getwd()
	if err != nil {
		return err
	}

	if verbose {
		fmt.Println("------------------------")
		fmt.Println("Binary:   ", binPath)
		fmt.Println("Test dir: ", testDir)
		fmt.Println("Out dir:  ", workDir)
		fmt.Println("Current dir: ", curDir)
		fmt.Println("Args: ", args)
		fmt.Println("........................")
	}

	cmd := exec.Command(binPath, args...)
	if verbose {
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	appDirs, err := filepath.Glob(workDir + "/*.app")
	if err != nil {
		return nil
	}

	if len(appDirs) > 0 {
		name := filepath.Base(appDirs[0])
		err := check(workDir+"/"+name, workDir+"/expected")
		if err != nil {
			return err
		}
	}

	return nil
}

var opts struct {
	Args struct {
		Test string
	} `positional-args:"yes"`

	Verbose  bool   `short:"v" long:"verbose" description:"Print debug information"`
	List     bool   `short:"l" long:"list" description:"print full list of the tests"`
	Binary   string `short:"b" long:"binary" default:"../macpack" description:"Specifies a macpack binary"`
	TestsDir string `short:"i" long:"inputdir" default:"./testdata" description:"Specifies a directory from which the tests are read"`
	OutDir   string `short:"o" long:"outdir" default:"./OUT" description:"Specifies a destination directory where output files are to be created"`
}

func main() {
	args := flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash|flags.HelpFlag)
	if _, err := args.Parse(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	verbose = opts.Verbose

	testDir, err := filepath.Abs(opts.TestsDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	binPath, err := filepath.Abs(opts.Binary)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if opts.List {
		dirs, err := filepath.Glob(testDir + "/*")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}

		for _, dir := range dirs {
			name := filepath.Base(dir)
			fmt.Println(name)
		}
		return
	}

	outDir, err := filepath.Abs(opts.OutDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	pattern := "*"
	if opts.Args.Test != "" {
		pattern = opts.Args.Test
	}

	dirs, err := filepath.Glob(testDir + "/" + pattern)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	passed := 0
	failed := 0
	for _, testDir := range dirs {
		name := filepath.Base(testDir)
		fmt.Printf("%s\n", name)
		err := runTest(binPath, testDir, outDir+"/"+name)
		if err == nil {
			fmt.Printf("  PASS\n")
			passed++
		} else {
			s := strings.Replace(err.Error(), "\n", "\n  ", -1)
			fmt.Fprintf(os.Stderr, "FAILED\n  %s\n", s)
			failed++
		}
	}

	fmt.Printf("\nTotals: %d passed, %d failed\n", passed, failed)
	if failed == 0 {
		os.Exit(3)
	}
	os.Exit(0)

}
