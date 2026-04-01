package postgres

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"gorm.io/gorm"
)

var wilayahTuplePattern = regexp.MustCompile(`\('([^']+)','([^']*)'\)`)

func EnsureIndonesiaRegionsSeeded(ctx context.Context, db *gorm.DB, dataPath string) error {
	if !db.Migrator().HasTable(&entity.AdministrativeRegion{}) {
		return nil
	}

	var count int64
	if err := db.WithContext(ctx).Model(&entity.AdministrativeRegion{}).Count(&count).Error; err != nil {
		return fmt.Errorf("count seeded regions: %w", err)
	}
	if count > 0 {
		return nil
	}

	file, err := os.Open(dataPath)
	if err != nil {
		return fmt.Errorf("open wilayah data: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024*16)

	batch := make([]entity.AdministrativeRegion, 0, 2000)
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		if err := db.WithContext(ctx).CreateInBatches(batch, 1000).Error; err != nil {
			return err
		}
		batch = batch[:0]
		return nil
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "(") {
			continue
		}

		matches := wilayahTuplePattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			code := match[1]
			name := strings.ReplaceAll(match[2], "''", "'")
			level, parentCode, ok := deriveRegionHierarchy(code)
			if !ok {
				continue
			}

			region := entity.AdministrativeRegion{
				Code:       code,
				Name:       decodeWilayahName(name),
				Level:      level,
				ParentCode: parentCode,
			}
			batch = append(batch, region)
			if len(batch) >= 2000 {
				if err := flush(); err != nil {
					return fmt.Errorf("seed region batch: %w", err)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan wilayah data: %w", err)
	}

	if err := flush(); err != nil {
		return fmt.Errorf("seed final region batch: %w", err)
	}

	return nil
}

func deriveRegionHierarchy(code string) (int, *string, bool) {
	switch strings.Count(code, ".") {
	case 0:
		return 1, nil, true
	case 1:
		parent := code[:2]
		return 2, &parent, true
	case 2:
		parent := code[:5]
		return 3, &parent, true
	case 3:
		parent := code[:8]
		return 4, &parent, true
	default:
		return 0, nil, false
	}
}

func decodeWilayahName(value string) string {
	return string(bytes.TrimSpace([]byte(value)))
}
