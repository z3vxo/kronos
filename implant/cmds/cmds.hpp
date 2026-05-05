#pragma once
#include "../hades/hades.h"

#define MSG_TYPE_OUTPUT 2


class commanders {

	void do_mv();
	void do_getprivs();
	void do_pwd();
	void do_cat();
	void do_rm();
	void do_ls();

public:
	BOOL Dispatch(PBYTE Data, UINT size, PBYTE OutBuffer);
};

extern commanders* g_Commander;