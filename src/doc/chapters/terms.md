{{term:term}}
## Terms

Instead of just using an logical anchor as target for a markdown
reference, it may be used as text also. Therefore, it must
be enriched by some text used as *term*.

The term name is an alias for the given text. It can be used to consistently
use some terminology across the complete document
tree defined and changeable at a single place.

{{term-anchor:term anchor}}
### Term Anchors

Anchors for terms are just defined by using the [{{anchor}}] syntax adding a text separated by a colon ( `:`).

For example

<pre>
# My Chapter
{{term:text}&rcub;
</pre>

Like regular logical anchors, a term anchor might be placed anywhere in
a Markdown document, therefore, it is possible to link to any location, 
not only section headers.

Another possibility is to define the term text by some text substitution commands
like [{{cmd-include}}] or [{{cmd-execute}}] by preceding those commands 
by `{term}{<term>}`:

  <pre>
  {{term}{&lt;term>}{&lt;cmd>}...}&rcub;
  </pre>

For example the term declaration

{{term}{text}{include}{../../../scan.go}{symbol}{func *([a-zA-Z]+) *\(}}
<pre>
  {{term}{text}{include}{../../../scan.go}{symbol}{func *([a-zA-Z]+) *\(}&rcub;
</pre>

imports the term value "{{{text}}}" from some source file. It is then used by <code>{{{text}}&rcub;</code> to refer to the value in any document of the document tree.

Optionally, formatting options man be given:
- <code>&grave;</code>: format as code
- `*`: format in bold
- `_`: format in italic

For example 

{{term}{`symbol}{include}{../../../scan.go}{symbol}{func *([a-zA-Z]+) *\(}}
<pre>
  {{term}{`symbol}{include}{../../../scan.go}{symbol}{func *([a-zA-Z]+) *\(}&rcub;
</pre>

defines some code symbol {{{symbol}}}, which is formatted as code.

### Term Usage

Term anchors may be used as usual for regular references, also.
But there are two additional use cases:
- Usage as link labeled with the term text.
  In this case the [logical reference]({{reference}}) appears inside the `[]` pair
  of a markdown reference, but without the target part, which is automatically inserted from the term anchor:

  <pre>
  A [{{&lt;name>}}] is a ....
  </pre>

  Such a reference is called *term reference*. It refers to an
  anchor providing an additional text. This term reference is substituted by a complete link for the term text and the term reference.

- Just as text without any hyperlink. In this case the [logical reference]({{reference}}) appears inside a `{}` pair:

  <pre>
  A {{{&lt;name>}}} is a ....
  </pre>

{{term-flavors}}
### Term Substitution Flavors

There are several flavors for using a term:
- The term text is used as given
- The term text is converted into its plural form. Here, the name is preceded by an asterisk (`*`). For example <code>[{{*term}&rcub;]</code>.
- The term text should be capitalized, for example if used as first word in a sentence. Here the first letter of the name is taken in upper case, for example <code>[{{Term}&rcub;]</code>.
- If the plural form as well as upper case should be substituted, both mechanisms can be combined, like in <code>[{{*Term}&rcub;]</code>

{{term-pattern:Term Extraction Pattern}}
### Term Extraction Pattern

The extraction filter for [{{cmd-include}}] and [{{cmd-execute}}] are using 
regular expressions. This has been used to extract term text assignments directly from
the source base.

To simplify this, *mdref* supports some standard patterns, which can be specified by name instead of a regular expression.

{{term}{`go-const}{include}{../../../patterns.go}{go-const}{go-const}{$(1/_/-)}}
{{term}{`go-const-value}{include}{../../../patterns.go}{go-const}{go-const-value}}
- {{{go-const}}}: the name of a Go constant ({{{go-const-value}}})
{{term}{`go-const--value}{include}{../../../patterns.go}{go-const-value}{go-const}{$(1/_/-)}}
{{term}{`go-const--value-value}{include}{../../../patterns.go}{go-const-value}{go-const-value}}
- {{{go-const--value}}}: the value of a Go constant ({{{go-const--value-value}}})
{{term}{`go-var}{include}{../../../patterns.go}{go-var}{go-const}{$(1/_/-)}}
{{term}{`go-var-value}{include}{../../../patterns.go}{go-var}{go-const-value}}
- {{{go-var}}}: the name of a Go variable ({{{go-var-value}}})
{{term}{`go-type}{include}{../../../patterns.go}{go-type}{go-const}{$(1/_/-)}}
{{term}{`go-type-value}{include}{../../../patterns.go}{go-type}{go-const-value}}
- {{{go-type}}}: the name of a Go type ({{{go-type-value}}})
{{term}{`go-func}{include}{../../../patterns.go}{go-func}{go-const}{$(1/_/-)}}
{{term}{`go-func-value}{include}{../../../patterns.go}{go-func}{go-const-value}}
- {{{go-func}}}: the name of a Go function ({{{go-func-value}}})

### Other Patterns

There are also other useful patterns, like:

<!--- begin other patterns --->
{{term}{`go-line-comment}{include}{../../../patterns.go}{go-line-comment}{go-const}{$(1/_/-)}}
{{term}{`go-line-comment-value}{include}{../../../patterns.go}{go-line-comment}{go-const-value}}
- {{{go-line-comment}}}:  ({{{go-line-comment-value}}})
  {{include}{../../../patterns.go}{go-line-comment}{go-line-comment}}
{{term}{`go-comment}{include}{../../../patterns.go}{go-comment}{go-const}{$(1/_/-)}}
{{term}{`go-comment-value}{include}{../../../patterns.go}{go-comment}{go-const-value}}
- {{{go-comment}}}:  ({{{go-comment-value}}})
  {{include}{../../../patterns.go}{go-comment}{go-comment}}
{{term}{`html-comment}{include}{../../../patterns.go}{html-comment}{go-const}{$(1/_/-)}}
{{term}{`html-comment-value}{include}{../../../patterns.go}{html-comment}{go-const-value}}
- {{{html-comment}}}:  ({{{html-comment-value}}})
  {{include}{../../../patterns.go}{html-comment}{html-comment}}

This complete list is generated by 
<pre>
{{include}{terms.md}{other patterns}}
</pre>
<!--- end other patterns --->

surrounded by the key selector for `other patterns`.

{{cmd-pattern}}
### Pattern Definition

With the `pattern` command patterns can be created as part of the document tree.

For example
```text
{{include}{_definitions.md}{pattern}}
```
defines a new pattern. It is used here with

```
{{include}{_definitions.md}{content}{my-pattern}&rcub;
```

to provide the following result:

```
{{include}{_definitions.md}{content}{my-pattern}}
```

The pattern and the content are defined in a file `_definitions.md`, which is evaluated but not
generated into the target folder.

```
{{include}{_definitions.md}}
```