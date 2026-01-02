//go:build js && wasm

package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"syscall/js"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

type Arrangements struct {
	inputBytes        []byte
	PupilsSheet       []byte
	RoomUtilSheet     []byte
	RegistersSheet    []byte
	InvigilatorsSheet []byte
	db                *sql.DB
}

func (a *Arrangements) RunChecksCallbackFunc() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		errCaught := false
		doc := js.Global().Get("document")
		checksDiv := doc.Call("getElementById", "checks")
		for _, cf := range []func() error{a.CheckP1P2Gap, a.AllPupilsHaveBothPapers} {
			err := cf()
			if err != nil {
				errCaught = true
				checksDiv.Call("insertAdjacentHTML", "beforeend", err.Error())
			}
		}
		if !errCaught {
			checksDiv.Call("insertAdjacentHTML", "beforeend", "All checks passed")
		}
		return nil
	})
}

func (a *Arrangements) GenerateCallbackFunc() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		err := a.ReadSheetIntoDb()
		if err != nil {
			// TODO: also show error on html page
			panic(err)
		}

		err = a.GeneratePupilView()
		if err != nil {
			// TODO: also show error on html page
			panic(err)
		}

		err = a.GenerateRegistersView()
		if err != nil {
			// TODO: also show error on html page
			panic(err)
		}

		err = a.GenerateInvigilatorsView()
		if err != nil {
			// TODO: also show error on html page
			panic(err)
		}

		err = a.GenerateRoomUtilView()
		if err != nil {
			// TODO: also show error on html page
			panic(err)
		}

		// create download links using b64 encoded file content
		filenameSuffix := time.Now().Format("2006-01-02_15:04")
		doc := js.Global().Get("document")
		dloadDiv := doc.Call("getElementById", "downloads")
		dloadDiv.Call("insertAdjacentHTML", "beforeend", fmt.Sprintf(`
        <a download="pupils_%s.xlsx" href="/" id="pupils_link">Download Pupils Sheet</a><br />
        <a download="registers_%s.xlsx" href="/" id="registers_link">Download Registers Sheet</a><br />
        <a download="room_utilization_%s.xlsx" href="/" id="room_util_link">Download Room Utilization Sheet</a><br />
        <a download="invigilators_%s.xlsx" href="/" id="invigilators_link">Download Invigilators Sheet</a><br />
        `, filenameSuffix, filenameSuffix, filenameSuffix, filenameSuffix))
		link1 := doc.Call("getElementById", "pupils_link")
		link1.Set("href", fmt.Sprintf("data:application/octet-stream;base64,%s", base64.StdEncoding.EncodeToString(a.PupilsSheet)))
		link2 := doc.Call("getElementById", "registers_link")
		link2.Set("href", fmt.Sprintf("data:application/octet-stream;base64,%s", base64.StdEncoding.EncodeToString(a.RegistersSheet)))
		link3 := doc.Call("getElementById", "room_util_link")
		link3.Set("href", fmt.Sprintf("data:application/octet-stream;base64,%s", base64.StdEncoding.EncodeToString(a.RoomUtilSheet)))
		link4 := doc.Call("getElementById", "invigilators_link")
		link4.Set("href", fmt.Sprintf("data:application/octet-stream;base64,%s", base64.StdEncoding.EncodeToString(a.InvigilatorsSheet)))

		return nil
	})
}

func main() {
	// do any callback setup, then wait forever
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	db.Exec(createTableQuery)
	a := Arrangements{
		db: db,
	}
	doc := js.Global().Get("document")
	doc.Call("getElementById", "fileinput").Set("oninput", a.ReadFile())
	doc.Call("getElementById", "generateButton").Set("onclick", a.GenerateCallbackFunc())
	doc.Call("getElementById", "generateButton").Set("disabled", true)
	doc.Call("getElementById", "checkButton").Set("onclick", a.RunChecksCallbackFunc())
	doc.Call("getElementById", "checkButton").Set("disabled", true)
	select {}
}
