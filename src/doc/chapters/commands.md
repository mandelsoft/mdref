## Commands
{{commands:command}}

Commands provide additional functionality to Markdown files.
All commands use the same basic annotation syntax:

```
{{<name>}{<arg>}...}
```

The following commands are supported:

- [`include`]({{cmd-include}}) include content of other file
- [`execute`]({{cmd-execute}}) unclude output of command execution

{{cmd-include}}
### Include

The include command uses the following syntax
<pre>
{{include}{&lt;filepath>}[{[&lt;startline>][:[&lt;endline>]]}]}
{{include}{&lt;filepath>}{&lt;key>}&rcub;
</pre>

this command included content from some file into the actual
markdown file.

In the first flavor numbering starts from 1, given start and end line are included.
If omitted the selection starts from the beginning or is taken to the end.

In the second flavor the content between two lines containing the pattern
`--- begin <key> ---` and `--- end <key> ---` is used.
Those patterns MUST occur exactly once.

This command can be used to embed some content of another file into the 
Markdown file, for example
parts of a Go file to provide some documentation consistent with actual
code like in the following example

<pre>
{{include}{../../../scan.go}{113:116}&rcub;
</pre>

which extracts the regular expressions used
to parse the annotations used by this tool.

```go
{{include}{../../../scan.go}{113:116}}
```

The second example uses the pattern syntax
to determine include content:

<pre>
{{include}{../../../cmds.go}{include args}&rcub;
</pre>

extracts the lines between the start and end pattern

```go
{{include}{../../../cmds.go}{example}}
```

which are the regexps used to parse the two argument flavors.

{{include}{../../../cmds.go}{filter}{(?m)^.*// ?(.*)$}}

The previous paragraph is taken from the source file
```go
// --- begin filter ---
{{include}{../../../cmds.go}{filter}}
// --- end filter ---
```
using

<pre>
{{include}{../../../cmds.go}{filter}{(?m)^.*// ?(.*)$}&rcub;
</pre>

{{cmd-execute}}
### Execute

The execute command uses the following syntax
<pre>
{{execute}{&lt;command>}{&lt;arg>}*}
{{execute}{&lt;command>}{&lt;arg>}*{&lt;extract>}[{[&lt;startline>][:[&lt;endline>]]}]}
{{execute}{&lt;command>}{&lt;arg>}*{&lt;extract>}{&lt;key>}&rcub;
</pre>

The command executes the given command with the given arguments and substitutes
the output. Every command argument is given as separate argument to the *mdref*
command. The directory of the file containing the command expression is used
as current working directory to resolve relative file names.

Optionally the special argument `{<extract>}` can be used to append 
some line selection arguments according to the [`include`](#include) command, line range as well as pattern selection and the optional additional filter argument.

The command in directory [`democmd`](../../democmd/main.go) outputs some
test content:

```
{{execute}{go}{run}{../../../democmd}{text}{some demo text}}
```

It can be completely substituted (as shown above) with the command

<pre>
{{execute}{go}{run}{../../../democmd}{text}{some demo text}&rcub;
</pre>

or you can select a dedicated line range with the filter pattern

<pre>
{{execute}{go}{run}{../../../democmd}{text}{some demo text}{&lt;extract>}{text}&rcub;
</pre>

which produces the following output:

```
{{execute}{go}{run}{../../../democmd}{text}{some demo text}{<extract>}{text}}
```

A selection by a number range (here just a single line) is possible, also:

<pre>
{{execute}{go}{run}{../../../democmd}{text}{some demo text}{&lt;extract>}{5}&rcub;
</pre>

substitutes line number 5

```
{{execute}{go}{run}{../../../democmd}{text}{some demo text}{<extract>}{5}}
```

