# Markdown Reference Generator


This markdown reference generator uses
a document tree with markdown files containing
a special syntax for tags and references
as input and generates an appropriate target tree with
references resolved to resolved Markdown links
and anchors.

- Never fix heading anymore with corrupting links all over the document tree
- Never move text blocks or even complete files around in the document tree with corrupting links all over the document tree
- Use terms all over the document tree, which are automatically linked to their explanation.

## Command Line Syntax

```shell
$ mdref [--list] <source folder> [<target folder>]
```

If called with a target folder a consistency check is
done. It checks, that for all used references the appropriate anchors are defined.

If called with target folder, the source tree
is evaluated and the generated files are provided in the target folder.

If the `--list` option is given, only the check is done and the reference index is printed.


The source folder may not only contain markdown files. The generator copies all non-markdown files in the same structure to the target folder.

## Reference and Anchor Syntax

anchors and references are character sequences
following the regular expression

<a id="name"></a>

```regexp
[a-z][a-z0-9.-]*
```

These names are location independent, so content
may be copied or even moved into a completely different folder structure without corrupting any link in the generated target folder.

Therefore, the anchors must be globally unique in the
complete document tree.

The generator supports two kinds of references:
- [Reference targets](doc/chapters/references.md#reference) 
- [Terms](doc/chapters/terms.md#term)
