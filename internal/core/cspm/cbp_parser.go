package cspm

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cloudboot/cloudboot-ng/internal/core/audit"
)

// CBPPackage represents the structure of a CloudBoot Package (.cbp)
type CBPPackage struct {
	Manifest  Manifest  `json:"manifest"`
	Watermark audit.Watermark `json:"watermark"`
	Signature string    `json:"signature"`
	// ProviderBinary is the encrypted provider binary
	ProviderBinary []byte `json:"-"`
}

// Manifest contains metadata about the provider
type Manifest struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Version          string   `json:"version"`
	Vendor           string   `json:"vendor"`
	Model            string   `json:"model"`
	SupportedHardware []string `json:"supported_hardware"`
	Description      string   `json:"description"`
	Author           string   `json:"author"`
	CreatedAt        string   `json:"created_at"`
}

// Note: Watermark type moved to internal/core/audit package to avoid duplication

// ParseCBP parses a .cbp package file
func ParseCBP(cbpPath string) (*CBPPackage, error) {
	// 打开ZIP文件
	reader, err := zip.OpenReader(cbpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open cbp file: %w", err)
	}
	defer reader.Close()

	pkg := &CBPPackage{}

	// 遍历ZIP文件内容
	for _, file := range reader.File {
		switch file.Name {
		case "meta/manifest.json":
			if err := parseManifest(file, pkg); err != nil {
				return nil, fmt.Errorf("failed to parse manifest: %w", err)
			}

		case "meta/watermark.json":
			if err := parseWatermark(file, pkg); err != nil {
				return nil, fmt.Errorf("failed to parse watermark: %w", err)
			}

		case "signature.sig":
			if err := parseSignature(file, pkg); err != nil {
				return nil, fmt.Errorf("failed to parse signature: %w", err)
			}

		case "bin/provider.enc":
			if err := parseProviderBinary(file, pkg); err != nil {
				return nil, fmt.Errorf("failed to parse provider binary: %w", err)
			}
		}
	}

	// 验证必需字段
	if pkg.Manifest.ID == "" {
		return nil, fmt.Errorf("manifest.json is missing or invalid")
	}
	if pkg.Signature == "" {
		return nil, fmt.Errorf("signature.sig is missing")
	}
	if len(pkg.ProviderBinary) == 0 {
		return nil, fmt.Errorf("provider.enc is missing")
	}

	return pkg, nil
}

func parseManifest(file *zip.File, pkg *CBPPackage) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &pkg.Manifest)
}

func parseWatermark(file *zip.File, pkg *CBPPackage) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &pkg.Watermark)
}

func parseSignature(file *zip.File, pkg *CBPPackage) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return err
	}

	pkg.Signature = string(data)
	return nil
}

func parseProviderBinary(file *zip.File, pkg *CBPPackage) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return err
	}

	pkg.ProviderBinary = data
	return nil
}

// CreateCBP creates a .cbp package file (used by build tools)
func CreateCBP(manifest Manifest, watermark audit.Watermark, encryptedBinary []byte, signature string, outputPath string) error {
	// 创建ZIP文件
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// 写入manifest.json
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	if err := writeZipFile(zipWriter, "meta/manifest.json", manifestData); err != nil {
		return err
	}

	// 写入watermark.json
	watermarkData, err := json.MarshalIndent(watermark, "", "  ")
	if err != nil {
		return err
	}
	if err := writeZipFile(zipWriter, "meta/watermark.json", watermarkData); err != nil {
		return err
	}

	// 写入加密的provider二进制
	if err := writeZipFile(zipWriter, "bin/provider.enc", encryptedBinary); err != nil {
		return err
	}

	// 写入签名
	if err := writeZipFile(zipWriter, "signature.sig", []byte(signature)); err != nil {
		return err
	}

	return nil
}

func writeZipFile(zipWriter *zip.Writer, name string, data []byte) error {
	writer, err := zipWriter.Create(name)
	if err != nil {
		return err
	}

	_, err = writer.Write(data)
	return err
}

// ExtractCBP extracts a .cbp package to a directory
func ExtractCBP(cbpPath string, destDir string) error {
	reader, err := zip.OpenReader(cbpPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		path := filepath.Join(destDir, file.Name)

		// 创建目录
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		// 创建父目录
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		// 写入文件
		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}
