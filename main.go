package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func hexify(buffer []byte, strings [][]string) {
	for y, row := range strings {
		for x := range row {
			strings[y][x] = hex.EncodeToString(buffer[x+y*len(row) : x+y*len(row)+1])
		}
	}
}

func hexOffsets(offset int64, columns int, offsets [][]string) {
	for y := range offsets {
		offsets[y][0] = fmt.Sprintf("%x", offset+int64(y*columns))
	}
}

func main() {
	columns := 16
	rows := 16
	pageSize := columns * rows

	offsets := make([][]string, rows)
	f1strings := make([][]string, rows)
	f2strings := make([][]string, rows)
	for y := 0; y < rows; y++ {
		offsets[y] = make([]string, 1)
		f1strings[y] = make([]string, columns)
		f2strings[y] = make([]string, columns)
	}

	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <file1> <file2>\n", os.Args[0])
		os.Exit(0)
	}

	f1, err := openCreamyFile(os.Args[1], pageSize)
	if err != nil {
		log.Fatalf("failed to open f1: %v", err)
	}
	f1.Read()

	f2, err := openCreamyFile(os.Args[2], pageSize)
	if err != nil {
		log.Fatalf("failed to open f1: %v", err)
	}
	f2.Read()

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	left := 0
	offset := widgets.NewTable()
	offset.Border = false
	offset.BorderRight = true
	offset.TextStyle = ui.NewStyle(ui.ColorWhite)
	offset.RowSeparator = false
	offset.TextAlignment = ui.AlignRight
	offset.SetRect(left, 0, left+10, rows+2)
	left += 10

	f1table := widgets.NewTable()
	f1table.Border = false
	f1table.BorderRight = true
	f1table.BorderStyle.Fg = ui.ColorBlack
	f1table.TextStyle = ui.NewStyle(ui.ColorWhite)
	f1table.RowSeparator = false
	f1table.PaddingTop = 0
	f1table.PaddingBottom = 0
	f1table.Title = f1.path
	f1table.SetRect(left, 0, left+columns*4, rows+2)
	left += columns * 4

	left += 2
	f2table := widgets.NewTable()
	f2table.Border = false
	f2table.BorderStyle.Fg = ui.ColorBlack
	f2table.TextStyle = ui.NewStyle(ui.ColorWhite)
	f2table.RowSeparator = false
	f2table.PaddingTop = 0
	f2table.PaddingBottom = 0
	f2table.Title = f2.path
	f2table.SetRect(left, 0, left+columns*4, rows+2)
	left += columns * 4

	bufferPos := 0
	bufferY := 0
	bufferX := 0
	render := func() {
		f2.At(f1.offset)

		hexify(f1.buffer, f1strings)
		hexify(f2.buffer, f2strings)

		for bufferPos = 0; bufferPos < len(f1.buffer); bufferPos++ {
			if f1.buffer[bufferPos] != f2.buffer[bufferPos] {
				bufferY = bufferPos / rows
				bufferX = bufferPos - (bufferY * rows)
				f1strings[bufferY][bufferX] = "[" + f1strings[bufferY][bufferX] + "](fg:red)"
				f2strings[bufferY][bufferX] = "[" + f2strings[bufferY][bufferX] + "](fg:red)"
			}
		}

		f1table.Rows = f1strings
		f2table.Rows = f2strings
		ui.Render(f1table)
		ui.Render(f2table)

		hexOffsets(f1.offset, columns, offsets)
		offset.Rows = offsets
		ui.Render(offset)
	}

	render()

	uiEvents := ui.PollEvents()
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "<Down>":
				f1.Next(int64(columns))
				render()
				break
			case "<Up>":
				f1.Last(int64(columns))
				render()
				break
			case "<Left>":
				found := false
				for !found && !f1.IsAtBeginning() {
					f1.Last(int64(pageSize))
					f2.Last(int64(pageSize))
					for bufferPos = 0; bufferPos < len(f1.buffer); bufferPos++ {
						if f1.buffer[bufferPos] != f2.buffer[bufferPos] {
							found = true
							break
						}
					}
				}
				render()
				break
			case "<Right>":
				found := false
				for !found && !f1.IsAtEnd() {
					f1.Next(int64(pageSize))
					f2.Next(int64(pageSize))
					for bufferPos = 0; bufferPos < len(f1.buffer); bufferPos++ {
						if f1.buffer[bufferPos] != f2.buffer[bufferPos] {
							found = true
							break
						}
					}
				}
				render()
				break
			case "<PageDown>":
				f1.Next(int64(pageSize))
				render()
				break
			case "<PageUp>":
				f1.Last(int64(pageSize))
				render()
				break
			case "<Home>":
				f1.Start()
				render()
				break
			case "<End>":
				f1.End()
				render()
				break
			case "q", "<C-c>":
				return
			}
		}
	}
}
