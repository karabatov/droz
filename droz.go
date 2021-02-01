package main

import (
	"bufio"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
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

func (config *Config) TagTargets(siteDir string) map[string]string {
	var t map[string]string

	for _, tag := range config.PublishTags {
		t[tag.Name] = filepath.Join(siteDir, "content", tag.Target)
	}

	return t
}

func loadConfig(from string) (Config, error) {
	c := Config{}

	file, err := ioutil.ReadFile(from)
	if err != nil {
		return c, err
	}
	err = yaml.Unmarshal(file, &c)
	return c, err
}

var (
	// Matches "202102012138 note title.md" and "202102012138.md".
	goodNoteName = regexp.MustCompile(`^(\d{12}).*\.md$`)
	// Matches the line with tags.
	tagLine = regexp.MustCompile(`^Tags: `)
	// Matches one tag without the pound sign.
	oneTag = regexp.MustCompile(`#(\S+)\s*`)
)

func processPublishTags(config *Config, notesDir string, siteDir string) {
	notes, err := ioutil.ReadDir(notesDir)
	if err != nil {
		fmt.Println("Could not read notes directory:", err)
		os.Exit(1)
	}

	//tagTargets := config.TagTargets(siteDir)

	for _, noteFile := range notes {
		if noteFile.IsDir() || !goodNoteName.MatchString(noteFile.Name()) {
			continue
		}
		notePath := filepath.Join(notesDir, noteFile.Name())
		tags, err := tagsFromFile(notePath)
		if err != nil {
			fmt.Println("Error reading file, skipping", noteFile.Name())
			continue
		}
		fmt.Println(tags)
	}
}

func tagsFromFile(notePath string) ([]string, error) {
	tags := []string{}

	f, err := os.Open(notePath)
	if err != nil {
		return tags, err
	}

	defer func() {
		if err = f.Close(); err != nil {
			fmt.Println("Failed to close file:", notePath)
		}
	}()

	s := bufio.NewScanner(f)
	for s.Scan() {
		if !tagLine.MatchString(s.Text()) {
			continue
		}

		for _, tagPair := range oneTag.FindAllStringSubmatch(s.Text(), -1) {
			tags = append(tags, tagPair[1])
		}

		break
	}

	return tags, nil
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

	// Load config.

	configFileName := filepath.Join(*notesDir, "sites", *configName+".yaml")
	config, err := loadConfig(configFileName)
	if err != nil {
		fmt.Println("Could not load config file:", err)
		os.Exit(1)
	}

	// TODO: Prepare tag mapping.

	// Process publish tags.

	if len(config.PublishTags) > 0 {
		processPublishTags(&config, *notesDir, *hugoDir)
	}

	// TODO: Process pages.

	fmt.Println(*notesDir)
	fmt.Println(*hugoDir)
	fmt.Println(*configName)
	fmt.Println(configFileName)
	fmt.Println(config)
}
