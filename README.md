# Events

Lightweight generic events library for go

## Installation

```bash
go get github.com/kmalaver/events
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
