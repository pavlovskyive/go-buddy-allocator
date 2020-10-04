# Golang Buddy Allocator

The buddy memory allocation technique is a memory allocation algorithm that divides memory into partitions to try to satisfy a memory request as suitably as possible. This system makes use of splitting memory into halves to try to give a best fit.

## Example Usage

```go
package main

import (
    "log"

    "github.com/pavlovskyive/go-buddy-allocator"
)

func main() {

    // Initalizing allocator.
    a := buddy.NewAllocator(1024)

    // Allocating block with size 16.
    // If there is no appropriate memory for block, allocator will return nil.
    x := a.Alloc(16)

    // Print info about free and allocated blocks.
    a.Log()

    // Reallocating block with size 16 to size 128.
    x = a.Realloc(x, 128)

    // Deallocating block with size 16.
    a.Free(x)
}
```

## Algorithm Description

If we start with a 512 bytes block allocated from the OS, we can split it to create two 256 bytes  buddies.
We can then take one of those and split it further into two 128 bytes buddies, and so on.

When allocating, we check to see if we have a free block of the appropriate size.
If not, we split a larger block as many times as necessary to get a block of a suitable size.
So if we want 32 bytes, we split the 128 bytes block into 64 bytes and then split one of those into 32 bytes.
It will look something like this:

<img src="https://i.stack.imgur.com/i4NNV.png" alt="buddy allocator visual demonstration" width=400 />

Deallocation block is simple: we just need to check is buddy block is free and combine them if it is. Adress of buddy block can be computed using binary XOR of adress of original block and size of this block.

Reallocation is basically deallocating block, allocating new one with needed size, and copying data from old block to new one.

---

Max size of Allocator (excluding memory) is **48 bytes on 64-bit** systems which is great because we don't need to store any metadata inside memory blocks.

*Note:* there is vet warning about possible misuse of unsafe.Pointer on line 138 because go-vet don't recognise binary XOR as arithmetic operation on pointer (but this doesn't seem to be fixed in near future at all).