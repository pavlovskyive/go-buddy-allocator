package buddy

import (
	"log"
	"unsafe"
)

//#include <string.h>
import "C"

// minPowOfTwo returns ceil power of two
func minPowOfTwo(number int) int {
	i := 1
	for number > i {
		i *= 2
	}
	return i
}

// minLog2 returns ceil log2(number)
func minLog2(number int) int {
	i := 0
	for 1<<i < number {
		i++
	}
	return i
}

// BlockQueue is queue type structure which stores pointers to free blocks
type BlockQueue struct {
	array []unsafe.Pointer
}

// Append adds element to queue
func (q *BlockQueue) Append(p unsafe.Pointer) {
	q.array = append(q.array, p)
}

// Dequeue removes first element of queue and returns it
func (q *BlockQueue) Dequeue() unsafe.Pointer {
	p := q.array[0]
	q.array = q.array[1:]

	return p
}

// RemoveAt removes element with given index from list
func (q *BlockQueue) RemoveAt(index int) {
	q.array = append(q.array[:index], q.array[index+1:]...)
}

// Length returns number of elements in queue
func (q *BlockQueue) Length() int {
	return len(q.array)
}

// Allocator handles memory allocation
type Allocator struct {
	// size of allocated memory
	size int
	// computed number of levels in allocator
	maxDepth int
	// lists of free elements
	freeQueues []BlockQueue
	// map that stores level of block with certain pointer
	allocatedBlocks map[unsafe.Pointer]int
}

// LevelOfSize returns possible depth of block with given size
func (a *Allocator) LevelOfSize(size int) int {
	return minLog2(a.size / size)
}

// SizeOfLevel returns size of blocks on given level
func (a *Allocator) SizeOfLevel(level int) int {
	return a.size / (1 << level)
}

// Alloc allocates block of given size and returns pointer to this block
func (a *Allocator) Alloc(size int) unsafe.Pointer {
	alignedSize := minPowOfTwo(size)
	level := a.LevelOfSize(alignedSize)

	pointer := a.FindFreeBlockOnLevel(level)

	if pointer == nil {
		log.Println("Cannot allocate memory block with size", size)
		return nil
	}

	// store level of block. Needed for freeing and reallocating this block later
	a.allocatedBlocks[pointer] = level
	return unsafe.Pointer(pointer)
}

// Free deallocates memory block
func (a *Allocator) Free(pointer unsafe.Pointer) {
	level := a.allocatedBlocks[pointer]
	delete(a.allocatedBlocks, pointer)

	buddyPointer := a.FindBuddy(pointer, level)

	a.freeQueues[level].Append(pointer)

	// combining blocks if buddy block is free
	for i, p := range a.freeQueues[level].array {
		if p == buddyPointer {
			a.freeQueues[level].RemoveAt(a.freeQueues[level].Length() - 1)
			a.freeQueues[level].RemoveAt(i)
			a.allocatedBlocks[pointer] = level - 1
			a.Free(pointer)
		}
	}
}

// FindFreeBlockOnLevel searches for free blocks on given level
func (a *Allocator) FindFreeBlockOnLevel(level int) unsafe.Pointer {
	if level < 0 {
		return nil
	}

	// split higher level block if no free block of given size
	if len(a.freeQueues[level].array) == 0 {
		higherLevelBlockPointer := a.FindFreeBlockOnLevel(level - 1)
		if higherLevelBlockPointer == nil {
			return nil
		}
		a.freeQueues[level].Append(higherLevelBlockPointer)
		a.freeQueues[level].Append(a.FindBuddy(higherLevelBlockPointer, level))
	}

	b := a.freeQueues[level].Dequeue()
	return b
}

// FindBuddy returns pointer to buddy of given block
func (a *Allocator) FindBuddy(pointer unsafe.Pointer, level int) unsafe.Pointer {
	return unsafe.Pointer(uintptr(pointer) ^ uintptr(a.SizeOfLevel(level)))
}

// Realloc changes size of given block
func (a *Allocator) Realloc(pointer unsafe.Pointer, size int) unsafe.Pointer {
	alignedSize := minPowOfTwo(size)
	oldSize := a.SizeOfLevel(a.allocatedBlocks[pointer])
	level := a.LevelOfSize(alignedSize)

	if a.allocatedBlocks[pointer] == level {
		return pointer
	}

	a.Free(pointer)
	newPointer := a.Alloc(alignedSize)

	// copy data from old pointer to new one
	if alignedSize < oldSize {
		C.memcpy(pointer, newPointer, C.size_t(alignedSize))
	} else {
		C.memcpy(pointer, newPointer, C.size_t(oldSize))
	}

	return newPointer
}

// Log prints out all allocated and free blocks
func (a *Allocator) Log() {
	log.Println("Allocated memory blocks:")

	if len(a.allocatedBlocks) == 0 {
		log.Println("   * none")
	}
	for pointer, level := range a.allocatedBlocks {
		log.Println("   * Pointer:", pointer, ", size:", a.SizeOfLevel(level))
	}

	log.Println("Free memory blocks:")

	for level, queue := range a.freeQueues {
		for _, pointer := range queue.array {
			log.Println("   * Pointer", pointer, ", size:", a.SizeOfLevel(level))
		}
	}
}

// NewAllocator initializes an allocator
func NewAllocator(size int) *Allocator {

	a := &Allocator{}
	if size < int(unsafe.Sizeof(int(0))) {
		log.Fatal("Too small size")
	}

	alignedSize := minPowOfTwo(size)

	a.size = alignedSize
	a.maxDepth = minLog2(alignedSize / int(unsafe.Sizeof(int(0))))
	a.freeQueues = make([]BlockQueue, a.maxDepth+1)
	a.allocatedBlocks = make(map[unsafe.Pointer]int)

	mem := C.malloc(C.ulong(alignedSize))
	a.freeQueues[0].Append(unsafe.Pointer(&mem))

	return a
}
