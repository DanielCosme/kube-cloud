package main

import (
	"fmt"

	"danicos.dev/daniel/kube-deploy/pkg/services"
	"danicos.dev/daniel/kube-deploy/pkg/target"
	"github.com/magefile/mage/mg"
)

var (
	r       target.Runner
	kube    target.Target
	secrets []map[string]string
)

func init() {
	kube = target.New("kubectl")
	Env := map[string]string{}
	r = target.NewRunner(Env, nil)
	secrets = []map[string]string{
		secretspath("./config/secrets/curious-ape", "./config/enc/curious-ape"),
		secretspath("./manifests/secrets", "./manifests/enc"),
		secretspath("./pkg/secrets", "./pkg/enc"),
	}
}

func Build_Manifests() error {
	c := target.NewA("go", "run", ".", "all")
	return r.RunV("build manifests", c)
}

func Build(stack string) error {
	c := target.NewA("go", "run", ".", stack)
	return r.RunV("build manifests", c)
}

func Apply(stack string) error {
	mg.Deps(mg.F(Build, stack))
	var ts []target.Target
	if stack != "secrets" {
		ts = []target.Target{
			target.NewA("kubectl", "apply", "-f", "manifests/"+stack+"/namespace.yaml"),
		}
	}
	ts = append(ts, target.NewA("kubectl", "apply", "-f", "manifests/"+stack))
	return runSteps("apply manifests", ts)
}

func Encrypt_Secrets() error {
	// NOTE: we assume AGE_KEY env variable is populated with the path to the encryption key.
	for idx, env := range secrets {
		r := target.NewRunner(env, nil)
		enc := target.NewA("./scripts/enc_dec.fish", "enc")
		err := r.RunV(fmt.Sprintf("Encrypt %d", idx), enc)
		if err != nil {
			return err
		}
	}
	return nil
}

func Decrypt_Secrets() error {
	// NOTE: we assume AGE_KEY env variable is populated with the path to the encryption key.
	for idx, env := range secrets {
		r := target.NewRunner(env, nil)
		ts := []target.Target{
			target.NewA("mkdir", "-p", env["SECRETS_PATH"]),
			target.NewA("./scripts/enc_dec.fish", "dec"),
		}
		err := runStepsR(fmt.Sprintf("Decrtypt-%d", idx), &r, ts)
		if err != nil {
			return err
		}
	}
	return nil
}

func secretspath(path, encPath string) map[string]string {
	return map[string]string{
		"SECRETS_PATH":     path,
		"ENC_SECRETS_PATH": encPath,
	}
}

func Observe() {
	fmt.Print(services.ObservabilityNamespace)
}

func Alloy() {
	fmt.Print(services.Alloy)
}

func Grafana() {
	fmt.Print(services.Grafana)
}

func Prometheus() {
	fmt.Print(services.Prometheus)
}

func runSteps(target string, ts []target.Target) error {
	var err error
	for _, t := range ts {
		if t.Silent {
			err = r.Run(target, t)
		} else {
			err = r.RunV(target, t)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func runStepsR(target string, runner *target.Runner, ts []target.Target) error {
	var err error
	for _, t := range ts {
		if t.Silent {
			err = runner.Run(target, t)
		} else {
			err = runner.RunV(target, t)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
