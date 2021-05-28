package pipeline

import (
	"fmt"
	"reflect"
)

type StepSpec struct {
	YamlInfo YamlInfo `yaml:"_,inline"`

	Command *YamlCommandStep `yaml:"command,inline"`
	Cmake   *YamlCmakeStep   `yaml:"cmake"`
}

type StepsSpec []StepSpec

type Step interface {
	Process(conf Config) error
}

type Steps []Step

func LoadSteps(spec StepsSpec) (res Steps, err error) {
	res = make(Steps, len(spec))

	for i, s := range spec {
		if res[i], err = LoadStep(s); err != nil {
			return
		}
	}

	return
}

func LoadStep(spec StepSpec) (Step, error) {
	type creator interface {
		create() (Step, error)
	}

	v := reflect.ValueOf(spec)
	for i := 0; i < v.NumField(); i++ {
		creator, ok := v.Field(i).Interface().(creator)
		if ok {
			return creator.create()
		}
	}

	return nil, fmt.Errorf("Unknown step type")
}

func (s *Steps) Process(conf Config) error {
	for i, step := range *s {

		fmt.Printf("\033[1;34mRunning %d step\033[0m\n", i)

		if err := step.Process(conf); err != nil {
			return err
		}
	}
	return nil
}
