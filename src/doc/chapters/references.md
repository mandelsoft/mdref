{{reference:reference}}
## References

Any reference in a Markdown document may
use the annotated reference syntax to refer to a global logical anchors.


<pre>
{{&lt;name>}}
</pre>

where name follows the [name syntax]({{name}}) for
anchors.

A typical link, like the one above, looks like this:

<pre>
[name syntax]({{name}&rcub;)
</pre>


{{anchor:anchor}}
### Anchors

Anchors are just defined by using the reference syntax outside the Markdown reference syntax.

For example

<pre>
{{my-chapter}&rcub;
# My Chapter
</pre>

It might be placed anywhere in a markdown document,
therefore, it is possible to link to any location, not only section headers.

If placed directly before or after a Markdown section heading
(and the option `--headings`) is given, no explicit anchor is generated,
but the Markdown anchor name is used for the references.
This keeps the generated file viewable by Markdown viewers not able to
work with anchors. 