package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/jroimartin/gocui"
)

var pwd string

type Window struct {
	id      string
	pwd     string
	tag     string
	content string
}

func NewWindow(id string, pwd string, tag string, content string) *Window {
	return &Window{id: id, pwd: pwd, tag: tag, content: content}
}

func (wi *Window) DisplayData() (int, int, string) {
	lines := strings.Split(wi.content, "\n")
	w := 0

	for _, l := range lines {
		if len(l) > w {
			w = len(l)
		}
	}

	title := wi.id + " " + wi.tag

	w = max(len(title), w) + 1

	h := len(lines) + 4

	return w, h, title
}

func (wi *Window) Draw(v *gocui.View) {
	w, h, title := wi.DisplayData()
	fmt.Fprint(v, title)
	fmt.Fprint(v, "\n")
	fmt.Fprint(v, "\n")
	fmt.Fprint(v, wi.content)
	fmt.Fprint(v, "\n")
	fmt.Fprint(v, "w: "+strconv.Itoa(w)+", h: "+strconv.Itoa(h))
}

func (wi *Window) Layout(g *gocui.Gui) error {

	w, h, _ := wi.DisplayData()

	v, err := g.SetView(wi.id, 0, 0, w, h)

	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		wi.Draw(v)
	}

	return nil
}

func flowLayout(g *gocui.Gui) error {
	views := g.Views()

	x := 0

	for _, view := range views {
		w, h := view.Size()

		_, err := g.SetView(view.Name(), x, 0, x+w+1, h+1)

		if err != nil && err != gocui.ErrUnknownView {
			return err
		}

		x += w + 2
	}

	return nil
}

func safeRun(s string, args ...string) string {
	cmd := exec.Command(s, args...)

	o, err := cmd.Output()

	if err != nil {
		return s + "error"
	}

	return string(o)
}

func makeDefaultWindowTag(localPwd string) string {
	suff := "New Del Look"

	if pwd == localPwd {
		return suff
	}

	return pwd + suff
}

func readWindowIds() ([]string, error) {
	var ids []string
	err := filepath.WalkDir("./fs", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() || path == "./fs" {
			return nil
		}

		id := strings.Replace(path, "fs/", "", 1)

		ids = append(ids, id)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return ids, nil
}

func statPath(p string) (fs.FileInfo, error) {

	var path string = p

	if strings.HasSuffix(path, "~") {
		// Remove the last character
		path = path[:len(path)-1]
	}

	stat, err := os.Stat(path)

	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return stat, nil
}

func readWindowContent(id string) string {
	byteContent, err := os.ReadFile("fs/" + id + "/content")

	var content string

	if err == nil {
		content = string(byteContent)
	}

	return content
}

func readWindowTag(id string) string {
	var tag string

	byteTag, err := os.ReadFile("fs/" + id + "/tag")

	if err == nil {
		tag = strings.Trim(string(byteTag), " ")
	} else {
		tag = makeDefaultWindowTag(pwd)
	}

	return tag
}

func windowsFromIds(ids []string) []*Window {
	var windows []*Window

	for _, id := range ids {
		window := NewWindow(id, pwd, readWindowTag(id), readWindowContent(id))
		windows = append(windows, window)
	}

	return windows
}

func managersFromWindows(windows []*Window) []gocui.Manager {
	var managers []gocui.Manager

	for _, v := range windows {
		managers = append(managers, v)
	}

	managers = append(managers)

	return managers
}

func main() {
	pwd, pwdErr := os.Getwd()

	if pwdErr != nil {
		panic(pwdErr.Error())
	}

	fl := gocui.ManagerFunc(flowLayout)

	topLvlWatcher, err := fsnotify.NewWatcher()
	windowWatcher, err := fsnotify.NewWatcher()

	if err != nil {
		panic(err.Error())
	}

	defer topLvlWatcher.Close()
	defer windowWatcher.Close()

	g, err := gocui.NewGui(gocui.OutputNormal)

	if err != nil {
		panic(err.Error())
	}

	windowIds, err := readWindowIds()

	if err != nil {
		panic(err.Error())
	}

	err = topLvlWatcher.Add("./fs")

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case event, ok := <-topLvlWatcher.Events:
				if !ok {
					return
				}

				path := event.Name

				stat, err := statPath(path)

				if err != nil {
					panic("error 1 " + path + err.Error())
				}

				chunks := strings.Split(path, "/")
				id := chunks[1]

				content := readWindowContent(id)

				tag := readWindowTag(id)

				if stat != nil && !stat.IsDir() {
					continue
				}

				if stat == nil || event.Op == fsnotify.Remove {

					g.Update(func(g *gocui.Gui) error {
						if err := g.DeleteView(id); err != nil {
							return err
						}

						return nil
					})

					continue
				}

				if event.Op == fsnotify.Create {
					g.Update(func(g *gocui.Gui) error {
						window := NewWindow(id, pwd, tag, content)

						w, h, _ := window.DisplayData()

						v, err := g.SetView(id, 0, 0, w, h)

						if err != nil && err != gocui.ErrUnknownView {
							panic(err.Error())
						}

						window.Draw(v)
						return nil
					})

					continue
				}

			case err, ok := <-topLvlWatcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	go func() {
		for {
			select {
			case event, ok := <-windowWatcher.Events:
				if !ok {
					return
				}

				path := event.Name

				// chunks := strings.Split(path, "/")
				// id := chunks[1]

				// content := readWindowContent(id)
				//
				// tag := readWindowTag(id)

				stat, err := statPath(path)

				if err != nil {
					panic("error 1 " + path + err.Error())
				}

				if stat != nil && stat.IsDir() {
					continue
				}

				// g.Update(func(g *gocui.Gui) error {
				// 	v, err := g.View(id)
				// 	if err != nil {
				// 		return err
				// 	}
				// 	v.Clear()
				// 	window := NewWindow(id, pwd, tag, content)
				//
				// 	window.Draw(v)
				//
				// 	return nil
				// })

			case err, ok := <-windowWatcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	g.SetManager(fl)

	for _, window := range windowsFromIds(windowIds) {
		windowWatcher.Add("./fs/" + window.id)
     w,h,_ := window.DisplayData()
     v, err := g.SetView(window.id, 0, 0, w, h)

     if err != nil && err != gocui.ErrUnknownView {
       panic(err.Error())
     }

     window.Draw(v)
	}

	defer g.Close()



	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
