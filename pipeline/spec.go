package pipeline

import (
	"fmt"
	"io/fs"
	"os"
	"strconv"

	"github.com/kballard/go-shellquote"
	"gopkg.in/yaml.v3"
)

type Spec struct {
	YamlInfo  YamlInfo        `yaml:"_,inline"`
	Propeties GlobalPropeties `yaml:"_,inline"`

	Modules []ModuleSpec
}

type GlobalPropeties struct {
	AppName       string `yaml:"app-name"`
	Version       string `yaml:"version"`
	CertificateID string `yaml:"certificate-id"`
}

type YamlCmakeStep struct {
	YamlInfo YamlInfo `yaml:"_,inline"`

	Args []string
}

func (spec Spec) handleErr(err error) error {
	str := err.Error()
	//fmt.Printf("%T\n", err)
	//str = strings.ReplaceAll(str, "not found in type pipeline.Spec", "is not allowed in ")
	return fmt.Errorf(str)
}

func (y *Spec) LoadSpec(fileName string) error {

	specFile, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer specFile.Close()

	dec := yaml.NewDecoder(specFile)
	dec.KnownFields(true)
	if err := dec.Decode(y); err != nil {
		return y.handleErr(err)
	}

	return nil
}

type YamlInfo struct {
	SpecLine int
}

func (y *YamlInfo) UnmarshalYAML(value *yaml.Node) error {
	y.SpecLine = value.Line
	return nil
}

type YamlStrings []string

func (s *YamlStrings) UnmarshalYAML(value *yaml.Node) error {

	if value.Kind == yaml.ScalarNode {
		var v string
		if err := value.Decode(&v); err != nil {
			return err
		}

		*s = []string{v}
		return nil
	}

	if value.Kind == yaml.SequenceNode {
		var v []string
		if err := value.Decode(&v); err != nil {
			return err
		}

		*s = []string(v)
		return nil
	}

	return nil
}

type YamlCommand []string

func (res *YamlCommand) UnmarshalYAML(yml *yaml.Node) error {
	mkError := func(err interface{}) error {
		return fmt.Errorf("incorrect command on line %d: %s", yml.Line, err)
	}

	var err error
	if yml.Kind == yaml.ScalarNode {
		var s string
		if err = yml.Decode(&s); err != nil {
			return mkError(err)
		}

		*res, err = shellquote.Split(s)
		if err != nil {
			return mkError(err)
		}
		return nil
	}

	if yml.Kind == yaml.SequenceNode {
		v := []string{}
		if err = yml.Decode(&v); err != nil {
			return mkError(err)
		}
		*res = v
		return nil
	}

	return mkError("Unsupported module command type")
}

// permissionSpec - represent unix file permission
type permissionSpec fs.FileMode

func (p *permissionSpec) UnmarshalYAML(value *yaml.Node) error {
	s := ""

	if err := value.Decode(&s); err != nil {
		return err
	}

	n, err := strconv.ParseInt(s, 8, 32)
	if err != nil {
		return err
	}
	*p = permissionSpec(n)
	return nil
}
