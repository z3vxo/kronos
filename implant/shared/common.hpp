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

inline PTEB GetTeb() { return (PTEB)__readgsqword(0x30); }
inline PPEB GetPEB() { return (PPEB)GetTeb()->ProcessEnvironmentBlock; }

BOOL InitAgent();

DWORD Hasher(PCHAR str);