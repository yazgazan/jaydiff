# JayDiff

[![Go Report Card](https://goreportcard.com/badge/github.com/yazgazan/jaydiff)](https://goreportcard.com/report/github.com/yazgazan/jaydiff)
[![GoDoc](https://godoc.org/github.com/yazgazan/jaydiff?status.svg)](https://godoc.org/github.com/yazgazan/jaydiff)
[![Build Status](https://travis-ci.org/yazgazan/jaydiff.svg?branch=master)](https://travis-ci.org/yazgazan/jaydiff)
[![Coverage Status](https://coveralls.io/repos/github/yazgazan/jaydiff/badge.svg?branch=master)](https://coveralls.io/github/yazgazan/jaydiff?branch=master)

# Install

## Downloading the compiled binary

- Download the latest version of the binary: [releases](https://github.com/yazgazan/jaydiff/releases)
- extract the archive and place the `jaydiff` binary in your `$PATH`

## From source

- Have go 1.8 or greater installed: [golang.org](https://golang.org/doc/install)
- run `go get -u github.com/yazgazan/jaydiff`

# Usage

```
Usage:
  jaydiff [OPTIONS] FILE_1 FILE_2

Application Options:
  -i, --ignore=     paths to ignore (glob)
      --indent=     indent string (default: "\t")
  -t, --show-types  show types
  -r, --report      output report format

Help Options:
  -h, --help        Show this help message
```

## Examples

Getting a full diff of two json files:

```diff
$ jaydiff --show-types old.json new.json

 map[string]interface {} map[
     a: float64 42
     b: []interface {} [
         float64 1
-        float64 3
+        float64 5
+        float64 4
     ]
     c: map[string]interface {} map[
-        a: string toto
+        a: string titi
-        b: float64 23
+        b: string 23
     ]
-    e: []interface {} []
-    f: float64 42
     g: []interface {} [1 2 3]
+    h: float64 42
 ]
```

Ignoring fields:

```diff
$ jaydiff --show-types \
	  --ignore='.b\[\]' --ignore='.d' --ignore='.c.[ac]' \
	    old.json new.json

 map[string]interface {} map[
     a: float64 42
     b: []interface {} [1 3]
     c: map[string]interface {} map[
-        b: float64 23
+        b: string 23
     ]
-    e: []interface {} []
-    f: float64 42
     g: []interface {} [1 2 3]
+    h: float64 42
 ]
```

Report format:

```diff
$ jaydiff --report --show-types old.json new.json

- .b[]: float64 3
+ .b[]: float64 5
+ .b[]: float64 4
- .c.a: string toto
+ .c.a: string titi
- .c.b: float64 23
+ .c.b: string 23
- .e: []interface {} []
- .f: float64 42
+ .h: float64 42
```

Ignore Excess values (useful when checking for backward compatibility):

```diff
$ jaydiff --report --show-types --ignore-excess old.json new.json

- .b[]: float64 3
+ .b[]: float64 5
- .c.a: string toto
+ .c.a: string titi
- .c.b: float64 23
+ .c.b: string 23
- .e: []interface {} []
- .f: float64 42
```

# Ideas

- Handle cyclic maps/structures properly
- JayPatch
- Have the diff lib support more types (Structs, interfaces (?), Arrays, ...)
- JSON-ish output (instead of go-ish)
