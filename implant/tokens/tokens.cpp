#include "tokens.hpp"



BOOL TokenMgr::Drop() {
	return TRUE;
}

BOOL TokenMgr::Lift() {
	return TRUE;
}


BOOL TokenMgr::InsertToken(HANDLE Token, UINT id, UINT pid) {
	return TRUE;
}

BOOL TokenMgr::DeleteToken(UINT id) {
	return TRUE;
}
HANDLE TokenMgr::GetToken(UINT id) {
	return (HANDLE)-1
}

BOOL TokenMgr::AmImpersonating() {
	return this->isImpersonating == 0;
}


TokenMgr g_TokenMgr = NULL;