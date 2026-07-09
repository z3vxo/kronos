#pragma once
#include <windows.h>

DWORD Hasher(PCHAR str);
BOOL LoadAPIS();
FARPROC GetProc(HMODULE dll, DWORD hash);
HMODULE GetModule(DWORD Hash);
