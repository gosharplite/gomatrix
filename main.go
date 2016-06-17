package main

import (
	"fmt"
	"github.com/gdamore/tcell"
	"math/rand"
	"os"
	"os/signal"
	"time"
)

var screen tcell.Screen

var characters = []rune{
	'上', '海', '卓', '易', '科', '技', '臺', '台', '北', '研', '究', '院',
	'D', 'r', 'o', 'i', 'T', 'a', 'i', 'p', 'e', 'i', '1', '0', '1',
	'S', 'h', 'a', 'n', 'g', 'h', 'a', 'i', 'B', 'e', 'i', 'j', 'i', 'n', 'g',
}

var streamDisplaysByColumn = make(map[int]*StreamDisplay)

type sizes struct {
	width  int
	height int
}

var curSizes sizes

var sizesUpdateCh = make(chan sizes)

func setSizes(width int, height int) {
	s := sizes{
		width:  width,
		height: height,
	}
	curSizes = s
	sizesUpdateCh <- s
}

func main() {

	rand.Seed(time.Now().UnixNano())

	var err error

	if screen, err = tcell.NewScreen(); err != nil {
		fmt.Println("Could not start tcell.")
		os.Exit(1)
	}

	err = screen.Init()
	if err != nil {
		fmt.Println("Could not start tcell.")
		os.Exit(1)
	}
	screen.HideCursor()
	screen.SetStyle(tcell.StyleDefault.
		Background(tcell.ColorBlack).
		Foreground(tcell.ColorBlack))
	screen.Clear()

	go func() {
		var lastWidth int

		for newSizes := range sizesUpdateCh {
			diffWidth := newSizes.width - lastWidth

			if diffWidth == 0 {
				continue
			}

			if diffWidth > 0 {
				for newColumn := lastWidth; newColumn < newSizes.width; newColumn++ {
					sd := &StreamDisplay{
						column:    newColumn,
						stopCh:    make(chan bool, 1),
						streams:   make(map[*Stream]bool),
						newStream: make(chan bool, 1),
					}
					streamDisplaysByColumn[newColumn] = sd

					go sd.run()

					sd.newStream <- true
				}
				lastWidth = newSizes.width
			}

			if diffWidth < 0 {
				for closeColumn := lastWidth - 1; closeColumn > newSizes.width; closeColumn-- {
					sd := streamDisplaysByColumn[closeColumn]

					delete(streamDisplaysByColumn, closeColumn)

					sd.stopCh <- true
				}
				lastWidth = newSizes.width
			}
		}
	}()

	setSizes(screen.Size())

	go func() {
		for {
			time.Sleep(40 * time.Millisecond)
			screen.Show()
		}
	}()

	eventChan := make(chan tcell.Event)
	go func() {
		for {
			event := screen.PollEvent()
			eventChan <- event
		}
	}()

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, os.Kill)

EVENTS:
	for {
		select {
		case event := <-eventChan:
			switch ev := event.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyCtrlZ, tcell.KeyCtrlC:
					break EVENTS

				case tcell.KeyCtrlL:
					screen.Sync()

				case tcell.KeyRune:
					switch ev.Rune() {
					case 'q':
						break EVENTS
					case 'c':
						screen.Clear()
					}
				}
			case *tcell.EventResize:
				w, h := ev.Size()
				setSizes(w, h)
			case *tcell.EventError:
				os.Exit(1)
			}
		case <-sigChan:
			break EVENTS
		}
	}

	screen.Fini()
}
