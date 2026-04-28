#pragma once

#include <windows.h>
#include "apidefs.hpp"
#include "nt.hpp"


template<typename T>
T* AllocMemory(SIZE_T size) {
	return (T*)HeapAlloc(GetProcessHeap(), HEAP_ZERO_MEMORY, size);
}

inline PTEB GetTeb() { return (PTEB)__readgsqword(0x30); }
inline PPEB GetPEB() { return (PPEB)GetTeb()->ProcessEnvironmentBlock; }

BOOL InitAgent();