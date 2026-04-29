#include "bytes.hpp"


BOOL bytes::EnsureBuffer(PBYTE& Buffer, UINT datasize) {
	UINT NewSize = this->size;
	if (this->index + datasize > this->size) {
		while (this->size < this->index + datasize) {
			NewSize *= 2;
		}

		PBYTE Temp = (PBYTE)HeapReAlloc(GetProcessHeap(), HEAP_ZERO_MEMORY, Buffer, NewSize);
		Buffer = Temp;
		return TRUE;
	}
	return TRUE;
}

void bytes::InitRead(PBYTE data, INT DataSize) {
	this->index = 0;
	this->InData = data;
	this->size = DataSize;
}




UINT bytes::Read4() {
	UINT val;
	memcpy(&val, this->InData + this->index, sizeof(val));
	this->index += 4;
	return val;
}

void bytes::ReadString(PBYTE buffer, UINT len) {
	memcpy(buffer, this->InData + this->index, len);
	this->index += len;
}


void bytes::InitWrite(PBYTE buffer, UINT DataSize) {
	this->index = 0;
	this->OutData = buffer;
	this->size = DataSize;
}

void bytes::Write4(UINT val) {
	this->EnsureBuffer(this->OutData, sizeof(val));
	memcpy(this->OutData + this->index, &val, sizeof(val));
	this->index += 4;
}

void bytes::Write1(BOOL val) {
	this->EnsureBuffer(this->OutData, 1);
	memcpy(this->OutData + this->index, &val, 1);
	this->index += 1;
}

void bytes::WriteString(PBYTE data, UINT len) {
	this->EnsureBuffer(this->OutData, len);
	memcpy(this->OutData + this->index, data, len);
	this->index += len;
}


bytes* g_ByteMgr = NULL;