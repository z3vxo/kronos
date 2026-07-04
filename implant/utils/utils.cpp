#include "../hades/hades.h"


INT IntFromStr(PCHAR str) {
	UINT val = 0;
	while (*str >= '0' && *str <= '9') {
		val = val * 10 + (*str - '0');
		str++;
	}
	return val;
}


NTSTATUS SetPrivilege(ULONG id, BOOL Enable) {
	BOOLEAN WasEnabled;
	NTSTATUS check; 
	check = hades->NtApis.RtlAdjustPrivilege(id, (BOOLEAN)Enable, FALSE, &WasEnabled);
	return check;
	
}