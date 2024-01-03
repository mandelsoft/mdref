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
$ {{execute}{go}{run}{..}{--syntax}}
```

If called without a target folder a consistency check is
done. It checks, that for all used references the appropriate anchors are defined.

If called with target folder, the source tree
is evaluated and the generated files are provided in the target folder.

If the `--list` option is given, additionally the reference index
and usage list is printed.


The source folder may not only contain markdown files. The generator copies all non-markdown files in the same structure to the target folder.

## What it does

The command scans a a folder tree for markdown file (`.md`) and processes special
[annotations]({{annotations}}).
The result is copied to a target folder keeping the original sub folder hierarchy.
Non-markdown files are just copied to the target folder tree.
While scanning the source folder tree folders with the name `local` are ignored.
Files contained in those folders might be used by special commands used in 
processed markdown files without being copied to the target folder hierarchy.

Processing means to resolve [references]({{overview-references}}) and evaluate some
[commands]({{overview-commands}}).

{{annotations}}
## General Annotation Syntax

Annotations used by this generator use a common syntax

```
{{<elementsyntax>}[{<argument>}...]}
```

Elements may be [{{*anchor}}], [{{*term-anchor}}], [{{*reference}}] or [{{*commands}}].


{{overview-references}}
## Reference and Anchor Syntax

Anchors and references are character sequences
following the regular expression

{{name}}

```regexp
[a-z][a-z0-9.-]*
```

These names are location independent, so content
may be copied or even moved into a completely different folder structure without corrupting any link in the generated target folder.

Therefore, the anchors must be globally unique in the
complete document tree.

The generator supports two kinds of references as well as anchors:
- [{{*Reference}}] 
- [{{*Term}}]

{{overview-commands}}
## Commands

Besides the generation of consistent references among Markdown documents
the generator also supports some useful extensions to enrich the content
of the Markdown file.

Such *commands* are described by [annotations]({{annotations}}) using
arguments.

The list of all supported commands can be found [here]({{commands}}).