
## Reference Targets

Any reference in a Markdown document may
use the annotated reference syntax to refer to global logical anchors.


<pre>
{{&lt;name>}}
</pre>

where name follows the [name syntax](../../README.md#name) for
anchors.

A typical link, like the one above, looks like this:

<pre>
[name syntax]({{name}&rcub;)
</pre>


### Anchors

Anchors are just defined by using the reference syntax outside the Markdown reference syntax.

For example

<pre>
{{my-chapter}&rcub;
# My Chapter
</pre>

It might be placed anywhere in a markdown document,
therefore, it is possible to link to any location, not only section headers.