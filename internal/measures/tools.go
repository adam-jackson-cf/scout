package measures

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"scout/internal/analyzers"
	"scout/internal/domain"
)

// Pinned scanner versions (design §8.2). A different version is `unmeasured`:
// CC and clone detection change between releases, so a series must not mix them.
const (
	LizardVersion = "1.21.7"
	JscpdVersion  = "4.0.8"

	// jscpd parameters of design §8.2.
	jscpdMinTokens = 50
	jscpdMinLines  = 5
	jscpdMaxLines  = 10000
	jscpdMaxSize   = "500kb"

	defaultTimeout = 15 * time.Minute
	maxToolOutput  = 512 << 20
)

// Tools locates and runs the pinned scanners inside the read-only sandbox.
type Tools struct {
	Lizard  string // binary name or path; default "lizard"
	Jscpd   string // binary name or path; default "jscpd"
	Timeout time.Duration
}

func (t Tools) lizard() string {
	if t.Lizard != "" {
		return t.Lizard
	}
	return "lizard"
}

func (t Tools) jscpd() string {
	if t.Jscpd != "" {
		return t.Jscpd
	}
	return "jscpd"
}

func (t Tools) timeout() time.Duration {
	if t.Timeout > 0 {
		return t.Timeout
	}
	return defaultTimeout
}

// LizardDigest identifies the lizard tool and invocation.
func LizardDigest() string { return digestOf("lizard", LizardVersion, "--xml", "-f", "-t 4") }

// JscpdDigest identifies the jscpd tool and invocation.
func JscpdDigest() string {
	return digestOf("jscpd", JscpdVersion, fmt.Sprint(jscpdMinTokens), fmt.Sprint(jscpdMinLines), fmt.Sprint(jscpdMaxLines), jscpdMaxSize, "json", "span=start+lines")
}

// ConfigDigest identifies the eligibility rules and thresholds.
func ConfigDigest() string {
	extensions := make([]string, 0, len(sourceExtensions))
	for extension := range sourceExtensions {
		extensions = append(extensions, extension)
	}
	sort.Strings(extensions)
	return digestOf("measures-config@1", strings.Join(extensions, ","), strings.Join(nonProductionGlobs, ","),
		fmt.Sprint(MinFileNLOC), fmt.Sprint(HighCC), "scopes=feature+dir2@1")
}

func digestOf(parts ...string) string { return domain.Digest(parts...) }

// scannerEnvironment gives a scanner no inherited credentials; PATH holds the
// system directories plus the directory of the scanner (and of node, which
// jscpd's launcher needs).
func scannerEnvironment(state string, extraPath ...string) []string {
	path := append(append([]string{}, extraPath...), "/usr/local/bin", "/usr/bin", "/bin", "/usr/sbin", "/sbin")
	return []string{"HOME=" + state, "TMPDIR=" + state, "PATH=" + strings.Join(path, ":"), "LC_ALL=C"}
}

type limitedBuffer struct {
	bytes.Buffer
	limit    int
	exceeded bool
}

func (b *limitedBuffer) Write(value []byte) (int, error) {
	if b.Len()+len(value) > b.limit {
		b.exceeded = true
		return len(value), nil
	}
	return b.Buffer.Write(value)
}

// runTool runs binary in the sandbox with workDir as the current directory and
// returns stdout. Errors never carry scanner output.
func (t Tools) runTool(ctx context.Context, workDir, state, binary string, args []string, extraPath ...string) ([]byte, error) {
	resolved, err := exec.LookPath(binary)
	if err != nil {
		return nil, fmt.Errorf("%s is not installed", filepath.Base(binary))
	}
	if err := analyzers.SandboxAvailable(); err != nil {
		return nil, errors.New("read-only scanner sandbox is not installed")
	}
	ctx, cancel := context.WithTimeout(ctx, t.timeout())
	defer cancel()
	command, err := analyzers.SandboxedCommand(ctx, workDir, state, resolved, args)
	if err != nil {
		return nil, err
	}
	stdout, stderr := &limitedBuffer{limit: maxToolOutput}, &limitedBuffer{limit: 64 << 10}
	command.Stdout, command.Stderr = stdout, stderr
	command.Env = scannerEnvironment(state, append([]string{filepath.Dir(resolved)}, extraPath...)...)
	if err := command.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("%s timed out", filepath.Base(binary))
		}
		return nil, fmt.Errorf("%s failed: %w", filepath.Base(binary), err)
	}
	if stdout.exceeded {
		return nil, fmt.Errorf("%s output exceeds the bound", filepath.Base(binary))
	}
	return stdout.Bytes(), nil
}

