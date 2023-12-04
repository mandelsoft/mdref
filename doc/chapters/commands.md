## Commands

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
var refExp = regexp.MustCompile(`\({{([a-z0-9.-]+)}}\)`)
var trmExp = regexp.MustCompile(`\[{{([*]?[A-Za-z][a-z0-9.-]*)}}\]`)
var tgtExp = regexp.MustCompile(`(?:^|[^([]){{([a-z][a-z0-9.-]*)(:([a-zA-Z][a-zA-Z0-9- ]+))?}}`)
var cmdExp = regexp.MustCompile(`{{([a-z]+)}((?:{[^}]+})+)}`)
```

The second example uses the pattern syntax
to determine include content:

<pre>
{{include}{../../../cmds.go}{include args}&rcub;
</pre>

extracts the lines between the start and end pattern

```
// --- begin include args ---
var includeExpNum = regexp.MustCompile("^{([^}]+)}(?:{([0-9]+)?(?::([0-9]+)?)?})?$")
var includeExpPat = regexp.MustCompile("^{([^}]+)}{([a-zA-Z -]+)}$")

// --- end include args ---
```

which are the regexps used to parse the two argument flavors.
