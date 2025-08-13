package indexer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"lamina/pkg/ai"
	"lamina/pkg/database"
	"os"
	"path"
	"strings"

	"github.com/gomutex/godocx"
	"github.com/ledongthuc/pdf"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"gorm.io/gorm/clause"
)

func (i *Indexer) indexFile(ctx context.Context, filePath string) error {
	// Get file content first
	content, err := i.getFileContent(filePath)
	if err != nil {
		return err
	}

	// Skip if no content extracted (unsupported file type)
	if len(content) == 0 {
		fmt.Printf("⏭️  Skipping unsupported file type: %s\n", filePath)
		return nil
	}

	// Check if we should reindex based on content
	shouldReindex, err := i.shouldReindexWithContent(filePath, content)
	if err != nil {
		return err
	}
	if !shouldReindex {
		fmt.Printf("⏭️  Skipping unchanged file: %s\n", filePath)
		return nil
	}

	info, _ := os.Stat(filePath)
	contentHash := fmt.Sprintf("%x", sha256.Sum256(content))

	// Generate embeddings
	embeddingFloats, err := ai.GenerateEmbedding(ctx, string(content))
	if err != nil {
		return err
	}

	vectorBlob, err := sqlite_vec.SerializeFloat32(embeddingFloats)
	if err != nil {
		return err
	}

	// Save to DB
	file := database.File{
		Path:        filePath,
		ContentHash: contentHash,
		Size:        info.Size(),
		ModTime:     info.ModTime(),
		Content:     string(content),
	}

	// Use ON CONFLICT DO UPDATE for proper upsert
	if err := database.Store.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "path"}},
		DoUpdates: clause.AssignmentColumns([]string{"content_hash", "size", "mod_time", "content", "updated_at"}),
	}).Create(&file).Error; err != nil {
		return err

	}

	// Save embedding
	// Try to update first
	result := database.Store.Exec(`
	    UPDATE vec_embeddings SET embedding = ? WHERE file_id = ?
	`, vectorBlob, file.ID)

	// If no rows were affected, insert new record
	if result.RowsAffected == 0 {
		err = database.Store.Exec(`
			INSERT INTO vec_embeddings(file_id, embedding) 
			VALUES (?, ?)
		    `, file.ID, vectorBlob).Error
		if err != nil {
			return err
		}
	}

	fmt.Printf("✅ Successfully indexed: %s\n", file.ContentHash)
	return nil
}

func (i *Indexer) shouldReindexWithContent(filePath string, content []byte) (bool, error) {
	// Get file info
	info, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}

	// Hash the extracted content
	currentHash := fmt.Sprintf("%x", sha256.Sum256(content))

	// Check if file exists in DB
	var existingFile database.File
	err = database.Store.Where("path = ?", filePath).First(&existingFile).Error
	if err != nil {
		// File not in DB, needs indexing
		return true, nil
	}

	// Compare hash and mod time
	return existingFile.ContentHash != currentHash ||
		existingFile.ModTime.Before(info.ModTime()), nil
}

func (i *Indexer) shouldReindex(filePath string) (bool, error) {
	// Get file info
	info, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}

	// Read and hash content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	currentHash := fmt.Sprintf("%x", sha256.Sum256(content))

	// Check if file exists in DB
	var existingFile database.File
	err = database.Store.Where("path = ?", filePath).First(&existingFile).Error

	if err != nil {
		// File not in DB, needs indexing
		return true, nil
	}

	// Compare hash and mod time
	return existingFile.ContentHash != currentHash ||
		existingFile.ModTime.Before(info.ModTime()), nil
}

func (i *Indexer) getFileContent(filepath string) ([]byte, error) {
	extension := strings.ToLower(path.Ext(filepath))
	switch extension {
	case ".txt", ".md", ".go", ".py", ".js", ".ts", ".html", ".css", ".json", ".xml", ".yaml", ".yml", ".sh", ".bat", ".sql", ".log":
		return os.ReadFile(filepath)

	case ".pdf":
		return i.extractPDFContent(filepath)

	case ".docx":
		return i.extractDocxContent(filepath)

	// Image and video files - skip for now
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".webp":
		fmt.Printf("🖼️  Skipping image file (not yet supported): %s\n", filepath)
		return []byte{}, nil

	case ".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm", ".m4v":
		fmt.Printf("🎥 Skipping video file (not yet supported): %s\n", filepath)
		return []byte{}, nil

	case ".mp3", ".wav", ".flac", ".ogg", ".aac", ".m4a":
		fmt.Printf("🎵 Skipping audio file (not yet supported): %s\n", filepath)
		return []byte{}, nil

	case "":
		// Files without extension - try to read as text
		content, err := os.ReadFile(filepath)
		if err != nil {
			return nil, err
		}

		// gotten from grok
		// Check if it's likely text content (simple heuristic)
		if i.isLikelyTextContent(content) {
			return content, nil
		}
		fmt.Printf("❓ Skipping binary file without extension: %s\n", filepath)
		return []byte{}, nil

	default:
		fmt.Printf("❓ Skipping unsupported file type %s: %s\n", extension, filepath)
		return []byte{}, nil
	}
}

func (i *Indexer) extractPDFContent(filepath string) ([]byte, error) {
	f, r, err := pdf.Open(filepath)
	if err != nil {
		fmt.Printf("📄 PDF extraction failed for path: %s\n", filepath)
	}
	defer f.Close()

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		fmt.Printf("📄 PDF extraction to Plain text failed for path: %s\n", filepath)
	}
	buf.ReadFrom(b)
	return buf.Bytes(), nil
}

func (i *Indexer) extractDocxContent(filepath string) ([]byte, error) {
	document, err := godocx.OpenDocument(filepath)
	if err != nil {
		fmt.Printf("📄 DOCX extraction failed for path: %s\n", filepath)
	}
	var buf bytes.Buffer

	xmlEncoder := xml.NewEncoder(&buf)

	err = document.Document.Body.MarshalXML(xmlEncoder, xml.StartElement{})
	if err != nil {
		fmt.Printf("📄 DOCX extraction, Marshaling failed for path: %s with error: %s\n", filepath, err.Error())
	}

	err = xmlEncoder.Flush()
	if err != nil {
		fmt.Printf("📄 DOCX extraction, Flushing failed for path: %s with error: %s\n", filepath, err.Error())
	}

	return buf.Bytes(), nil
}

func (i *Indexer) isLikelyTextContent(content []byte) bool {
	if len(content) == 0 {
		return false
	}

	// gotten from grok
	// Check first 512 bytes for null bytes (common in binary files)
	checkLen := len(content)
	if checkLen > 512 {
		checkLen = 512
	}

	nullBytes := 0
	for i := 0; i < checkLen; i++ {
		if content[i] == 0 {
			nullBytes++
		}
	}

	// gotten from grok
	// If more than 1% null bytes, probably binary
	return float64(nullBytes)/float64(checkLen) < 0.01
}
