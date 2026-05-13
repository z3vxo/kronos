#include "files.hpp"


BOOL FileMgr::InsertTask(UINT32 TaskID, HANDLE handle) {
    for (FileTasks* cur = this->head; cur != NULL; cur = cur->next) {
        if (cur->TaskID == TaskID) {
            return FALSE;
        }
    }

    FileTasks* task = (FileTasks*)HeapAlloc(GetProcessHeap(), HEAP_ZERO_MEMORY, sizeof(FileTasks));
    if (!task) {
        return FALSE;
    }

    task->TaskID = TaskID;
    task->hProc = handle;
    task->Status = FileOnGoing;

    task->next = this->head;
    this->head = task;

    return TRUE;
}


BOOL FileMgr::ProcessEntry(FileTasks* task) {

	return TRUE;
}

BOOL FileMgr::CheckTasks() {
    FileTasks* prev = NULL;
    FileTasks* cur = this->head;

    while (cur) {
        FileTasks* next = cur->next;
        BOOL remove = this->ProcessEntry(cur);
        if (remove) {
            if (prev) {
                prev->next = next;
            }
            else {
                this->head = next;
            }

            if (cur->hProc && cur->hProc != INVALID_HANDLE_VALUE) {
                hades->WinApis.CloseHandle(cur->hProc);
                cur->hProc = NULL;
            }

            HeapFree(GetProcessHeap(), 0, cur);
        }
        else {
            prev = cur;
        }

        cur = next;
    }


    return TRUE;
}



FileMgr* g_FileMgr = NULL;