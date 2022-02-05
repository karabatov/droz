package main

import (
	"bufio"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
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
	t := map[string]string{}

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
	// Matches title line.
	titleLine = regexp.MustCompile(`^#\s(.*)$`)
)

func processPublishTags(config *Config, notesDir string, siteDir string) {
	notes, err := ioutil.ReadDir(notesDir)
	if err != nil {
		fmt.Println("Could not read notes directory:", err)
		os.Exit(1)
	}

	tagTargets := config.TagTargets(siteDir)

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

		for _, tag := range tags {
			if targetDir, ok := tagTargets[tag]; ok {
				transferNote(noteFile.Name(), notesDir, targetDir)
			}
		}
	}
}

func tagsFromFile(notePath string) ([]string, error) {
	tags := []string{}

	f, err := os.OpenFile(notePath, os.O_RDONLY, 0644)
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

func transferNote(noteFileName string, notesDir string, targetDir string) {
	fmt.Println("Exporting note", noteFileName, "to", targetDir)

	noteId := goodNoteName.FindStringSubmatch(noteFileName)[1]
	targetFileName := filepath.Join(targetDir, noteId+".md")

	// Let's make an assumption (for now) that title and tags come before other lines.
	title := ""
	slug := ""
	date := fmt.Sprintf("%s-%s-%s", noteId[:4], noteId[4:6], noteId[6:8])

	// Open source file for reading.
	fromFile, err := os.OpenFile(filepath.Join(notesDir, noteFileName), os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println("Couldn't open note:", err)
		return
	}

	defer func() {
		if err = fromFile.Close(); err != nil {
			fmt.Println("Failed to close note:", err)
		}
	}()

	// Open target file for writing.
	targetFile, err := os.OpenFile(targetFileName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Couldn't open target:", err)
		return
	}

	defer func() {
		if err = targetFile.Close(); err != nil {
			fmt.Println("Failed to close target:", err)
		}
	}()

	// Whether we have written front matter.
	frontMatter := false
	// Skip empty lines after title and tags.
	emptySkipped := false

	w := bufio.NewWriter(targetFile)
	s := bufio.NewScanner(fromFile)
	for s.Scan() {
		if !frontMatter {
			if titleLine.MatchString(s.Text()) {
				title = titleLine.FindStringSubmatch(s.Text())[1]
				slug = slugFromTitle(title)
				// TODO: Marshal YAML instead.
				w.WriteString("---\n")
				w.WriteString("title: \"" + title + "\"\n")
				w.WriteString("date: " + date + "\n")
				w.WriteString("slug: \"" + slug + "\"\n")
				w.WriteString("---\n")
				frontMatter = true
			}
			// Skip lines until we find the first level 1 heading.
			continue
		}
		if !emptySkipped {
			if s.Text() == "" {
				continue
			}
			emptySkipped = true
		}
		if tagLine.MatchString(s.Text()) {
			continue
		}
		w.WriteString(s.Text() + "\n")
	}
	w.Flush()
	copyNoteFiles(noteId, slug, notesDir, targetDir)
}

func copyNoteFiles(noteId string, slug string, notesDir string, targetDir string) {
	sourceNoteFiles := filepath.Join(notesDir, "files", noteId)
	files, err := ioutil.ReadDir(sourceNoteFiles)
	if err != nil {
		// Not copying any files.
		return
	}

	// Make sure target directory exists.
	targetNoteFiles := filepath.Join(targetDir, slug, "files", noteId)
	if os.MkdirAll(targetNoteFiles, 0644) != nil {
		return
	}

	for _, file := range files {
		if file.IsDir() {
			// Don't copy directories for now.
			continue
		}
		_, err := copyFile(filepath.Join(sourceNoteFiles, file.Name()), filepath.Join(targetNoteFiles, file.Name()))
		if err != nil {
			fmt.Println("Failed to copy file:", err)
		}
	}
}

func copyFile(from string, to string) (int64, error) {
	source, err := os.Open(from)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(to)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func slugFromTitle(title string) string {
	// Partly lifted from https://github.com/mrvdot/golang-utils/blob/master/utils.go.
	// To be improved.

	seenColon := false
	return strings.Map(func(r rune) rune {
		if seenColon {
			return -1
		}
		switch {
		case r == ' ', r == '-':
			return '-'
                case r == '+':
			return 'p'
		case r == '_', unicode.IsLetter(r), unicode.IsDigit(r):
			return r
		case r == ':':
			seenColon = true
			return -1
		default:
			return -1
		}
		return -1
	}, strings.ToLower(strings.TrimSpace(title)))
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
}
