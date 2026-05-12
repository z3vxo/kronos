#pragma once
#include "../hades/hades.h"

#define MSG_TYPE_OUTPUT 2
#define STATUS_OK 0
#define STATUS_ERROR 1
#define RESP_NO_DATA 1
#define RESP_HAS_DATA 0
#define TASK_TYPE_NO_PARSE 0

class bytes {
public:
	INT ReadIndex;
	INT WriteIndex;
	INT size;
	PBYTE InData;
	PBYTE OutData;

	UINT BeginTask();
	void EndOk(UINT SucessCode);
	void EndErr(UINT errCode);
	void EndOkData(UINT TaskType, UINT len, PBYTE data);

	BOOL EnsureBuffer(PBYTE& Buffer, UINT size);
	void InitRead(PBYTE data, INT DataSize);
	UINT Read4();
	void ReadInto(PBYTE dst, UINT len);
	PCHAR ReadString(UINT Len);
	void  FreeString(PCHAR s);
	void InitWrite();
	void Write4(UINT val);
	void Write8(ULONGLONG val);
	void Write1(BOOL val);
	void WriteString(PBYTE Data, UINT Length);


};

extern bytes* g_ByteMgr;