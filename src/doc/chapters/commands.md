## Commands
{{commands:command}}

Commands provide additional functionality to Markdown files.
All commands use the same basic annotation syntax:

```
{{<name>}{<arg>}...}
```

The following commands are supported:

### Include

The include command uses the following syntax
<pre>
{{include}{&lt;filepath>}[{[&lt;startline>][:[&lt;endline>]]}]}
{{include}{&lt;filepath>}{<key>}&rcub;
</pre>


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
{{include}{../../../scan.go}{105:108}&rcub;
</pre>

which extracts the regular expressions used
to parse the annotations used by this tool.

```go
{{include}{../../../scan.go}{105:108}}
```

The second example uses the pattern syntax
to determine include content:

<pre>
{{include}{../../../cmds.go}{include args}&rcub;
</pre>

extracts the lines between the start and end pattern

```
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

