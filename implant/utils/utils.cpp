#include "../hades/hades.h"


INT IntToStr(DWORD val, CHAR* out) {
	INT i = 0;

	if (val == 0) {
		out[i++] = '0';
		out[i++] = '\0';
		return i;
	}

	CHAR tmp[16];
	INT j = 0;

	while (val > 0) {
		tmp[j++] = '0' + (val % 10);
		val /= 10;
	}
	while (j > 0) {
		out[i++] = tmp[--j];
	}
	out[i] = '\0';
	return i;
}