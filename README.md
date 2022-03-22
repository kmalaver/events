# Events

Simple go type safe events library.

## Installation

```bash
go get github.com/go-events/events
```

## Usage

```go
package main

import "github.com/kmalaver/events"

func main() {
  // Create new event
  e := events.New[string]()

  // Subscribe to event
  e.Subscribe(func(s string) {
    fmt.Println(s)
  })

  // Dispatch event
  e.Dispatch("Hello world!")
}

```	
