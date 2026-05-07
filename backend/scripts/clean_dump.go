package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	inputPath := "backend/database/migrations/schema_dump.sql"
	outputPath := "backend/database/migrations/schema_dump_clean.sql"

	file, err := os.Open(inputPath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	output, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer output.Close()

	scanner := bufio.NewScanner(file)
	writer := bufio.NewWriter(output)

	for scanner.Scan() {
		line := scanner.Text()

		// 1. Remove psql commands starting with \
		if strings.HasPrefix(line, "\\") {
			continue
		}

		// 2. Remove OWNER TO statements
		if strings.Contains(line, "OWNER TO") {
			continue
		}
		
		// 3. Remove AS $BODY$ if it's not needed (not here, pg_dump uses $$)
		
		_, _ = writer.WriteString(line + "\n")
	}

	writer.Flush()
	fmt.Println("Cleaned SQL dump created at", outputPath)
}
