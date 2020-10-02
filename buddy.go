// package main

// import (
// 	"fmt"

// 	"github.com/golang-collections/go-datastructures/bitarray"
// )

// // Allocator allocates memory
// type Allocator struct {
// 	bitArray bitarray.BitArray
// }

// // NewAllocator inits allocator
// func NewAllocator(size int) *Allocator {
// 	a := Allocator{}
// 	a.bitArray = bitarray.NewBitArray(uint64(size))
// 	return &a
// }

// func main() {
// 	a := NewAllocator(1024)
// 	fmt.Println(a.bitArray.Capacity())
// }

package main

import (
	"log"
	"unsafe"
)

func minPowOfTwo(number int) int {
	i := 1
	for number > i {
		i *= 2
	}
	return i
}

func minLog2(number int) int {
	i := 0
	for 1<<i < number {
		i++
	}
	return i
}

// Block stores pointer to block of memory
type Block struct {
	offset int
}

// BlockQueue stores free blocks
type BlockQueue struct {
	array []Block
}

// Append adds element to queue
func (q *BlockQueue) Append(b Block) {
	q.array = append(q.array, b)
}

// Dequeue removes first element of queue and returns it
func (q *BlockQueue) Dequeue() Block {
	b := q.array[0]
	q.array = q.array[1:]

	return b
}

// Allocator handles memory allocation
type Allocator struct {
	ptr        unsafe.Pointer
	size       int
	maxDepth   int
	freeQueues []BlockQueue
}

// Alloc allocates block of given size
func (a *Allocator) Alloc(size int) unsafe.Pointer {
	alignedSize := minPowOfTwo(size)

	level := minLog2(a.size / alignedSize)
	block := a.FindBlockOnLevel(level)

	if block == nil {
		log.Println("Cannot allocate memory block with size", size)
		return nil
	}
	return unsafe.Pointer(uintptr(a.ptr) + uintptr(block.offset))
}

// FindBlockOnLevel searches for free blocks on needed level
func (a *Allocator) FindBlockOnLevel(level int) *Block {
	if level < 0 {
		return nil
	}

	if len(a.freeQueues[level].array) == 0 {
		higherLevelBlock := a.FindBlockOnLevel(level - 1)
		if higherLevelBlock == nil {
			return nil
		}
		a.freeQueues[level].Append(Block{higherLevelBlock.offset})
		a.freeQueues[level].Append(Block{higherLevelBlock.offset + a.size/(1<<level)})
	}

	b := a.freeQueues[level].Dequeue()
	return &b
}

// NewAllocator creates instance of allocator
func NewAllocator(size int) *Allocator {

	a := &Allocator{}
	if size < 32 {
		log.Fatal("Too small size")
	}

	alignedSize := minPowOfTwo(size)

	a.size = alignedSize
	a.maxDepth = minLog2(alignedSize / int(unsafe.Sizeof(int(0))))
	a.freeQueues = make([]BlockQueue, a.maxDepth+1)
	a.freeQueues[0].Append(Block{})

	mem := make([]int, alignedSize/int(unsafe.Sizeof(int(0))))
	a.ptr = unsafe.Pointer(&mem)

	return a
}

func main() {
	a := NewAllocator(1024)

	log.Println("Size of allocator in worst scenario: ", unsafe.Sizeof(Allocator{}))
	log.Println("Pointer to start of allocated memory: ", a.ptr)

	log.Println()

	log.Println(a.freeQueues)

	log.Println()

	log.Println(a.Alloc(256))
	log.Println(a.freeQueues)

	log.Println()

	log.Println(a.Alloc(8))
	log.Println(a.freeQueues)
}
