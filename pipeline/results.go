package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
)

type ResultSpec struct {
	YamlInfo YamlInfo `yaml:"_,inline"`

	Source      YamlStrings
	Sources     YamlStrings
	Destination string
	Rename      string

	Permission permissionSpec
}

type ResultsSpec []ResultSpec

type Results []Result

type Result struct {
	Spec ResultSpec

	sources []string
	//destination string
	//srcFiles []string
	//destDir  string
	//conf     Config
}

func LoadResults(spec ResultsSpec) (res Results, err error) {
	res = make(Results, len(spec))

	for i, s := range spec {
		if res[i], err = LoadResult(s); err != nil {
			return
		}
	}

	if len(res) == 0 {
		return res, FieldNotFoundError("results")
	}

	return
}

func LoadResult(spec ResultSpec) (res Result, err error) {
	if err = spec.validate(); err != nil {
		return
	}

	res = Result{
		Spec: spec,

		sources: []string{},
		//destination: spec.Destination,
		//conf:     Config{},
	}

	// Sources and Source are aliases
	res.sources = append(res.Spec.Sources, res.Spec.Source...)

	return res, nil
}

func (spec ResultSpec) validate() error {
	sources := append(spec.Sources, spec.Source...)

	if len(sources) == 0 {
		return FieldNotFoundError("source")
	}

	if spec.Destination == "" {
		return FieldNotFoundError("destination")
	}

	for _, s := range sources {
		if filepath.IsAbs(s) {
			return NewError(`The source "%s" is absolute.\n`+
				`The source field must be relative path or mask to the working directory.`, s)
		}
	}

	if filepath.IsAbs(spec.Destination) {
		return NewError(`The destination "%s" is absolute.\n`+
			`The destination field must be relative path to the bundle directory.`, spec.Destination)
	}

	return nil
}

// func (r *Result) LoadSpec(yml *yaml.Node) error {
// 	if err := r.doLoadSpec(yml); err != nil {
// 		return NewError("Incorrect result file on line %d", yml.Line).Wrap(err)
// 	}

// 	return nil
// }

// func (r *ResultFile) doLoadSpec(yml *yaml.Node) error {
// 	tmp := struct {
// 		Source      string
// 		Sources     stringlist
// 		Destination string
// 		Permissions permissions
// 		Skip_Dirs   bool
// 		Rename      string
// 	}{}

// 	tmp.Skip_Dirs = true

// 	if err := yml.Decode(&tmp); err != nil {
// 		return err
// 	}

// 	if tmp.Source != "" {
// 		tmp.Sources = append(tmp.Sources, tmp.Source)
// 	}

// 	if len(tmp.Sources) == 0 {
// 		return FieldNotFoundError("source")
// 	}

// 	if tmp.Destination == "" {
// 		return FieldNotFoundError("destination")
// 	}

// 	for _, s := range tmp.Sources {
// 		if filepath.IsAbs(s) {
// 			return NewError(`The source "%s" is absolute.\n`+
// 				`The source field must be relative path or mask to the working directory.`, s)
// 		}
// 	}

// 	if filepath.IsAbs(tmp.Destination) {
// 		return NewError(`The destination "%s" is absolute.\n`+
// 			`The destination field must be relative path to the bundle directory.`, tmp.Destination)
// 	}

// 	r.specLine = yml.Line
// 	r.Sources = tmp.Sources
// 	r.Destination = tmp.Destination
// 	fmt.Printf("$$$ %d\n", tmp.Permissions)
// 	r.Permission = int(tmp.Permissions)
// 	r.SkipDirs = tmp.Skip_Dirs
// 	r.Rename = tmp.Rename

// 	return nil
// }

// func (r ResultFile) copyFiles(src, dest string) error {
// 	var err error
// 	var srcFile *os.File
// 	var dstFile *os.File

// 	if srcFile, err = os.Open(src); err != nil {
// 		return fmt.Errorf("can't copy %s file to %s: %e", src, dest, err)
// 	}
// 	defer srcFile.Close()

// 	if dstFile, err = os.Create(src); err != nil {
// 		return fmt.Errorf("can't copy %s file to %s: %e", src, dest, err)
// 	}
// 	defer dstFile.Close()

