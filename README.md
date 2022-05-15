# An embeddable Prolog

[golog](https://github.com/mndrix/golog) is a Prolog interpreter in Go.

This package extending the golog and making it a **pragmatic embedding** language.
With some helper functions provided by `prolog`, calling Prolog from Golang is very simple.

### Usage

The package is fully go-getable, so, just type

  `go get github.com/rosbit/prolog`

to install.

#### 1. Instantiate a Prolog interpreter

```go
package main

import (
  "github.com/rosbit/prolog"
  "fmt"
)

func main() {
  ctx := pl.NewProlog()
  ...
}
```

#### 2. Load a Prolog script

Suppose there's a Prolog file named `music.pl` like this:

```prolog
listen(ergou, bach).
listen(ergou, beethoven).
listen(ergou, mozart).
listen(xiaohong, mj).
listen(xiaohong, dylan).
listen(xiaohong, bach).
listen(xiaohong, beethoven).
```

one can load the script like this:

```go
   if err := ctx.LoadFile("music.pl"); err != nil {
      // error processing
   }
```

#### 3. Prepare arguments and variables

```go
   // query Who listens Music
   args := []interface{}{pl.PlVar("Who"), pl.PlVar("Music")}

   // query Who listens "bach"
   args := []interface{}{pl.PlVar("Who"), "bach"}

   // query Which Music "ergou" listens
   args := []interface{}{"ergou", pl.PlVar("Music")}

   // check whether "ergou" listens "bach"
   args := []interface{}{"ergou", "bach"}
```

#### 4. Query the goal with arguments and variables

```go
   rs, ok, err := ctx.Query("listen", args...)
```

#### 5. Check the result

```go
   // error checking
   if err != nil {
      // error processing
      return
   }

   // proving checking with result `false`
   if !ok {
      // the result is false
      return
   }

   // proving checking with result `true`
   if rs == nil {
      // the result is true
      return
   }

   // result set processing
   for res := range rs {
      fmt.Printf("res: %#v\n", res)
   }
```

The full usage sample can be found [sample/main.go](https://github.com/rosbit/prolog/blob/master/sample/main.go).

### Status

The package is not fully tested, so be careful.

### Contribution

Pull requests are welcome! Also, if you want to discuss something send a pull request with proposal and changes.
__Convention:__ fork the repository and make changes on your fork in a feature branch.
