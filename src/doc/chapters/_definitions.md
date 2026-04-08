Here we define some new pattern

<!--- begin pattern --->
{{pattern}{my-pattern}{(?m)^.*content: ?(.*)$}}
<!--- end pattern --->

and define some match for it:

<!--- begin content --->
Only the prefixed content will be matched by 
the pattern defined above:
content: this is some test content
content: provided by a definition file.
<!--- end content --->
	
{{variable}{root}{../../..}}
