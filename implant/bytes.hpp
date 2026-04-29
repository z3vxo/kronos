#pragma once
#include "common.hpp"

class bytes {
public:
	INT index;
	INT size;
	PBYTE InData;
	PBYTE OutData;

	BOOL EnsureBuffer(PBYTE& Buffer, UINT size);
	void InitRead(PBYTE data, INT DataSize);
	UINT Read4();
	void ReadString(PBYTE Out, UINT Len);
	void InitWrite(PBYTE buffer, UINT len);
	void Write4(UINT val);
	void Write1(BOOL val);
	void WriteString(PBYTE Data, UINT Length);

};

extern bytes* g_ByteMgr;