// verifyVersion checks the scanner reports exactly the pinned version.
func (t Tools) verifyVersion(ctx context.Context, binary, want, state string, extraPath ...string) error {
	output, err := t.runTool(ctx, state, state, binary, []string{"--version"}, extraPath...)
	if err != nil {
		return err
	}
	if got := strings.TrimSpace(string(output)); got != want {
		return fmt.Errorf("%s version %q is not the pinned %s", filepath.Base(binary), got, want)
	}
	return nil
}

func newState() (string, error) {
	state, err := os.MkdirTemp("", "scout-measures-")
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(state)
	if err != nil {
		_ = os.RemoveAll(state)
		return "", err
	}
	return resolved, nil
}

// --- lizard ---------------------------------------------------------------

// Function is one function measured by lizard.
type Function struct {
	Name string `json:"name"`
	Line int    `json:"line"`
	CC   int    `json:"cc"`
	NLOC int    `json:"nloc"`
}

// FileMeasure is lizard's file-level NLOC and its functions.
type FileMeasure struct {
	Path      string     `json:"path"`
	NLOC      int        `json:"nloc"`
	Functions []Function `json:"functions"`
}

// RunLizard measures paths (repository-relative) under root.
func (t Tools) RunLizard(ctx context.Context, root string, paths []string) (map[string]FileMeasure, error) {
	if len(paths) == 0 {
		return map[string]FileMeasure{}, nil
	}
	state, err := newState()
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(state)
	if err := t.verifyVersion(ctx, t.lizard(), LizardVersion, state); err != nil {
		return nil, err
	}
	list := filepath.Join(state, "files.txt")
	if err := os.WriteFile(list, []byte(strings.Join(paths, "\n")+"\n"), 0o600); err != nil {
		return nil, err
	}
	output, err := t.runTool(ctx, root, state, t.lizard(), []string{"--xml", "-t", "4", "-f", list})
	if err != nil {
		return nil, err
	}
	return ParseLizardXML(output)
}

// ParseLizardXML reads lizard's --xml report: a Function measure
// (Nr, NCSS=NLOC, CCN per `name(...) at path:line`) and a File measure
// (Nr, NCSS=NLOC, CCN, Functions per path).
func ParseLizardXML(data []byte) (map[string]FileMeasure, error) {
	var document struct {
		Measures []struct {
			Type  string `xml:"type,attr"`
			Items []struct {
				Name   string `xml:"name,attr"`
				Values []int  `xml:"value"`
			} `xml:"item"`
		} `xml:"measure"`
	}
	if err := xml.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("decode lizard report: %w", err)
	}
	files := map[string]FileMeasure{}
	for _, measure := range document.Measures {
		if measure.Type != "File" {
			continue
		}
		for _, item := range measure.Items {
			if len(item.Values) < 2 {
				return nil, fmt.Errorf("lizard file item %q is malformed", item.Name)
			}
			files[item.Name] = FileMeasure{Path: item.Name, NLOC: item.Values[1]}
		}
	}
	for _, measure := range document.Measures {
		if measure.Type != "Function" {
			continue
		}
		for _, item := range measure.Items {
			if len(item.Values) < 3 {
				return nil, fmt.Errorf("lizard function item %q is malformed", item.Name)
			}
			at := strings.LastIndex(item.Name, " at ")
			if at < 0 {
				return nil, fmt.Errorf("lizard function item %q has no location", item.Name)
			}
			location := item.Name[at+4:]
			colon := strings.LastIndex(location, ":")
			if colon < 0 {
				return nil, fmt.Errorf("lizard function item %q has no line", item.Name)
			}
			var line int
			if _, err := fmt.Sscanf(location[colon+1:], "%d", &line); err != nil {
				return nil, fmt.Errorf("lizard function item %q has no line", item.Name)
			}
			path := location[:colon]
			name := item.Name[:at]
			if paren := strings.Index(name, "("); paren > 0 {
				name = name[:paren]
			}
			file, ok := files[path]
			if !ok {
				continue
			}
			file.Functions = append(file.Functions, Function{Name: strings.TrimSpace(name), Line: line, CC: item.Values[2], NLOC: item.Values[1]})
			files[path] = file
		}
	}
	return files, nil
}

