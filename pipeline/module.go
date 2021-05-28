package pipeline

import (
	"os"
)

type ModuleSpec struct {
	YamlInfo YamlInfo `yaml:"_,inline"`

	Name string

	Steps   []StepSpec `yaml:"steps"`
	Results []ResultSpec
}

type Module struct {
	YamlInfo YamlInfo

	Name string

	Steps   Steps
	Results Results
}

type Modules []Module

func LoadModules(spec []ModuleSpec) (res Modules, err error) {
	res = make([]Module, len(spec))

	for i, yamlModule := range spec {
		if res[i], err = LoadModule(yamlModule); err != nil {
			return
		}
	}

	return
}

func LoadModule(spec ModuleSpec) (res Module, err error) {
	defer func() {
		if err != nil {
			err = NewError(`Failed to load module "%s" (defined on %d line)`, spec.Name, spec.YamlInfo.SpecLine).Wrap(err)
		}
	}()

	res = Module{
		YamlInfo: spec.YamlInfo,
		Name:     spec.Name,
	}

	if res.Steps, err = LoadSteps(spec.Steps); err != nil {
		return
	}

	if res.Results, err = LoadResults(spec.Results); err != nil {
		return
	}

	return
}

func (m *Module) Process(conf Config) (err error) {
	defer func() {
		if err != nil {
			err = NewError("Failed to process module %s (defined on %d line)", m.Name, m.YamlInfo.SpecLine).Wrap(err)
		}
	}()

	// Change working directory ............
	prevDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(prevDir)

	// Prepare .............................
	if err = m.prepare(conf); err != nil {
		return
	}

	// Process Steps ........................
	if err = m.Steps.Process(conf); err != nil {
		return
	}

	// Process Results ......................
	//fmt.Printf("\033[1;34mCopy results files\033[0m\n")
	if err = m.Results.Process(conf); err != nil {
		return
	}

	return nil
}

func (m *Module) prepare(conf Config) error {
	if err := createDir(conf.WorkDir); err != nil {
		return err
	}

	if err := os.Chdir(conf.WorkDir); err != nil {
		return err
	}

	return nil
}
