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
```
{{include}{<filepath>}[{[<startline>][:[<endline>]]}]}
```

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
