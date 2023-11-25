## Commands
{{commands}}

Commands provide additional functionality to Markdown files.
All commands use the same basic annotation syntax:

```
{{<name>}{<arg>}...}
```

The following commands are supported:

### Include

The include command uses the following syntax
<pre>
{{include}{<filepath>}[{[<startline>][:[<endline>]]}]}
{{include}{<filepath>}{<key>}&rcub;
</pre>

In the first flavor numberimg starts from 1, given start and end line are included.
If omitted the selection starts from the beginng or is taken to the end.

In the second flavotr the content between two lines containing the pattern
`--- begin <key> ---` and `--- end <key> ---` is used.
Those pattern MUST occur exactly once.

This way it is possible to include some
content from another file, for example
a Go file like in the following example

<pre>
{{include}{../../../scan.go}{90:93}&rcub;
</pre>

which extracts the regular expressions used
to parse the annotations used by this tool.

```go
{{include}{../../../scan.go}{90:93}}
```

<pre>
{{include}{../../../cmds.go}{include args}&rcub;
</pre>

extracts the lines between the start and end pattern

```
{{include}{../../../cmds.go}{example}}
```

which are the regexps used to parse the two argument flavors.
