package service

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ledongthuc/pdf"
)

type PDFService interface {
	ExtractText(filePath string) (string, error)
}

type pdfService struct{}

func NewPDFService() PDFService{
	return &pdfService{}
}

func (ps *pdfService) ExtractText(filepath string) (string, error) {
	file, reader, err := pdf.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to open pdf: %w", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	totalPages := reader.NumPage()

	// extract text
	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		page := reader.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}

		buf.WriteString(text)
		buf.WriteString("\n")
	}

	// clean extracted text
	cleanedText := ps.cleanText(buf.String())
	if cleanedText == "" {
		return "", fmt.Errorf("no text content found in pdf")
	}

	return cleanedText, nil

}

func (ps *pdfService) cleanText(text string) string {
	text = strings.TrimSpace(text)

	lines := strings.Split(text, "\n")
	var cleaned []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}

	return strings.Join(cleaned, "\n")
} 


