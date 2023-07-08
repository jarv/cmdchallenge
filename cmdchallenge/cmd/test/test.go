package main

import (
	_ "embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

type ChInfo struct {
	Slug *string `yaml:"slug,omitempty"`
}

//go:embed challenges.yaml
var challengesYAML string

func main() {

	fmt.Println("hello world")
	var c []ChInfo

	if err := yaml.Unmarshal([]byte(challengesYAML), &c); err != nil {
		panic(err)
	}

	fmt.Println(*c[0].Slug)

}
