#include "bytes.hpp"


// fix this later
//BOOL bytes::EnsureBuffer(UINT datasize) {
//	if (this->index + datasize > this->size) {
//		while (this->size < this->index + datasize) {
//			if (this->InData) {
//				PBYTE temp = (PBYTE)HeapReAlloc(GetProcessHeap(), HEAP_ZERO_MEMORY, this->InData, this->size * 2);
//				if (temp == NULL) { return FALSE; };
//				this->InData = temp;
//			}
//			if (this->OutData) {
//				PBYTE temp = (PBYTE)HeapReAlloc(GetProcessHeap(), HEAP_ZERO_MEMORY, this->OutData, this->size * 2);
//				if (temp == NULL) { return FALSE; }
//				this->OutData = temp;
//			}
//		}
//	}
//}

void bytes::InitRead(PBYTE data, INT DataSize) {
	this->index = 0;
	this->InData = data;
	this->size = DataSize;
}



// TODO, add bounds checking to all below
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
	memcpy(this->OutData + this->index, &val, sizeof(val));
	this->index += 4;
}

void bytes::WriteString(PBYTE data, UINT len) {
	memcpy(this->OutData + this->index, data, len);
	this->index += len;
}


bytes* g_ByteMgr = NULL;