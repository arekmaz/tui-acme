package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/jroimartin/gocui"
)

type View struct {
  id string
  pwd string
  tag string
  content string
}

type Label struct {
	name string
	body string
}

var pwd string

func NewLabel(name string, body string) *Label {
	return &Label{name: name, body: body}
}

func (wi *Label) Layout(g *gocui.Gui) error {
	lines := strings.Split(wi.body, "\n")
	w := 0

	for _, l := range lines {
		if len(l) > w {
			w = len(l)
		}
	}

	h := len(lines) + 2

	w = w + 1

	v, err := g.SetView(wi.name, 0, 0, w, h)

	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprint(v, wi.name)
		fmt.Fprint(v, "\n")
		fmt.Fprint(v, wi.body)
    fmt.Fprint(v, "w: " + strconv.Itoa(w) + ", h: " + strconv.Itoa(h))
	}

	return nil
}

func flowLayout(g *gocui.Gui) error {
	// maxX, maxY := g.Size()
	//
	//  paddingX := .5
	//  paddingY := .5
	//
	//  width := maxX/2 - int(paddingX * 2)
	//  height := maxY - int(paddingY * 2)
	//
	// if v, err := g.SetView("hello", maxX/2-width/2, maxY/2-height/2, maxX/2+width/2, maxY/2+height/2); err != nil {
	// 	if err != gocui.ErrUnknownView {
	// 		return err
	// 	}
	//
	//    pwd, err := os.Getwd()
	//
	//    if (err != nil) {
	//      return err
	//    }
	//
	//    fmt.Fprintln(v, "Newcol Kill Putall Dump Exit Dupa")
	// 	fmt.Fprintln(v, pwd)
	//    ls := exec.Command("ls")
	//
	//    o, err := ls.Output()
	//
	//    if err != nil {
	//      return err
	//    }
	//
	// 	fmt.Fprintln(v,string(o))
	// }

	views := g.Views()

	x := 0

	for _, view := range views {
		w, h := view.Size()

		_, err := g.SetView(view.Name(), x, 0, x + w + 1, h + 1)

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
  suff :=  "New Del Look"

  if (pwd == localPwd) {
    return suff
  }

  return pwd + suff
}


func main() {
  pwd, pwdErr := os.Getwd()

  if pwdErr != nil {
   panic(pwdErr.Error())
  }

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

  pwd, err := os.Getwd()

  if err != nil {
		log.Panicln(err)
  }

  l1 := NewLabel(pwd, safeRun("ls"))
	l2 := NewLabel("l2", safeRun("ls"))
	l3 := NewLabel("l3", "a")
	l4 := NewLabel("l4", "flow\nlayout")
	l5 := NewLabel("l5", "!")
	fl := gocui.ManagerFunc(flowLayout)
	g.SetManager(l1, l2, l3, l4, l5, fl)

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