// --- jscpd ----------------------------------------------------------------

// Span is an inclusive line range of a file.
type Span struct {
	Path       string `json:"path"`
	Start, End int
}

// Clone is one clone pair reported by jscpd.
type Clone struct {
	A, B Span
}

// RunJscpd detects clones among paths (repository-relative, already the eligible
// population) under root and returns them with each file's line count.
func (t Tools) RunJscpd(ctx context.Context, root string, paths []string) ([]Clone, error) {
	if len(paths) == 0 {
		return nil, nil
	}
	state, err := newState()
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(state)
	nodeDir := ""
	if node, err := exec.LookPath("node"); err == nil {
		nodeDir = filepath.Dir(node)
	}
	var extra []string
	if nodeDir != "" {
		extra = append(extra, nodeDir)
	}
	if err := t.verifyVersion(ctx, t.jscpd(), JscpdVersion, state, extra...); err != nil {
		return nil, err
	}
	absolute := make([]string, len(paths))
	for i, path := range paths {
		absolute[i] = filepath.Join(root, filepath.FromSlash(path))
	}
	output := filepath.Join(state, "report")
	config, err := json.Marshal(map[string]any{
		"path": absolute, "minTokens": jscpdMinTokens, "minLines": jscpdMinLines,
		"maxLines": jscpdMaxLines, "maxSize": jscpdMaxSize, "reporters": []string{"json"},
		"output": output, "silent": true, "absolute": true, "gitignore": false,
	})
	if err != nil {
		return nil, err
	}
	configPath := filepath.Join(state, "jscpd.json")
	if err := os.WriteFile(configPath, config, 0o600); err != nil {
		return nil, err
	}
	if _, err := t.runTool(ctx, root, state, t.jscpd(), []string{"--config", configPath}, extra...); err != nil {
		return nil, err
	}
	report, err := os.ReadFile(filepath.Join(output, "jscpd-report.json"))
	if err != nil {
		return nil, errors.New("jscpd wrote no report")
	}
	return ParseJscpdReport(report, root)
}

// ParseJscpdReport reads jscpd's json report; file names are made relative to
// root.
func ParseJscpdReport(data []byte, root string) ([]Clone, error) {
	var report struct {
		Duplicates []struct {
			Lines      int       `json:"lines"`
			FirstFile  jscpdFile `json:"firstFile"`
			SecondFile jscpdFile `json:"secondFile"`
		} `json:"duplicates"`
	}
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("decode jscpd report: %w", err)
	}
	clones := make([]Clone, 0, len(report.Duplicates))
	for _, duplicate := range report.Duplicates {
		first, err := duplicate.FirstFile.span(root, duplicate.Lines)
		if err != nil {
			return nil, err
		}
		second, err := duplicate.SecondFile.span(root, duplicate.Lines)
		if err != nil {
			return nil, err
		}
		clones = append(clones, Clone{A: first, B: second})
	}
	return clones, nil
}

type jscpdFile struct {
	Name  string `json:"name"`
	Start int    `json:"start"`
}

// span is the clone's line range in this file: its first line and jscpd's clone
// length. jscpd 4.0.8's reported `end` is not used: for the second file of a
// pair it is inconsistent with the clone (observed on fastapi: 40 of 148 spans
// end before they start or far beyond the clone), while the first file's end
// always equals start+lines-1.
func (f jscpdFile) span(root string, lines int) (Span, error) {
	path := f.Name
	if filepath.IsAbs(path) {
		relative, err := filepath.Rel(root, path)
		if err != nil || strings.HasPrefix(relative, "..") {
			return Span{}, fmt.Errorf("jscpd reported a path outside the repository")
		}
		path = relative
	}
	if f.Start < 1 || lines < 1 {
		return Span{}, errors.New("jscpd reported an invalid line span")
	}
	return Span{Path: filepath.ToSlash(path), Start: f.Start, End: f.Start + lines - 1}, nil
}
