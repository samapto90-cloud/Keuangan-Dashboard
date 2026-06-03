package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/xuri/excelize/v2"
)

func main() {
	for _, fp := range os.Args[1:] {
		if fp == "0" || fp == "76" {
			continue
		}
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
			start := 43
			end := len(rows)
			if len(os.Args) > 2 {
				start, _ = strconv.Atoi(os.Args[2])
			}
			if len(os.Args) > 3 {
				end, _ = strconv.Atoi(os.Args[3])
			}
			for i := start; i < end && i < len(rows); i++ {
				row := rows[i]
				parts := []string{}
				for j, c := range row {
					if j >= 8 {
						break
					}
					if c != "" {
						if len(c) > 60 {
							c = c[:60] + "…"
						}
						parts = append(parts, fmt.Sprintf("[%d]%s", j+1, c))
					}
				}
				fmt.Printf("R%d: %s\n", i+1, joinParts(parts))
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
