package agent

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"spec-collector/internal/models"
)

//go:embed 3rdparty/cdi.zip
var cdiZip []byte

func extractCDI(dir string) error {
	r, err := zip.NewReader(bytes.NewReader(cdiZip), int64(len(cdiZip)))
	if err != nil {
		return err
	}
	for _, f := range r.File {
		fpath := filepath.Join(dir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0755)
			continue
		}
		os.MkdirAll(filepath.Dir(fpath), 0755)
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(fpath)
		if err != nil {
			rc.Close()
			return err
		}
		io.Copy(out, rc)
		out.Close()
		rc.Close()
	}
	return nil
}

type cdiEntry struct {
	drive  string
	model  string
	iface  string
	sizeMB float64
	health int
}

func parseCDIOutput(text string) []models.DiskInfo {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	var cur cdiEntry
	var drives []cdiEntry

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			if cur.drive != "" && cur.model != "" && cur.iface != "" {
				drives = append(drives, cur)
			}
			cur = cdiEntry{}
			continue
		}
		switch {
		case strings.HasPrefix(line, "Drive :"):
			cur.drive = strings.TrimSpace(strings.TrimPrefix(line, "Drive :"))
		case strings.HasPrefix(line, "Model :"):
			cur.model = strings.TrimSpace(strings.TrimPrefix(line, "Model :"))
			if idx := strings.Index(cur.model, "(firmware:"); idx > 0 {
				cur.model = strings.TrimSpace(cur.model[:idx])
			}
		case strings.HasPrefix(line, "Interface :"):
			cur.iface = strings.TrimSpace(strings.TrimPrefix(line, "Interface :"))
		case strings.HasPrefix(line, "Health :"):
			s := strings.TrimSpace(strings.TrimPrefix(line, "Health :"))
			s = strings.TrimSuffix(s, "%")
			s = strings.TrimSpace(s)
			fmt.Sscanf(s, "%d", &cur.health)
		case strings.HasPrefix(line, "Size :"):
			s := strings.TrimSpace(strings.TrimPrefix(line, "Size :"))
			s = strings.ReplaceAll(s, ",", "")
			s = strings.ReplaceAll(s, " ", "")
			if strings.HasSuffix(s, "GB") {
				v := strings.TrimSuffix(s, "GB")
				fmt.Sscanf(v, "%f", &cur.sizeMB)
				cur.sizeMB *= 1024
			} else if strings.HasSuffix(s, "MB") {
				v := strings.TrimSuffix(s, "MB")
				fmt.Sscanf(v, "%f", &cur.sizeMB)
			}
		}
	}
	if cur.drive != "" && cur.model != "" && cur.iface != "" {
		drives = append(drives, cur)
	}

	var result []models.DiskInfo
	for _, d := range drives {
		t := classifyCDIIface(d.iface, d.model)
		sz := 0.0
		if d.sizeMB > 0 {
			sz = d.sizeMB / 1024
		}
		result = append(result, models.DiskInfo{
			Model:     d.model,
			SizeGB:    sz,
			Type:      t,
			HealthPct: d.health,
		})
	}
	return result
}

func classifyCDIIface(iface, model string) string {
	i := strings.ToLower(iface)
	m := strings.ToLower(model)
	switch {
	case strings.Contains(i, "nvme") || strings.Contains(i, "nvm express"):
		gen := nvmeGen(model)
		if gen != "" {
			return "NVMe " + gen
		}
		return "NVMe"
	case strings.Contains(i, "sata") || strings.Contains(i, "serial ata"):
		if strings.Contains(m, "ssd") {
			return "SATA SSD"
		}
		if strings.Contains(m, "hdd") || strings.Contains(m, "hard") {
			return "SATA HDD"
		}
		return "SATA SSD"
	case strings.Contains(i, "usb"):
		return "USB"
	case strings.Contains(i, "sas"):
		return "SAS"
	default:
		if strings.Contains(m, "nvme") {
			gen := nvmeGen(model)
			if gen != "" {
				return "NVMe " + gen
			}
			return "NVMe"
		}
		if strings.Contains(m, "ssd") {
			return "SSD"
		}
		return iface
	}
}

func runEmbeddedCrystalDiskInfo() []models.DiskInfo {
	dir, err := os.MkdirTemp("", "sc-cdi-*")
	if err != nil {
		return nil
	}
	defer os.RemoveAll(dir)

	if err := extractCDI(dir); err != nil {
		return nil
	}

	exePath := filepath.Join(dir, "DiskInfo64.exe")
	if _, err := os.Stat(exePath); err != nil {
		exePath = filepath.Join(dir, "DiskInfo32.exe")
		if _, err := os.Stat(exePath); err != nil {
			return nil
		}
	}

	cmd := exec.Command(exePath, "/CopyExit", "/Silent")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return nil
	}

	txtPath := filepath.Join(dir, "DiskInfo.txt")
	out, err := os.ReadFile(txtPath)
	if err != nil {
		return nil
	}

	return parseCDIOutput(string(out))
}
