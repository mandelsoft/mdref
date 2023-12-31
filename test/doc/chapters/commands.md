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
{{include}{../../../scan.go}{109:112}&rcub;
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

```go
// --- begin include args ---
var includeExpNum = regexp.MustCompile("^{([^}]+)}(?:{([0-9]+)?(?:(:)([0-9]+)?)?}(?:{(.*)})?)?$")
var includeExpPat = regexp.MustCompile("^{([^}]+)}{([a-zA-Z -]+)}(?:{(.*)})?$")

// --- end include args ---
```

which are the regexps used to parse the two argument flavors.

An optional third argument can be used to specify a filter regular
expression. It must contain one matching group. The
selected file range is matched by this regular expression and
the content of the first matching group of the all matches is
concatenated. If the expression uses the multi-line mode, the matches
are suffixed with a newline.


The previous paragraph is taken from the source file
```go
// --- begin filter ---
// An optional third argument can be used to specify a filter regular
// expression. It must contain one matching group. The
// selected file range is matched by this regular expression and
// the content of the first matching group of the all matches is
// concatenated. If the expression uses the multi-line mode, the matches
// are suffixed with a newline.
// --- end filter ---
```
using

<pre>
{{include}{../../../cmds.go}{filter}{(?m)^.*// ?(.*)$}&rcub;
</pre>

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
this is a demo file
// --- begin text ---
some demo text
// --- end text ---
this is line 5 of the demo output
and some other text.

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
some demo text
```

A selection by a number range (here just a single line) is possible, also:

<pre>
{{execute}{go}{run}{../../../democmd}{text}{some demo text}{&lt;extract>}{5}&rcub;
</pre>

substitutes line number 5

```
this is line 5 of the demo output
```

