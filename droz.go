package main

import (
	"errors"
	"flag"
	"fmt"
	_ "gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

type PublishTag struct {
	Name   string
	Target string
}

type Page struct {
	Id     string
	Target string
}

type Config struct {
	PublishTags []PublishTag `yaml:"publish_tags"`
	Pages       []Page
}

func loadConfig(from string) (Config, error) {
	c := Config{}
	return c, errors.New("not implemented")
}

func main() {
	flag.Usage = func() {
		exe := filepath.Base(os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", exe)
		flag.PrintDefaults()
	}

	pwd := filepath.Dir(os.Args[0])
	var (
		notesDir   = flag.String("notes", pwd, "Notes `directory`.")
		hugoDir    = flag.String("to", "", "Hugo website root `directory`.")
		configName = flag.String("config", "", "Config `name` for website export.")
	)

	flag.Parse()

	if *notesDir == "" {
		flag.Set(*notesDir, pwd)
	}

	if *hugoDir == "" || *configName == "" {
		flag.Usage()
		os.Exit(1)
	}

	configFileName := filepath.Join(*notesDir, "sites", *configName+".yaml")
	_, err := loadConfig(configFileName)
	if err != nil {
		fmt.Println("Could not load config file:", err)
		os.Exit(1)
	}

	fmt.Println(*notesDir)
	fmt.Println(*hugoDir)
	fmt.Println(*configName)
	fmt.Println(configFileName)
}
