#include "cmds.hpp"
#include <stdio.h>


enum CMD_ACTION_CODES {
	CMD_CODE_PS,
	CMD_CODE_CMD,
	CMD_CODE_CAT,
	CMD_CODE_LS,
	CMD_CODE_RM,
	CMD_CODE_MV,
};



BOOL commanders::Dispatch(PBYTE Data, UINT size, PBYTE OutBuffer) {
	g_ByteMgr->InitRead(Data, size);

	UINT TotalCommands = g_ByteMgr->Read4();
	for (INT i = 0; i < TotalCommands; i++) {
		UINT CmdCode = g_ByteMgr->Read4();
		switch (CmdCode)
		{
		case CMD_CODE_MV:
			do_mv();
			break;
		default:
			break;
		}

	}

	return TRUE;
}

void commanders::do_getprivs() {
	return;
}

void commanders::do_pwd() {
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

void commanders::do_mv() {
	return;
}




commanders* g_Commander = NULL;