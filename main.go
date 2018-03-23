package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
)

var (
	jmx *JMX
	cfg *Conf
	// domainView *gocui.View
	beanView *gocui.View
	confView *gocui.View
	logView  *gocui.View
	active   = 0
	beanIdx  = 0
	beans    []string
	used     []bool
)

func main() {
	url, err := parseURL()
	if err != nil {
		panic(err.Error())
	}
	jmx = NewJMX(url)
	cfg = NewConf(url)
	buildView()
}

func parseURL() (string, error) {
	if len(os.Args) != 2 {
		return "", errors.New("host url not specified")
	}
	return os.Args[1], nil
}

func buildView() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		logInfo(err.Error())
	}
	defer g.Close()
	g.Highlight = true
	g.Cursor = true
	g.SelFgColor = gocui.ColorGreen
	g.SetManagerFunc(layout)
	bindActions(g)
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		logInfo(err.Error())
	}
}

func layout(g *gocui.Gui) error {
	first := false
	if logView == nil {
		first = true
	}
	// if err := domainViewLayout(g); err != nil {
	// 	return err
	if err := beanViewLayout(g); err != nil {
		return err
	} else if err := confViewLayout(g); err != nil {
		return err
	} else if err := logViewLayout(g); err != nil {
		return err
	}
	if first {
		refreshViewContent()
	}
	return nil
}

func bindActions(g *gocui.Gui) {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		panic(err.Error())
	}

	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		panic(err.Error())
	}

	// if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
	// 	panic(err.Error())
	// }

	if err := g.SetKeybinding("", 'j', gocui.ModNone, cursorDown); err != nil {
		panic(err.Error())
	}

	if err := g.SetKeybinding("", 'k', gocui.ModNone, cursorUp); err != nil {
		panic(err.Error())
	}

	if err := g.SetKeybinding("bean", gocui.KeyEnter, gocui.ModNone, selectBean); err != nil {
		panic(err.Error())
	}

	if err := g.SetKeybinding("", 's', gocui.ModNone, saveConf); err != nil {
		panic(err.Error())
	}
}

func saveConf(g *gocui.Gui, v *gocui.View) error {
	err := cfg.save()
	if err != nil {
		logInfo("save cfg to clipboard failed for: %s", err.Error())
	} else {
		logInfo("save cfg to clipboard success")
	}
	return nil
}

func selectBean(g *gocui.Gui, v *gocui.View) error {
	_, y := v.Cursor()
	content, err := v.Line(y)
	logInfo("get select mbean %s", content)
	if err != nil {
		return err
	}
	if content == "" {
		return nil
	}
	cfg.addRule(content)
	toggleBeanUsed(content)

	refreshConf()
	refreshBean()
	return nil
}

func toggleBeanUsed(content string) {
	for i, b := range beans {
		if b == content {
			used[i] = !used[i]
		}
	}
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	x, y := v.Cursor()
	y--
	if y < 0 && beanIdx > 0 {
		beanIdx--
		refreshBean()
		logInfo("bean idx %d", beanIdx)
		return nil
	} else if y < 0 {
		return nil
	}
	return v.SetCursor(x, y)
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	x, y := v.Cursor()
	_, ylimit := v.Size()
	y++
	if y >= ylimit && beanIdx < len(beans)-ylimit {
		beanIdx++
		refreshBean()
		return nil
	} else if y >= ylimit {
		return nil
	}
	return v.SetCursor(x, y)
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func nextView(g *gocui.Gui, v *gocui.View) error {
	nextIndex := (active + 1) % 3
	name := ""
	if nextIndex == 0 {
		name = "domains"
	} else if nextIndex == 1 {
		name = "bean"
	} else if nextIndex == 2 {
		name = "conf"
	}

	if _, err := setCurrentViewOnTop(g, name); err != nil {
		return err
	}
	active = nextIndex
	return nil
}

func setCurrentViewOnTop(g *gocui.Gui, name string) (*gocui.View, error) {
	if v, err := g.SetCurrentView(name); err != nil {
		return nil, err
	} else {
		return v, nil
	}
}

func confViewLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("conf", maxX/2, 0, maxX-1, maxY*2/3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Config"
		confView = v
	}
	return nil
}

// func domainViewLayout(g *gocui.Gui) error {
// 	maxX, maxY := g.Size()
// 	if v, err := g.SetView("domains", 0, 0, maxX/2, maxY/3); err != nil {
// 		if err != gocui.ErrUnknownView {
// 			return err
// 		}
// 		v.Title = "Domains"
// 		v.Autoscroll = true
// 		v.Wrap = true
// 		v.Highlight = true
// 		v.SelBgColor = gocui.ColorCyan
// 		v.SelFgColor = gocui.ColorMagenta
// 		domainView = v
// 		setCurrentViewOnTop(g, "domains")
// 	}
// 	return nil
// }

func beanViewLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("bean", 0, 0, maxX/2, maxY*2/3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Beans"
		v.Highlight = true
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorMagenta
		beanView = v
		setCurrentViewOnTop(g, "bean")
	}
	return nil
}

func logViewLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("log", 0, maxY*2/3, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Log"
		v.Autoscroll = true
		v.Wrap = true
		logView = v
	}
	return nil
}

func refreshViewContent() {
	logInfo("prepare to refresh content")
	// if d, err := jmx.domains(); err == nil {
	// 	domainView.Clear()
	// 	fmt.Fprint(domainView, strings.Join(d, "\n"))
	// } else {
	// 	logInfo("refresh domain failed for %s", err.Error())
	// }

	if bs, err := jmx.beans(); err == nil {
		for _, b := range bs {
			if b != "" {
				beans = append(beans, b)
			}
		}
		used = make([]bool, len(beans), len(beans))
		logInfo("get beans with length %d", len(beans))
	} else {
		logInfo("refresh beans failed for %s", err.Error())
	}
	refreshBean()
	refreshConf()
}

func refreshBean() {
	var result []string
	for i, b := range beans[beanIdx:] {
		if used[i+beanIdx] {
			result = append(result, fmt.Sprintf("\033[35;7m%s\033[0m", b))
		} else {
			result = append(result, b)
		}
	}
	beanView.Clear()
	fmt.Fprint(beanView, strings.Join(result, "\n"))
}

func refreshConf() {
	if c, err := cfg.dump(); err == nil {
		confView.Clear()
		fmt.Fprint(confView, c)
	}
}

func logInfo(format string, args ...interface{}) (int, error) {
	mesg := fmt.Sprintf(time.Now().String()+"  "+format+"\n", args...)
	return fmt.Fprintln(logView, mesg)
}
