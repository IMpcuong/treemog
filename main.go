package main

import (
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/container"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

const CWD string = "/Users"

type block struct {
	title string
	size  uint32
	child []block
}

func accessibleEntry(relPath string) bool {
	if _, err := os.Stat(relPath); os.IsNotExist(err) ||
		syscall.Access(relPath, syscall.O_RDONLY) != nil ||
		!strings.Contains(relPath, string(os.PathSeparator)) {
		return false
	}
	return true

}

func listRawFiles(relPath string) (string, error) {
	if !accessibleEntry(relPath) {
		return "", fmt.Errorf("Not found baby!!!")
	}

	lsBinaries, _ := exec.LookPath("ls")
	lsArgs := []string{"-l", "-a", relPath}
	consumer := make(chan string)
	go func(args []string, c chan string) {
		stdout, _ := exec.Command(lsBinaries, args...).Output()
		c <- string(stdout)
	}(lsArgs, consumer)
	return <-consumer, nil
}

func convertToTreeMap(raw string) map[int]block {
	blockMap := make(map[int]block)
	if raw == "" || !strings.Contains(raw, "\n") {
		blockMap[-1] = block{}
		return blockMap
	}

	lines := strings.Split(raw, "\n")
	for incr, line := range lines[1 : len(lines)-1] {
		segments := strings.Split(line, " ")
		for i := len(segments) - 1; i > 0; i-- {
			if segments[i] == "" {
				segments = append(segments[:i], segments[i+1:]...)
			}
		}
		blockStat := block{}
		blockStat.title = segments[8]
		s, _ := strconv.Atoi(segments[4])
		blockStat.size = uint32(s)
		blockStat.child = []block{}
		blockMap[incr] = blockStat
	}
	return blockMap
}

func main() {
	a := app.New()
	w := a.NewWindow("TreeMog")

	queryBar := widget.NewEntry()
	queryBar.SetPlaceHolder("Do it...")

	pressSearchFnOnce := func() {
		files, err := listRawFiles(queryBar.Text)
		if err != nil {
			queryBar.SetText("")
			return
		}
		blocks := convertToTreeMap(files)
		objs := make([]fyne.CanvasObject, 0, len(blocks))
		for k, b := range blocks {
			rContent := fmt.Sprintf("%d> %s: %d", k, b.title, b.size)
			data := canvas.NewText(rContent, color.White)
			b := canvas.NewLine(color.White)
			c := container.NewBorder(b /* top */, b /* bot */, nil, nil, data)
			objs = append(objs, c)
		}
		w.SetContent(container.NewVBox(queryBar,
			container.NewGridWithRows(len(blocks), objs...)))
	}

	w.SetContent(container.NewVBox(
		queryBar,
		widget.NewButtonWithIcon("Query", theme.SearchIcon(), pressSearchFnOnce),
	))

	w.ShowAndRun()
}
