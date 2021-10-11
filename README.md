# resource-ticker

This module wraps RAM and CPU resource information gathering.

Cgroups and cgroups2 are supported. If none of them is active, fallback to procfs provides resource information.


## How to use

```go
package main

import (
    "log"
    "github.com/arivum/resource-ticker/pkg/resources"
)

func main() {
    if ticker, err := resources.NewResourceTicker(resources.WithCPUFloatingAvg(1)); err != nil {
        log.Fatal(err)
    }

    resourceChan, errChan := ticker.Run()

    for {
		select {
		case r := <-resourceChan:
			log.Printf("$+v\n", r.RAM)
			log.Printf("$+v\n", r.CPU)
		case err := <-errChan:
			log.Println(err)
		}
	}
}
```
