package pipeline

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// YamlCommandStep defines command step fileds
type YamlCommandStep struct {
	YamlInfo YamlInfo `yaml:"_,inline"`

	Command YamlCommand
}

type CommandStep struct {
	command  []string
	specLine int
}

//lint:ignore U1000 used with reflection
func (spec YamlCommandStep) create() (Step, error) {
	res := CommandStep{
		command:  spec.Command,
		specLine: spec.YamlInfo.SpecLine,
	}

	return &res, nil
}

// func NewCommandStep(spec YamlCommandStep) (res CommandStep, err error) {
// 	res = CommandStep{
// 		command: spec.Command,
// 	}

// 	//return fmt.Errorf("incorrect command on line %d: %s", yml.Line, err)

// 	return res, nil
// }

// func parseCommandYaml(yml yaml.Node) ([]string, error) {
// 	mkError := func(err interface{}) error {
// 		return fmt.Errorf("incorrect command on line %d: %s", yml.Line, err)
// 	}

// 	res := []string{}
// 	var err error
// 	if yml.Kind == yaml.ScalarNode {
// 		var s string
// 		if err = yml.Decode(&s); err != nil {
// 			return res, mkError(err)
// 		}

// 		res, err = shellquote.Split(s)
// 		if err != nil {
// 			return res, mkError(err)
// 		}
// 		return res, nil
// 	}

// 	if yml.Kind == yaml.SequenceNode {
// 		if err = yml.Decode(&res); err != nil {
// 			return res, mkError(err)
// 		}
// 		return res, nil
// 	}

// 	return res, mkError("unsupported YAML type")

// }

// func (s *CommandStep) LoadSpec(yml yaml.Node) error {
// 	var err error
// 	s.specLine = yml.Line
// 	s.command, err = parseCommandYaml(yml)
// 	return err
// }

func (s *CommandStep) Process(conf Config) error {
	if err := s.doRun(conf); err != nil {
		res := NewError("Failed to process command step (defined on %d line)\n - command: %v", s.specLine, s.command)
		res.Wrap(err)
		return res
	}

	return nil
}

func (s *CommandStep) doRun(conf Config) error {
	args := make([]string, len(s.command))
	for i, a := range s.command {
		args[i] = conf.Env.Expand(a)
	}

	switch args[0] {
	case "set", "export":
		return s.callSetCommand(args, conf.Env)

	case "unset":
		return s.callUnsetCommand(args, conf.Env)
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = conf.Env.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	//fmt.Printf("@@@ %#v\n", args)
	err := cmd.Run()
	if err != nil {
		return NewError("The command '%s' retuned a non-zerro code.",
			strings.Join(cmd.Args, " ")).Wrap(err)

		//		return fmt.Errorf("the command '%s' retuned a non-zerro code: %s",
		//			strings.Join(cmd.Args, " "),
		//			err)
	}

	return nil
}

// callUnsetCommand method print or set environment variable value
func (s *CommandStep) callSetCommand(args []string, env *Environ) error {
	if len(args) == 1 {
		for _, e := range os.Environ() {
			fmt.Println(e)
		}
		return nil
	}

	for _, arg := range args[1:] {
		kv := strings.SplitN(arg, "=", 2)
		if len(kv) < 2 {
			return fmt.Errorf("incorrect command: %s", strings.Join(args, " "))
		}

		env.Setenv(kv[0], kv[1])
	}
	return nil
}

// callUnsetCommand method unset environment variable
func (s *CommandStep) callUnsetCommand(args []string, env *Environ) error {
	if len(args) == 1 {
		return fmt.Errorf("%s: not enough argument", args[0])
	}

	for _, arg := range args[1:] {
		env.Unsetenv(arg)
	}
	return nil
}
