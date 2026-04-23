package bytemgr

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/z3vxo/kronos/internal/database"
)

func Write4(w io.Writer, val any) error {
	return binary.Write(w, binary.LittleEndian, val)
}

func WriteString(w io.Writer, str string) error {
	strLen := uint32(len(str))
	if err := Write4(w, strLen); err != nil {
		return err
	}
	if _, err := w.Write([]byte(str)); err != nil {
		return err
	}
	return nil
}

// loop of tasks
/* [TASK AMOUNT] 4 BYTES
 * ->
 * [CMD CODE] 4 BYTES
 * ->
 * [TASK ID] 4 BYTES
 * ->
 * [PARAM 1 LEN] 4 BYTES
 * [PARAM 1 STR] N BYTES
 * ->
 * [PARAM 2 LEN] 4 BYTES
 * [PARAM 2 STR] N BYTES
 *
 */

func CraftCmdBytes(tasks []database.Task) ([]byte, error) {
	var buffer bytes.Buffer
	err := Write4(&buffer, int32(len(tasks)))
	if err != nil {
		return nil, err
	}
	for _, c := range tasks {
		if err := Write4(&buffer, int32(c.CmdCode)); err != nil {
			return nil, err
		}
		if err := Write4(&buffer, int32(c.TaskID)); err != nil {
			return nil, err
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
