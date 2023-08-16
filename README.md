# go-bundler-client

Golang client for ERC-4337-spec bundlers.
Uses types from [Stackup's bundler](https://github.com/stackup-wallet/stackup-bundler).

### Example

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mdehoog/go-bundler-client"
)

func main() {
	ctx := context.Background()
	c, err := bundler_client.DialContext(ctx, "http://localhost:4337")
	if err != nil {
		log.Fatalf("Failed to connect to bundler: %v", err)
	}

	chainId, err := c.ChainId(ctx)
	if err != nil {
		log.Fatalf("Failed to retrieve chainId from bundler: %v", err)
	}
	fmt.Printf("ChainId: %d\n", chainId.Uint64())
}

```
