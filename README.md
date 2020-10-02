# Golang Buddy Allocator

Simple buddy memory allocator written in Golang for "Operating Systems-2" course.

The **buddy allocator** works by repeatedly splitting memory blocks in half to create two smaller "buddies" until we get a block of the desired size.

If we start with a 512 bytes block allocated from the OS, we can split it to create two 256 bytes  buddies.
We can then take one of those and split it further into two 128 bytes buddies, and so on.

When allocating, we check to see if we have a free block of the appropriate size.
If not, we split a larger block as many times as necessary to get a block of a suitable size.
So if we want 32 bytes, we split the 128 bytes block into 64 bytes and then split one of those into 32 bytes.
It will look something like this:

<img src="https://i.stack.imgur.com/i4NNV.png" alt="buddy allocator visual demonstration" width=400 />

Finding of free block can be implemented via:

- Lists of free blocks with fixed sizes of powers of two

- Bitfields where 1 means current leaf is allocated

- Tree struct

My choise was to use lists.

Max size of Allocator (excluding memory) is **48 bytes on 64-bit** systems which is great because we don't need to store any metadata inside memory blocks.
