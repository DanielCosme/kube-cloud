package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"danicos.dev/daniel/go-kube/pkg/stack"
	"danicos.dev/daniel/kube-deploy/pkg/curiousape"
	"danicos.dev/daniel/kube-deploy/pkg/gitea"
	"danicos.dev/daniel/kube-deploy/pkg/observe"
	"danicos.dev/daniel/kube-deploy/pkg/proxy"
	"danicos.dev/daniel/kube-deploy/pkg/secrets"
	"danicos.dev/daniel/kube-deploy/pkg/static"
	"danicos.dev/daniel/kube-deploy/pkg/temporal"
)

func main() {
	var err error
	var command string
	if len(os.Args) > 1 {
		command = os.Args[1]
	} else {
		log.Fatal("need one argument")
	}
	fmt.Println("Building: " + command)

	stacks := map[string]stack.Stack{
		"gitea":         gitea.Stack(),
		"curious-ape":   curiousape.Stack(),
		"secrets":       secrets.Stack(),
		"web-pages":     static.Stack(),
		"observability": observe.Stack(),
		"temporal":      temporal.Stack(),
		"proxy":         proxy.Stack(),
	}

	parentDir := "manifests"
	if s, ok := stacks[command]; ok {
		err = s.MarshalYaml(parentDir)
	} else if command == "all" {
		for _, v := range stacks {
			err = v.MarshalYaml(parentDir)
		}
	} else {
		err = errors.New("stack not found")
	}
	if err != nil {
		log.Fatal(err)
	}
}
