#include "../hades/hades.h"
// #include "../shared/list.hpp"



// struct Token {
// 	UINT id;
// 	UINT StolenProcessPID;
// 	HANDLE TokenHandle;
// };

// class TokenMgr {
// 	HANDLE RealToken;
// 	HANDLE TempImpersonateToken;
// 	HANDLE CurrentToken;
// 	List<Token> Tokens;
// 	BOOL isImpersonating;


// public:
// 	BOOL Drop();
// 	BOOL Lift();
// 	BOOL AmImpersonating();
// 	BOOL InsertToken(HANDLE Token, UINT id, UINT pid);
// 	HANDLE GetToken(UINT id);
// 	BOOL DeleteToken(UINT id);



// };

// extern TokenMgr* g_TokenMgr;