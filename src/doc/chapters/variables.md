{{cmd-variable:variable}}
## Variables

It is possible to predefine variable values usable in
other commands in file path arguments (e.g. [{{cmd-include}}]) or execution argument for [{{cmd-execute}}]. They are defined by the special command `variable` which produces
no substitution.

<pre>
{{variable}{&lt;name>}{&lt;value>}&rcub;
</pre>

A variable name starts with a letter, followed by letters, digits and underscores (`_`).

Variable values are used by a substitution expression of the form `$(<name>)` inside
appropriate command arguments. Variable definitions may again use variables, but there must be no cycle.

For example, a variable for the folder of a command, based on a variable
`root` for the project folder, can be used as argument of an [{{cmd-execute}}]
command:

<pre>
{{variable}{democmd}{$(root)/democmd}&rcub;
{{execute}{go}{run}{$(democmd)}{text}{some demo text}&rcub;
</pre>

### Predefined Variables

The following variables are predefined. Their values are determined by
the generation and the document containing the substitution, and their
names cannot be used for own variable definitions:

- `DOC_ROOT`: the absolute path of the source folder used for the generation.
- `REL_DIR_TO_DOC_ROOT`: the relative path from the folder of the document
  containing the substitution to the source folder (`.` for documents
  directly in the source folder).

Own variables may use predefined variables. The predefined variables
used by an own variable are not evaluated at the location of its
definition, but for every substitution: the value of an own variable
always reflects the values of the predefined variables at the location
where it is used. Therefore, for example, a variable for the root folder
of a project can be defined once, independently of the locations of the
documents using it.

For example, with the definitions

<pre>
{{variable}{root}{$(REL_DIR_TO_DOC_ROOT)/..}&rcub;
{{variable}{democmd}{$(root)/democmd}&rcub;
</pre>

the own variables are resolved once, keeping the references to the
predefined variables:

```
root    = $(REL_DIR_TO_DOC_ROOT)/..
democmd = $(REL_DIR_TO_DOC_ROOT)/../democmd
```

For the source folder `src` (`mdref src <target>`), they are evaluated for the
documents `src/README.md` and `src/doc/chapters/variables.md` as follows:

| Expression | in `src/README.md` | in `src/doc/chapters/variables.md` |
|---|---|---|
| `$(DOC_ROOT)` | `<source folder>` | `<source folder>` |
| `$(REL_DIR_TO_DOC_ROOT)` | `.` | `../..` |
| `$(root)` | `./..` | `../../..` |
| `$(democmd)` | `./../democmd` | `../../../democmd` |

Here, `<source folder>` is the absolute path of the folder `src`. In both
documents, `$(root)` refers to the same folder (the folder containing `src`).
