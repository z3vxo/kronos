#include "bytes.hpp"
#include <stdio.h>

BOOL bytes::EnsureBuffer(PBYTE& Buffer, UINT datasize) {
	UINT NewSize = this->size;
	if (this->WriteIndex + datasize > this->size) {
		this->size = NewSize;
		while (NewSize < this->WriteIndex + datasize) {
			NewSize *= 2;
		}

		PBYTE Temp = (PBYTE)HeapReAlloc(GetProcessHeap(), HEAP_ZERO_MEMORY, Buffer, NewSize);
		Buffer = Temp;
		this->size = NewSize;

		return TRUE;
	}
	return TRUE;
}

void bytes::InitRead(PBYTE data, INT DataSize) {

	this->ReadIndex = 0;
	this->InData = data;
	this->size = DataSize;
}




UINT bytes::Read4() {
	UINT val;
	memcpy(&val, this->InData + this->ReadIndex, sizeof(val));
	this->ReadIndex += 4;
	return val;
}

void bytes::ReadString(PBYTE buffer, UINT len) {
	memcpy(buffer, this->InData + this->ReadIndex, len);
	this->ReadIndex += len;
}


void bytes::InitWrite() {
	if (this->OutData) { HeapFree(GetProcessHeap(), 0, this->OutData); }
	this->OutData = AllocMemory<BYTE>(BASE_BUFFER_SIZE);
	this->WriteIndex = 0;
	this->size = BASE_BUFFER_SIZE;
}

void bytes::Write4(UINT val) {
	this->EnsureBuffer(this->OutData, sizeof(val));
	memcpy(this->OutData + this->WriteIndex, &val, sizeof(val));
	this->WriteIndex += 4;

}

void bytes::Write1(BOOL val) {

	this->EnsureBuffer(this->OutData, 1);
	memcpy(this->OutData + this->WriteIndex, &val, 1);
	this->WriteIndex += 1;
}

void bytes::WriteString(PBYTE data, UINT len) {
	this->EnsureBuffer(this->OutData, len);
	memcpy(this->OutData + this->WriteIndex, data, len);
	this->WriteIndex += len;

}


bytes* g_ByteMgr = NULL;