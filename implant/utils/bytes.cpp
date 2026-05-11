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




/*
  [task output count] 4 bytes
 // looped
 [TASKID] 4 BYTES
 [STATUS] 4 bytes -> if 0 == success read next, if 1 == read read4() for error code
 [TASK_TYPE] 4 BYTES -> parse this, jump to handler and parse it, if task type == 0, continue below, single string output no further parsing
 ------
 [HAS_DATA] 4 BYTES -> if > 1 lookup in success map else below
 [OUTPUT LEN] 4 BYTES
 [OUTPUT DATA] N BYTES
*/
UINT bytes::BeginTask() {
	UINT taskID = this->Read4();
	this->Write4(taskID);
	return taskID;
}

void bytes::EndOk(UINT SuccessCode) {
	this->Write4(STATUS_OK);
	this->Write4(TASK_TYPE_NO_PARSE);
	this->Write4(SuccessCode);
}

void bytes::EndErr(UINT errCode) {
	this->Write4(STATUS_ERROR);
	this->Write4(errCode);
}

void bytes::EndOkData(UINT TaskType, UINT len, PBYTE data) {
	this->Write4(STATUS_OK);
	this->Write4(TaskType);
	this->Write4(RESP_HAS_DATA);
	this->Write4(len);
	this->WriteString(data, len);
}


UINT bytes::Read4() {
	UINT val;
	memcpy(&val, this->InData + this->ReadIndex, sizeof(val));
	this->ReadIndex += 4;
	return val;
}

PCHAR bytes::ReadString(UINT len) {

	PCHAR buf = AllocMemory<CHAR>(len + 1);
	if (!buf) return NULL;

	PBYTE src = this->InData + this->ReadIndex;
	for (UINT i = 0; i < len; i++) {
		buf[i] = (CHAR)src[i];
	}
	buf[len] = '\0';

	this->ReadIndex += len;
	return buf;
}

void bytes::FreeString(PCHAR s) {
	if (s) HeapFree(GetProcessHeap(), 0, s);
}

void bytes::ReadInto(PBYTE dst, UINT len) {

	PBYTE src = this->InData + this->ReadIndex;
	for (UINT i = 0; i < len; i++) {
		dst[i] = src[i];
	}
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