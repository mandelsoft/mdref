<!-- DO NOT MODIFY                   -->
<!-- this file is generated by mdref -->
<!-- from src/README.md              -->

# Markdown Reference Generator

This markdown reference generator uses
a document tree with markdown files containing
a special annotation syntax for tags and references
as input and generates an appropriate target tree with
references resolved to consistent Markdown links
and anchors.

- Never fix headings anymore with corrupting links all over the document tree
- Never move text blocks or even complete files around in the document tree with corrupting links all over the document tree
- Use terms all over the document tree, which are automatically linked to their explanation.
- Provide example documentation consistent with working code.


## Command Line Syntax

```shell
$ mdref {<options>} [<source dir> [<target dir>]]

Flags:
  --version      just print the program version
  --help         this help text
  --headings     prefer using standard heading anchors
  --skip-execute omit the evaluation of the execute statement (for test purposes, only)
  --skip-source  omit source reference in generation comment
  --list         print reference index and usage list
```

If called without a target folder a consistency check is
done. It checks, that for all used references the appropriate anchors are defined.

If called with target folder, the source tree
is evaluated and the generated files are provided in the target folder.

If the `--list` option is given, additionally the reference index
and usage list is printed.


The source folder may not only contain markdown files. The generator copies all non-markdown files in the same structure to the target folder.

## What it does

The command scans a folder tree for markdown files (`.md`) and processes special
[annotations](#general-annotation-syntax).
The result is copied to a target folder preserving the original sub folder hierarchy.
Non-markdown files are just copied to the target folder tree.
While scanning the source folder tree folders with the name `local` are ignored.
Files contained in those folders might be used by special [commands](doc/chapters/commands.md) used in 
processed markdown files without being copied to the target folder hierarchy.

Processing means to resolve [references](#reference-and-anchor-syntax) and evaluate some
[commands](#commands).

## General Annotation Syntax

Annotations used by this generator use a common syntax

```
{{<elementsyntax>}[{<argument>}...]}
```

Elements may be [anchors](doc/chapters/references.md#anchors), [term anchors](doc/chapters/terms.md#anchors), [references](doc/chapters/references.md) or [commands](doc/chapters/commands.md).


## Reference and Anchor Syntax

Anchors and references are character sequences
following the regular expression

<a id="name"></a>

```regexp
[a-z][a-z0-9.-]*
```

These names are location independent, so content
may be copied or even moved into a completely different folder structure without corrupting any link in the generated target folder.

Therefore, the anchors must be globally unique in the
complete document tree.

The generator supports two kinds of references as well as anchors:
- [References](doc/chapters/references.md) 
- [Terms](doc/chapters/terms.md)

## Commands

Besides the generation of consistent references among Markdown documents
the generator also supports some useful extensions to enrich the content
of the Markdown file.

Such *commands* are described by [annotations](#general-annotation-syntax) using
arguments.

The list of all supported commands can be found [here](doc/chapters/commands.md).