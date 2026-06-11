package bytemgr

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/z3vxo/kronos/internal/database"
	"github.com/z3vxo/kronos/internal/files"
)

func Write4(w io.Writer, val any) error {
	return binary.Write(w, binary.LittleEndian, val)
}

func WriteString(w io.Writer, str string) error {
	strLen := uint32(len(str) + 1)
	if err := Write4(w, strLen); err != nil {
		return err
	}
	if _, err := w.Write([]byte(str)); err != nil {
		return err
	}
	if _, err := w.Write([]byte{0}); err != nil {
		return err
	}
	return nil
}





func CraftCmdBytes(tasks []database.Task, fileMgr *files.Manager) ([]byte, error) {
	var buffer bytes.Buffer
	if err := Write4(&buffer, int32(len(tasks))); err != nil {
		return nil, err
	}
	for _, c := range tasks {
		if err := Write4(&buffer, int32(c.CmdCode)); err != nil {
			return nil, err
		}
		if err := Write4(&buffer, c.TaskID); err != nil {
			return nil, err
		}

		if c.CmdCode == 13 {
			if err := processUploadCmd(&buffer, c, fileMgr); err != nil {
				return nil, err
			}
			continue
		}

		if c.Param1 != "" {
			if err := WriteString(&buffer, c.Param1); err != nil {
				return nil, err
			}
		}
		if c.Param2 != "" {
			if err := WriteString(&buffer, c.Param2); err != nil {
				return nil, err
			}
		}
	}

	return buffer.Bytes(), nil
}

func processUploadCmd(buf *bytes.Buffer, cmd database.Task, fileMgr *files.Manager) error {
	uuid := cmd.Param2
	entry, ok := fileMgr.Uploads[uuid]
	if !ok {
		return fmt.Errorf("upload entry not found: %s", uuid)
	}

	data, final, err := fileMgr.ReadUploadChunk(uuid)
	if err != nil {
		return err
	}

	switch entry.Status {
	case files.UploadStatusNotStarted:
		if final {
			if err := Write4(buf, uint32(UPLOAD_START_NON_CHUNKED)); err != nil {
				return err
			}
		} else {
			if err := Write4(buf, uint32(UPLOAD_START_CHUNKED)); err != nil {
				return err
			}
		}
		if err := WriteString(buf, cmd.Param1); err != nil {
			return err
		}
		if !final {
			entry.Status = files.UploadStatusOngoing
		} else {
			entry.Status = files.UploadStatusDone
		}

	case files.UploadStatusOngoing:
		if final {
			if err := Write4(buf, uint32(UPLOAD_DONE)); err != nil {
				return err
			}
		} else {
			if err := Write4(buf, uint32(UPLOAD_CONTINUE)); err != nil {
				return err
			}
		}
		if final {
			entry.Status = files.UploadStatusDone
		}
	}

	if err := Write4(buf, uint32(len(data))); err != nil {
		return err
	}
	if _, err := buf.Write(data); err != nil {
		return err
	}

	return nil
}
