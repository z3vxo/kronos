#include "../hades/hades.h"


INT IntFromStr(PCHAR str) {
	UINT val = 0;
	while (*str >= '0' && *str <= '9') {
		val = val * 10 + (*str - '0');
		str++;
	}
	return val;
}