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
TODO
```

## Post-initial version

* Page mapping (timestamp to URL, `"202101261901" -> /about/`)
* Tag mapping for readability (`internal_tag -> "External Tag"`).
* Tag descriptions (single tag page would have text describing the tag).
