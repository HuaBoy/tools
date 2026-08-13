package knowledge

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
)

// ParseDocument 解析文档内容为纯文本
func ParseDocument(fileName string, data []byte) (string, error) {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".txt", ".md", ".markdown", ".log":
		return parseText(data)
	case ".json":
		return parseJSON(data)
	case ".csv":
		return parseCSV(data)
	case ".docx":
		return parseDocx(data)
	case ".pdf":
		return parsePDF(data)
	default:
		return "", fmt.Errorf("不支持的文件格式: %s", ext)
	}
}

// parseText 纯文本直接返回
func parseText(data []byte) (string, error) {
	return string(data), nil
}

// parseJSON 将 JSON 格式化为可读文本
func parseJSON(data []byte) (string, error) {
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		// JSON 解析失败则返回原始文本
		return string(data), nil
	}

	pretty, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return string(data), nil
	}
	return string(pretty), nil
}

// parseCSV 解析 CSV 为可读文本
func parseCSV(data []byte) (string, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return string(data), nil
	}

	var sb strings.Builder
	for _, record := range records {
		sb.WriteString(strings.Join(record, "\t"))
		sb.WriteString("\n")
	}
	return sb.String(), nil
}

// parseDocx 解析 DOCX（通过 ZIP + XML，纯标准库实现）
func parseDocx(data []byte) (string, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("无法解析 DOCX 文件: %w", err)
	}

	var sb strings.Builder
	for _, file := range zipReader.File {
		// 只解析正文内容
		if file.Name != "word/document.xml" {
			continue
		}

		f, err := file.Open()
		if err != nil {
			return "", err
		}
		content, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			return "", err
		}

		// 解析 XML 提取文本
		text, err := extractDocxText(content)
		if err != nil {
			return "", err
		}
		sb.WriteString(text)
	}

	if sb.Len() == 0 {
		return "", fmt.Errorf("DOCX 文件中未找到正文内容")
	}
	return sb.String(), nil
}

// docxBody DOCX XML 结构
type docxBody struct {
	XMLName xml.Name   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main document"`
	Body    docxInnerBody `xml:"body"`
}

type docxInnerBody struct {
	Paragraphs []docxParagraph `xml:"p"`
}

type docxParagraph struct {
	Runs []docxRun `xml:"r"`
}

type docxRun struct {
	Text []docxText `xml:"t"`
}

type docxText struct {
	Value string `xml:",chardata"`
}

// extractDocxText 从 document.xml 提取纯文本
func extractDocxText(content []byte) (string, error) {
	var doc struct {
		Body struct {
			Paragraphs []struct {
				Runs []struct {
					Text []struct {
						Value string `xml:",chardata"`
					} `xml:"t"`
				} `xml:"r"`
			} `xml:"p"`
		} `xml:"body"`
	}

	// 尝试带命名空间解析
	decoder := xml.NewDecoder(bytes.NewReader(content))
	var sb strings.Builder
	curText := false
	var textBuf strings.Builder

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			// 解析失败回退到简单字符串处理
			return extractTextFallback(content), nil
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "t" {
				curText = true
				textBuf.Reset()
			}
		case xml.CharData:
			if curText {
				textBuf.Write(t)
			}
		case xml.EndElement:
			if t.Name.Local == "t" {
				sb.WriteString(textBuf.String())
				curText = false
			} else if t.Name.Local == "p" {
				sb.WriteString("\n")
			}
		}
	}

	_ = doc
	return sb.String(), nil
}

// extractTextFallback 简单文本提取（不解析命名空间）
func extractTextFallback(content []byte) string {
	var sb strings.Builder
	decoder := xml.NewDecoder(bytes.NewReader(content))
	curText := false

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}
		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "t" {
				curText = true
			}
		case xml.CharData:
			if curText {
				sb.Write(t)
			}
		case xml.EndElement:
			if t.Name.Local == "t" {
				curText = false
			} else if t.Name.Local == "p" {
				sb.WriteString("\n")
			}
		}
	}
	return sb.String()
}

// parsePDF 解析 PDF（使用 ledongthuc/pdf 纯 Go 库）
func parsePDF(data []byte) (string, error) {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("解析 PDF 失败: %w", err)
	}

	var sb strings.Builder
	numPages := reader.NumPage()
	for i := 1; i <= numPages; i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		content, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		sb.WriteString(content)
		sb.WriteString("\n\n")
	}

	if strings.TrimSpace(sb.String()) == "" {
		return "", fmt.Errorf("PDF 无可提取文本（可能为扫描件或加密）")
	}
	return sb.String(), nil
}
