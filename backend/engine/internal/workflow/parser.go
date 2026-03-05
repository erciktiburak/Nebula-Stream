package workflow

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Definition struct {
	Version  string         `yaml:"version"`
	Name     string         `yaml:"name"`
	Triggers []Trigger      `yaml:"triggers"`
	Steps    []Step         `yaml:"steps"`
	Meta     map[string]any `yaml:"meta,omitempty"`
}

type Trigger struct {
	Type string `yaml:"type"`
}

type Step struct {
	ID    string         `yaml:"id"`
	Type  string         `yaml:"type"`
	Input map[string]any `yaml:"input,omitempty"`
}

func ParseFile(path string) (Definition, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Definition{}, fmt.Errorf("read workflow file: %w", err)
	}

	return ParseYAML(raw)
}

func ParseYAML(raw []byte) (Definition, error) {
	var def Definition
	if err := yaml.Unmarshal(raw, &def); err != nil {
		return Definition{}, fmt.Errorf("decode workflow yaml: %w", err)
	}

	if err := def.Validate(); err != nil {
		return Definition{}, err
	}

	return def, nil
}

func (d Definition) Validate() error {
	if d.Version == "" {
		return errors.New("workflow version is required")
	}

	if d.Name == "" {
		return errors.New("workflow name is required")
	}

	if len(d.Triggers) == 0 {
		return errors.New("at least one trigger is required")
	}

	for i, t := range d.Triggers {
		if t.Type == "" {
			return fmt.Errorf("trigger[%d] type is required", i)
		}
	}

	if len(d.Steps) == 0 {
		return errors.New("at least one step is required")
	}

	seen := make(map[string]struct{}, len(d.Steps))
	for i, step := range d.Steps {
		if step.ID == "" {
			return fmt.Errorf("step[%d] id is required", i)
		}
		if step.Type == "" {
			return fmt.Errorf("step[%d] type is required", i)
		}
		if _, ok := seen[step.ID]; ok {
			return fmt.Errorf("duplicate step id: %s", step.ID)
		}
		seen[step.ID] = struct{}{}
	}

	return nil
}
