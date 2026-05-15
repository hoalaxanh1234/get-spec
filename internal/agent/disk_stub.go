//go:build !windows

package agent

import "spec-collector/internal/models"

func getDisksIOCTL() []models.DiskInfo {
	return nil
}
