package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

// convertCEL converts a CEL file to CLX.
func convertCEL(source []byte, widths []int) ([]byte, error) {
	converted, err := assets.ConvertCELToCLX(source, widths)
	if err != nil {
		return nil, fmt.Errorf("convert CEL: %w", err)
	}

	return converted, nil
}

// convertCL2 converts a CL2 file to CLX.
func convertCL2(source []byte, widths []int) ([]byte, error) {
	converted, err := assets.ConvertCL2ToCLX(source, widths)
	if err != nil {
		return nil, fmt.Errorf("convert CL2: %w", err)
	}

	return converted, nil
}

// convertPCX converts a PCX with an optional transparent palette index.
func convertPCX(source []byte, frames, transparent int) ([]byte, error) {
	var transparentColor []byte
	if transparent >= 0 {
		transparentColor = []byte{byte(transparent)}
	}

	converted, err := assets.ConvertPCXToCLX(source, frames, transparentColor...)
	if err != nil {
		return nil, fmt.Errorf("convert PCX: %w", err)
	}

	return converted, nil
}

// validFormat reports whether format is a supported source container.
func validFormat(format string) bool {
	switch format {
	case "cel", "cl2", "pcx":
		return true
	default:
		return false
	}
}

// parseInts parses a comma-separated integer list; empty means none.
func parseInts(list string) ([]int, error) {
	if list == "" {
		return nil, nil
	}

	fields := strings.Split(list, ",")
	values := make([]int, len(fields))

	for i, field := range fields {
		value, err := strconv.Atoi(strings.TrimSpace(field))
		if err != nil {
			return nil, fmt.Errorf("width %q is not an integer: %w", field, err)
		}

		values[i] = value
	}

	return values, nil
}
