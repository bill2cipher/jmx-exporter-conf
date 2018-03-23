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
	jmx        *JMX
	cfg        *Conf
	domainView *gocui.View
	beanView   *gocui.View
	confView   *gocui.View
	logView    *gocui.View
)

func main() {
	url, err := parseURL()
	if err != nil {
		panic(err.Error())
	}
	jmx = NewJMX(url)
	cfg = NewConf(url)
	go scheduleRefresh()
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

	g.SetManagerFunc(layout)
	bindClose(g)
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		logInfo(err.Error())
	}
}

func layout(g *gocui.Gui) error {
	if err := domainViewLayout(g); err != nil {
		return err
	} else if err := beanViewLayout(g); err != nil {
		return err
	} else if err := confViewLayout(g); err != nil {
		return err
	} else if err := logViewLayout(g); err != nil {
		return err
	}
	return nil
}

func bindClose(g *gocui.Gui) {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		panic(err.Error())
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func confViewLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("conf", maxX/2, 0, maxX, maxY*2/3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Config"
		confView = v
	}
	return nil
}

func domainViewLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("domains", 0, 0, maxX/2, maxY/3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Domains"
		v.Editable = true
		v.Wrap = true
		domainView = v
	}
	return nil
}

func beanViewLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("bean", 0, maxY/3, maxX/2, maxY*2/3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Beans"
		beanView = v
	}
	return nil
}

func logViewLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("log", 0, maxY*2/3, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Log"
		logView = v
		logView.Autoscroll = true
		logView.Editable = true
	}
	return nil
}

func scheduleRefresh() {
	time.Sleep(2 * time.Second)
	refreshViewContent()
	t := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-t.C:
			refreshViewContent()
		}
	}
}

func refreshViewContent() {
	logInfo("prepare to refresh content")
	if d, err := jmx.domains(); err == nil {
		domainView.Clear()
		fmt.Fprint(domainView, strings.Join(d, "\n"))
	} else {
		logInfo("refresh domain failed for %s", err.Error())
	}

	if b, err := jmx.beans(); err == nil {
		beanView.Clear()
		fmt.Fprint(beanView, strings.Join(b, "\n"))
	} else {
		logInfo("refresh beans failed for %s", err.Error())
	}

	if c, err := cfg.dump(); err == nil {
		confView.Clear()
		fmt.Fprint(confView, c)
	}
}

func logInfo(format string, args ...interface{}) (int, error) {
	return fmt.Fprintf(logView, format, args...)
}
