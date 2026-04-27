#include <windows.h>


template<typename T>
T* AllocMemory(SIZE_T size) {
	return (T*)HeapAlloc(GetProcessHeap(), HEAP_ZERO_MEMORY, size);
}