// 	if _, err = io.Copy(dstFile, srcFile); err != nil {
// 		return fmt.Errorf("can't copy %s file to %s: %e", src, dest, err)
// 	}

// 	return nil
// }

func (r Results) Process(conf Config) error {
	for _, f := range r {
		if err := f.Process(conf); err != nil {
			return err
		}
	}
	return nil
}

func (r Result) Process(conf Config) error {
	if err := r.doProcess(conf); err != nil {
		res := NewError("Failed to copy result file (defined on %d line)\n - Sources: %s\n - Destination: %s", r.Spec.YamlInfo.SpecLine, r.sources, r.Spec.Destination)
		res.Wrap(err)
		return res
	}

	return nil
}

func (r *Result) doProcess(conf Config) error {

	destDir := conf.Env.Expand(r.Spec.Destination)
	if !filepath.IsAbs(destDir) {
		destDir = conf.BundleDir + "/" + destDir
	}

	srcFiles, err := r.findSrcFiles(conf.WorkDir)
	if err != nil {
		return err
	}

	// check ..........................
	if len(srcFiles) == 0 {
		return fmt.Errorf("No source files found")
	}

	if len(srcFiles) > 1 && r.Spec.Rename != "" {
		return fmt.Errorf("You cannot use the mask for source files and the rename parameter at the same time.")
	}

	// Do run ........................
	if err := createDir(destDir); err != nil {
		return err
	}

	for _, src := range srcFiles {

		dest := ""
		if r.Spec.Rename != "" {
			dest = destDir + "/" + r.Spec.Rename
		} else {
			dest = destDir + "/" + filepath.Base(src)
		}

		fi, err := os.Stat(src)
		if err != nil {
			return err
		}

		perm := fi.Mode().Perm()
		if r.Spec.Permission > 0 {
			perm = os.FileMode(r.Spec.Permission)
		}

		switch srcMode := fi.Mode(); {

		case srcMode.IsRegular():
			if err := r.copyRegular(src, dest, perm); err != nil {
				return err
			}

		case srcMode.IsDir():
			if err := r.copyDir(src, dest, perm); err != nil {
				return err
			}

		}
	}

	return nil
}

func (r Result) findSrcFiles(dir string) ([]string, error) {
	res := []string{}

	for _, src := range append(r.Spec.Sources, r.Spec.Source...) {

		found, err := filepath.Glob(dir + "/" + src)
		if err != nil {
			return res, fmt.Errorf(`Cannot access "%s". %s.`, src, err)
		}

		if len(found) == 0 {
			return res, fmt.Errorf(`Cannot access "%s". No such file.`, src)
		}

		res = append(res, found...)
	}
	return res, nil
}

func (r Result) copyRegular(src, dest string, perm os.FileMode) error {
	//fmt.Printf("@@@ copy file\n\tsrc:  %s\n\tdest: %s\n\t%d\n", src, dest, perm)
	if err := copyFileWithPerm(src, dest, os.FileMode(perm)); err != nil {
		return err
	}
	return nil
}

func (r Result) copyDir(src, dest string, perm os.FileMode) error {
	if isFileExists(dest) {
		return fmt.Errorf("Can't copy result file %s to %s, file already exists.", src, dest)
	}

	fmt.Printf("copy dir\n\tsrc:  %s\n\tdest: %s\n", src, dest)

	return nil
}

// type resultFile struct {
// 	sources     []string
// 	destination string
// 	permission  int
// }

// func loadResultFile(yml yaml.Node) (resultFile, error) {
// 	res := resultFile{}

// 	{
// 		//		tmp := struct {			Source: string `yaml:"source"`}
// 		//err := yml.Decode(tmp)
// 		//if
// 	}

// 	//	tmp := struct {
// 	//Source: string `yaml:"source"`,
// 	///Destination: string `yaml:"des"`,
// 	//Permission : int `yaml:"permission"`
// 	//}{}

// 	//permission: 0666,

// 	return res, nil
// }

// // func (r *fileResult) Run() error {
// // 	return nil
// // }
