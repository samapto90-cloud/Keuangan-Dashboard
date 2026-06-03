package main

import (
	"fmt"
	"os"

	"github.com/xuri/excelize/v2"
)

func main() {
	for _, fp := range os.Args[1:] {
		fmt.Println("================================================================================")
		fmt.Println("FILE:", fp)
		f, err := excelize.OpenFile(fp)
		if err != nil {
			fmt.Println("ERR:", err)
			continue
		}
		for _, sn := range f.GetSheetList() {
			rows, _ := f.GetRows(sn)
			fmt.Printf("--- Sheet: %s (%d rows) ---\n", sn, len(rows))
			limit := len(rows)
			if limit > 30 {
				limit = 30
			}
			for i := 0; i < limit; i++ {
				row := rows[i]
				parts := []string{}
				for j, c := range row {
					if j >= 25 {
						break
					}
					if c != "" {
						if len(c) > 55 {
							c = c[:55] + "…"
						}
						parts = append(parts, fmt.Sprintf("[%d]%s", j+1, c))
					}
				}
				if len(parts) > 0 {
					fmt.Printf("R%d: %s\n", i+1, joinParts(parts))
				}
			}
		}
		f.Close()
	}
}

func joinParts(p []string) string {
	s := ""
	for i, x := range p {
		if i > 0 {
			s += " | "
		}
		s += x
	}
	return s
}
