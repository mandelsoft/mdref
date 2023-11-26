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
- Provide examples documentation consistent with working code.

## Command Line Syntax

```shell
$ mdref {<option>} <source folder> [<target folder>]

Flags:
  --version   just print the program version
  --help      print the help text
  --headings  prefer to use standard Markown heading anchors
  --list      print reference index and usage list 
```

If called without a target folder a consistency check is
done. It checks, that for all used references the appropriate anchors are defined.

If called with target folder, the source tree
is evaluated and the generated files are provided in the target folder.

If the `--list` option is given, additionally the reference index
and usage list is printed.


The source folder may not only contain markdown files. The generator copies all non-markdown files in the same structure to the target folder.

## General Annotation Syntax

Annotations used by this generator use a common syntax

```
{{<elementsyntax>}}
```

Elements may be [{{*anchor}}], [{{*term-anchor}}], [{{*reference}}] or [{{*commands}}].


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

## Commands

Besides the generation of consistent references among Markdown documents
the generator also supports some useful extensions to enrich the content
of the Markdown file.

The list of all supported commands can be found [here]({{commands}}).