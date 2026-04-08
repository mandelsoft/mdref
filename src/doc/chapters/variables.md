{{cmd-variable:variable}}
## Variables

It is possible to predefine variable values usable in
other commands in file path arguments (e.g. [{{cmd-include}}]) or execution argument for [{{cmd-execute}}]. They are defined by the special command `variable` which produces
no substitution.

<pre>
{{variable}{&lt;name>}{&lt;value>}&rcub;
</pre>

Variable values are used by a substitution expression of the form `$(<name>)` inside
appropriate command arguments. Variable definitions may again use variables, but there must be no cycle.

e.g.:
<pre>
{{include}{commands.md}{variable}}
</pre>
