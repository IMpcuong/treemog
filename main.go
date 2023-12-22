package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"fyne.io/fyne/app"
	"fyne.io/fyne/container"
	"fyne.io/fyne/widget"
)

const CWD string = "~/github"

type block struct {
	title string
	size  uint32
	child []block
}

func accessibleEntry(relPath string) bool {
	return syscall.Access(relPath, syscall.O_RDONLY) == nil
}

func listRawFiles(relPath string) string {
	if !accessibleEntry(relPath) {
		return os.DevNull
	}

	lsBinaries, _ := exec.LookPath("ls")
	lsArgs := []string{"-l", "-a", relPath}
	consumer := make(chan string)
	go func(args []string, c chan string) {
		stdout, _ := exec.Command(lsBinaries, args...).Output()
		c <- string(stdout)
	}(lsArgs, consumer)
	return <-consumer
}

func convertToTreeMap(raw string) map[int]block {
	if !strings.Contains(raw, "\n") {
		return map[int]block{}
	}

	blockMap := make(map[int]block)
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
	files := listRawFiles(CWD)
	fmt.Printf("%+v", convertToTreeMap(files))
	blocks := convertToTreeMap(files)

	a := app.New()
	w := a.NewWindow("TreeMog")

	treeMap := widget.NewLabel("Files")
	w.SetContent(container.NewVBox(
		treeMap,
		widget.NewButton("Click me", func() {
			for _, b := range blocks {
				treeMap.SetText(b.title + ": " + strconv.Itoa(int(b.size)))
			}
		}),
	))

	w.ShowAndRun()
}
