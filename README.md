# Go Persistent Vector Implementation

This repository contains a Go implementation of a persistent vector (also known as a persistent array or immutable vector). Persistent data structures are immutable; instead of modifying the existing data structure, operations return a *new* version with the changes applied, while the original remains unchanged. This offers advantages in concurrency, versioning, and reasoning about program state.

## Features ✨

Here's what's currently implemented:

## Vectors:

| Feature        | Status |
|----------------|--------|
| Append         | ✅     |
| Set            | ✅     |
| Immutability   | ✅     |
| Persistence    | ✅     |
| Generic Types  | ✅     |
| Deletion       | 🚧     |
| Iteration      | 🚧     |

## Usage

```go
package main

import (
	"fmt"
	"go-persistent-tree"
)

func main() {
	// Create a new persistent vector with initial values and a node width of 2^5 (32)
	initialValues := []int{1, 2, 3, 4, 5}
	vec := persistent_tree.NewPersistentVec(initialValues, 5) // power = 5, width = 32

	fmt.Println("Initial Vector:", vec.ToGenericVec()) // Output: [1 2 3 4 5]

	// Append more values
	vec2 := vec.Append(6, 7, 8)
	fmt.Println("Vector after append:", vec2.ToGenericVec()) // Output: [1 2 3 4 5 6 7 8]
	fmt.Println("Original Vector:", vec.ToGenericVec()) // Output: [1 2 3 4 5] (unchanged)

	// Set value on specific index
	vec3 := vec2.Set(0, 100)
	fmt.Println("Vector after set:", vec3.ToGenericVec()) // Output: [100 2 3 4 5 6 7 8]
	fmt.Println("Vector before set:", vec2.ToGenericVec()) // Output: [1 2 3 4 5 6 7 8] (unchanged)
}
```

## Notes

*   The choice of `power` (and therefore `width`) can impact performance. Smaller values of `power` result in a deeper tree, while larger values result in a wider tree. Experiment to find the best value for your use case. A value of 5 (width=32) is a reasonable starting point.
*   The `String()` method is primarily for debugging and can be verbose for large vectors.
*   This implementation prioritizes immutability and persistence. Performance may not be as high as mutable arrays for some operations.

## References

*   [Go implementation of persistent data structures](https://github.com/tobgu/peds)
*   [Clojure implementation of a persistent vector written in Java](https://github.com/clojure/clojure/blob/0b73494c3c855e54b1da591eeb687f24f608f346/src/jvm/clojure/lang/PersistentVector.java)
*   [Series of articles with a detailed breakdown of persistent data structures](https://hypirion.com/musings/understanding-persistent-vector-pt-1)