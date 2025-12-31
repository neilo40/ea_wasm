//go:build js && wasm

package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"syscall/js"

	"github.com/xuri/excelize/v2"
)

func (a *Arrangements) ReadFile() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		fileInput := js.Global().Get("document").Call("getElementById", "fileinput")
		fileInput.Get("files").Call("item", 0).Call("arrayBuffer").Call("then", js.FuncOf(func(v js.Value, x []js.Value) any {
			data := js.Global().Get("Uint8Array").New(x[0])
			dst := make([]byte, data.Get("length").Int())
			js.CopyBytesToGo(dst, data)
			a.inputBytes = dst
			doc := js.Global().Get("document")
			doc.Call("getElementById", "generateButton").Set("disabled", false)

			return nil
		}))
		return nil
	})
}

func (a *Arrangements) ReadSheetIntoDb() error {
	// parse sheet
	f, err := excelize.OpenReader(bytes.NewReader(a.inputBytes))
	if err != nil {
		return err
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.Rows(sheetName)
	if err != nil {
		return err
	}

	var colMap map[string]int

	for rows.Next() {
		row, err := rows.Columns()
		if err != nil {
			return err
		}
		// headers
		if colMap == nil {
			colMap = make(map[string]int)
			for i, cell := range row {
				colMap[cell] = i
			}
			continue
		}

		// done if we reached any empty rows
		if len(row) == 0 {
			break
		}

		// regular rows
		args := parseSheetRow(row, colMap)
		_, err = a.db.Exec(insertQuery, args...)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseSheetRow(row []string, colMap map[string]int) []any {
	args := make([]any, 0, 11)
	args = append(args, strings.TrimSpace(row[colMap["Name"]]))
	args = append(args, strings.TrimSpace(row[colMap["Surname"]]))
	args = append(args, strings.TrimSpace(row[colMap["Subject"]]))
	args = append(args, strings.TrimSpace(row[colMap["Paper"]]))
	args = append(args, strings.TrimSpace(row[colMap["Level"]]))
	args = append(args, strings.TrimSpace(row[colMap["Date"]]))
	args = append(args, strings.TrimSpace(row[colMap["Location"]]))
	args = append(args, timeToMins(strings.TrimSpace(row[colMap["Start"]])))
	args = append(args, timeToMins(strings.TrimSpace(row[colMap["Finish"]])))
	exTime := ""
	if len(row) > colMap["Ex Time"] {
		exTime = row[colMap["Ex Time"]]
	}
	args = append(args, exTime)
	notes := ""
	if len(row) > colMap["Notes"] {
		notes = row[colMap["Notes"]]
	}
	args = append(args, notes)

	return args
}

func timeToMins(t string) int {
	// mins since midnight for the given time t
	parts := strings.Split(t, ":")
	if len(parts) < 2 {
		return 0
	}
	totalMins := 0
	hours, _ := strconv.Atoi(parts[0])
	totalMins += 60 * hours
	mins, _ := strconv.Atoi(parts[1])
	totalMins += mins
	return totalMins
}

func minsToTime(m int) string {
	hours := m / 60
	mins := m % 60
	return fmt.Sprintf("%02d:%02d", hours, mins)
}
