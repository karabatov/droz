# Droz

*Droz* is a command-line tool for extracting and preprocessing Markdown notes from a Zettelkasten-style archive to be published as part of a [Hugo](https://gohugo.io)-generated website.

The tool is named after the [Jaquet-Droz automata](https://en.wikipedia.org/wiki/Jaquet-Droz_automata).

## What it does

Given the following structure in `~/Documents/notes`:

```
files/
  202101261901/
    image.jpg
sites/
  website.yaml
202101261901 sample note.md
```

And the note’s content:

```
# Sample note
Tags: #website_publish

Note text.

![Image](files/202101261901/image.jpg)
```

If a Hugo website is in `~/source/website`, running *Droz* will produce:

```
content/
  posts/
    sample-note/
      files/
        202101261901/
          image.jpg
    202101261901.md
```

And the post content:

```
---
title: "Sample note"
date: 2021-01-26
slug: "sample-note"
---
Note text.

![Image](files/202101261901/image.jpg)
```

This relies on post permalinks being configured like so in `config.toml`, otherwise the relative image links would be broken:

```
[permalinks]
  posts = "posts/:slug/"
```

## Configuration

The contents of `sites/website.yaml` (*Droz* configuration) is:

```yaml
publish_tags:              # Tags for copying multiple files.
  - name: website_publish  # Copy files tagged #website_publish…
    target: posts          # …to <Hugo dir>/content/posts/.
    
pages:                     # Mapping individual notes to pages.
  - id: 202101261901       # Put note with the specified id (timestamp)…
    target: about.md       # …to <Hugo dir>/content/about.md.
```

## Usage

Run `droz -to=~/source/website -config=website` in the notes directory.

* `-to=<absolute path>` is where the Hugo website is located.
* `website` is the name of the file `sites/website.yaml` in the notes directory.
* Notes directory can be specified with the parameter `-notes`.

## Requirements

* *Droz* must be run in the notes directory.
* The config named `website` must be located in `sites/website.yaml` in the note directory.

* Note file names must follow the pattern `<timestamp> some text.md`…
* …where `<timestamp>` is `YYYYMMDDHHmm`.
* Files for a note must be stored in `files/<timestamp>/*.*`.
* The first line of the note must be a first-level heading `# Heading`. It becomes the title and slug and is stripped from the file.
* The note’s tags must be on one line starting with `Tags:` separated by spaces (the line will be stripped).

These requirements may be relaxed as the tool gets developed.

## Post-initial version

* Page mapping (timestamp to URL, `"202101261901" -> /about/`)
* Tag mapping for readability (`internal_tag -> "External Tag"`).
* Tag descriptions (single tag page would have text describing the tag).
* Last modified date from the latest Git modification date.
