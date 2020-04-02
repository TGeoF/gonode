// Demo code for the Table primitive.
package main

import (
	"database/sql"
	"fmt"
	"github.com/gdamore/tcell"
	_ "github.com/mattn/go-sqlite3"
	"gitlab.com/tslocum/cview"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

const batchSize = 100
const appName = "CYBER SPRAWL CLASSICS DATABASE 3.15.4alpha"

func loadRows(tableName string, offset int, content *cview.Table, db *sql.DB) {
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM %s LIMIT $1 OFFSET $2", tableName), batchSize, offset)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	columnNames, err := rows.Columns()
	if err != nil {
		log.Fatal(err)
	}
	for index, name := range columnNames {
		content.SetCell(0, index, &cview.TableCell{Text: name, Align: cview.AlignCenter, Color: tcell.ColorYellow})
	}

	//read the rows
	columns := make([]interface{}, len(columnNames))
	columnPointers := make([]interface{}, len(columns))
	for index := range columnPointers {
		columnPointers[index] = &columns[index]
	}
	for rows.Next() {
		// read the columns
		err := rows.Scan(columnPointers...)
		if err != nil {
			log.Fatal(err)
		}

		// transfer them to the table
		row := content.GetRowCount()
		for index, column := range columns {
			switch value := column.(type) {
			case int64:
				content.SetCell(row, index, &cview.TableCell{Text: strconv.Itoa(int(value)), Align: cview.AlignRight, Color: tcell.ColorDarkCyan})
			case float64:
				content.SetCell(row, index, &cview.TableCell{Text: strconv.FormatFloat(value, 'f', 2, 64), Align: cview.AlignRight, Color: tcell.ColorDarkCyan})
			case string:
				content.SetCellSimple(row, index, value)
			case time.Time:
				t := value.Format("2006-01-02")
				content.SetCell(row, index, &cview.TableCell{Text: t, Align: cview.AlignRight, Color: tcell.ColorDarkMagenta})
			case []uint8:
				str := make([]byte, len(value))
				for index, num := range value {
					str[index] = byte(num)
				}
				content.SetCell(row, index, &cview.TableCell{Text: string(str), Align: cview.AlignRight, Color: tcell.ColorGreen})
			case nil:
				content.SetCell(row, index, &cview.TableCell{Text: "NULL", Align: cview.AlignCenter, Color: tcell.ColorRed})
			default:
				// We've encountered a type that we don't know yet.
				t := reflect.TypeOf(value)
				str := "?nil?"
				if t != nil {
					str = "?" + t.String() + "?"
				}
				content.SetCellSimple(row, index, str)
			}
		}
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	// show how much we've loaded
	// frame.Clear()
	// loadMore := ""
	// if content.GetRowCount()-1 < rowCount {
	// loadMore = " - press Enter to load more"
	// }
	// loadMore = fmt.Sprintf("Loaded %d of %d rows%s", content.GetRowCount()-1, rowCount, loadMore)
	// frame.AddText(loadMore, false, cview.AlignCenter, tcell.ColorYellow).AddText(appName, true, cview.AlignLeft, tcell.ColorGreen)
}

func main() {
	app := cview.NewApplication()
	content := cview.NewTable().SetBorders(true).SetFixed(1, 0).SetBordersColor(tcell.ColorBlue)
	content.SetBorder(true).SetTitle("Content (F2)")
	tables := cview.NewList()
	tables.ShowSecondaryText(false).
		SetDoneFunc(func() {
			tables.Clear()
			app.SetFocus(tables)
		})
	tables.SetBorder(true).SetTitle("Tables (F1)")

	flex := cview.NewFlex().
		AddItem(tables, 0, 1, true).
		AddItem(content, 0, 3, false)

	frame := cview.NewFrame(flex).AddText(appName, true, cview.AlignLeft, tcell.ColorGreen)

	db, err := sql.Open("sqlite3", "./csc.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	t, err := db.Query("select name from sqlite_master where type = 'table';")
	if err != nil {
		log.Fatal(err)
	}
	defer t.Close()

	tablesList := []string{}
	for t.Next() {
		var tableName string
		if err := t.Scan(&tableName); err != nil {
			log.Fatal(err)
		}
		tablesList = append(tablesList, tableName)
	}
	sort.Strings(tablesList)
	for _, tab := range tablesList {
		tab = strings.ToUpper(tab)
		tables.AddItem(tab, "", 0, nil)
	}

	tables.SetChangedFunc(func(i int, tableName string, t string, s rune) {
		content.Clear()

		// how many rows does the table have?
		// var rowCount int
		// err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&rowCount)
		// if err != nil {
		// log.Fatal(err)
		// }

		// load the first batch of rows
		loadRows(tableName, 0, content, db)

	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			app.Stop()
			return nil
		case tcell.KeyF1:
			app.SetFocus(tables)
		case tcell.KeyF2:
			app.SetFocus(content)
		}
		return event
	})

	tables.SetCurrentItem(0)

	if err := app.SetRoot(frame, true).Run(); err != nil {
		fmt.Printf("Error running application: %s\n", err)
	}
}
