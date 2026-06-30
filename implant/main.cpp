#include "hades/hades.h"


#ifdef BUILD_EXE
int main() {
	if (!RunHades()) { return 1; }

	return 0;

}
#endif

#ifdef BUILD_DLL
DWORD WINAPI HadesThread(LPVOID lpParam) {
	RunHades();
	return 0;
}

BOOL WINAPI DllMain(HINSTANCE hinstDLL, DWORD fdwReason, LPVOID lpvReserved) {
	if (fdwReason == DLL_PROCESS_ATTACH) {
		CreateThread(NULL, 0, HadesThread, NULL, 0, NULL);
	}
	return TRUE;
}
#endif

