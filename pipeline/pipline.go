package pipeline

import (
	"fmt"
)

type Pipeline struct {
	Propeties GlobalPropeties

	WorkDir string
	OutDir  string

	Modules []Module
}

func New() Pipeline {
	return Pipeline{}
}

func (p *Pipeline) Load(specFile string) error {
	if err := p.doLoad(specFile); err != nil {
		return NewError(`Failed to load SPEC`).Wrap(err)
	}
	return nil
}

func (p *Pipeline) doLoad(specFileName string) error {

	spec := Spec{}
	if err := spec.LoadSpec(specFileName); err != nil {
		return err
	}

	if spec.Propeties.AppName == "" {
		return fmt.Errorf("Property %s not specified", "app-name")
	}

	p.Propeties = spec.Propeties

	var err error
	if p.Modules, err = LoadModules(spec.Modules); err != nil {
		return err
	}

	return nil
}

func (p *Pipeline) Process() error {
	fmt.Printf("*************************************************\n")
	fmt.Printf("* Name:    %v\n", p.Propeties.AppName)
	fmt.Printf("* Version: %v\n", p.Propeties.Version)
	fmt.Printf("* CertificateId: %v\n", p.Propeties.CertificateID)
	fmt.Printf("*************************************************\n")

	if err := p.doProcess(); err != nil {
		return err
	}

	return nil
}

func (p *Pipeline) doProcess() error {

	if err := p.prepareRun(); err != nil {
		return err
	}

	for i, module := range p.Modules {

		workDir := fmt.Sprintf("%s/%03d-%s",
			p.WorkDir,
			i+1,
			safeString(module.Name))

		conf := p.newConf(workDir)

		if err := module.Process(conf); err != nil {
			return err
		}
	}

	return nil
}

func (p Pipeline) prepareRun() error {
	conf := p.newConf(p.WorkDir)

	if err := deletDir(conf.BundleDir); err != nil {
		return err
	}

	if err := createDir(conf.BundleDir); err != nil {
		return err
	}

	return nil
}

func (p Pipeline) newConf(workDir string) Config {
	conf := Config{}
	conf.AppName = p.Propeties.AppName
	conf.Version = p.Propeties.Version
	conf.CertificateID = p.Propeties.CertificateID

	conf.initDirs(p.OutDir, workDir)
	conf.initEnv()

	return conf
}
