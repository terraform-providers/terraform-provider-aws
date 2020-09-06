# gogrep

	go get mvdan.cc/gogrep

Search for Go code using syntax trees. Work in progress.

	gogrep -x 'if $x != nil { return $x, $*_ }'

### Instructions

	usage: gogrep commands [packages]

A command is of the form "-A pattern", where -A is one of:

       -x  find all nodes matching a pattern
       -g  discard nodes not matching a pattern
       -v  discard nodes matching a pattern
       -a  filter nodes by certain attributes
       -s  substitute with a given syntax tree
       -w  write source back to disk or stdout

A pattern is a piece of Go code which may include wildcards. It can be:

       a statement (many if split by semicolonss)
       an expression (many if split by commas)
       a type expression
       a top-level declaration (var, func, const)
       an entire file

Wildcards consist of `$` and a name. All wildcards with the same name
within an expression must match the same node, excluding "_". Example:

       $x.$_ = $x // assignment of self to a field in self

If `*` is before the name, it will match any number of nodes. Example:

       fmt.Fprintf(os.Stdout, $*_) // all Fprintfs on stdout

`*` can also be used to match optional nodes, like:

	for $*_ { $*_ }    // will match all for loops
	if $*_; $b { $*_ } // will match all ifs with condition $b

Regexes can also be used to match certain identifier names only. The
`.*` pattern can be used to match all identifiers. Example:

       fmt.$(_ /Fprint.*/)(os.Stdout, $*_) // all Fprint* on stdout

The nodes resulting from applying the commands will be printed line by
line to standard output.

Here are two simple examples of the -a operand:

       gogrep -x '$x + $y'                   // will match both numerical and string "+" operations
       gogrep -x '$x + $y' -a 'type(string)' // matches only string concatenations
