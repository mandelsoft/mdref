## Terms

Instead of using a reference as target for a markdown
reference it may be used as complete markdown linked text.
In this case the [logical reference](references.md) appears inside the `[]` pair
of a markdown reference, but without the target part:

<pre>
A [{{&lt;name>}}] is a ....
</pre>

Such a reference is called *term reference*. It refers to an
anchor providing an additional text. This term reference is substituted by a complete link for the term text and the term reference.

### Term Substitution Flavors

There are several flavors for using a term:
- The term text is used as given
- The term text is converted into its plural form. Here, the name is preceded by an asterisk (`*`). For example <code>[{{*term}&rcub;]</code>.
- The term text should be capitalized, for example if used as first word in a sentence. Here the first letter of the name is taken in upper case, for example <code>[{{Term}&rcub;]</code>.
- If the plural form as well as upper case should be substituted, both mechanisms can be combined, like in <code>[{{*Term}&rcub;]</code>

### Anchors

Anchors for terms are just defined by using the reference syntax outside the
Markdown reference syntax and adding a text separated by a colon ( `:`).

For example

<pre>
# My Chapter
{{term:term}&rcub;
</pre>

It might be placed anywhere in a Markdown document,
therefore, it is possible to link to any location, not only section headers.

Term anchors may be used for regular references, also.