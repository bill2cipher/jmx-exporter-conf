package main

import (
	"fmt"
	"strings"

	"github.com/jroimartin/gocui"
)

func (v *View) bindActions(g *gocui.Gui) {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, v.quit); err != nil {
		panic(err.Error())
	}

	if err := g.SetKeybinding("", 'q', gocui.ModNone, v.quit); err != nil {
		panic(err.Error())
	}

	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, v.nextView); err != nil {
		panic(err.Error())
	}

	if err := g.SetKeybinding("bean", 'j', gocui.ModNone, v.beanCursorDown); err != nil {
		panic(err.Error())
	}

	if err := g.SetKeybinding("bean", 'k', gocui.ModNone, v.beanCursorUp); err != nil {
		panic(err.Error())
	}

	if err := g.SetKeybinding("domain", 'j', gocui.ModNone, v.domainCusorDown); err != nil {
		panic(err.Error())
	}

	if err := g.SetKeybinding("domain", 'k', gocui.ModNone, v.domainCusorUp); err != nil {
		panic(err.Error())
	}

	if err := g.SetKeybinding("bean", gocui.KeyEnter, gocui.ModNone, v.selectBean); err != nil {
		panic(err.Error())
	}

	if err := g.SetKeybinding("", 's', gocui.ModNone, v.saveConf); err != nil {
		panic(err.Error())
	}
}

func (v *View) saveConf(g *gocui.Gui, view *gocui.View) error {
	err := v.cfg.save()
	if err != nil {
		v.logInfo("save cfg to clipboard failed for: %s", err.Error())
	} else {
		v.logInfo("save cfg to clipboard success")
	}
	return nil
}

func (v *View) domainCusorDown(g *gocui.Gui, view *gocui.View) error {
	x, y := view.Cursor()
	_, ylimit := view.Size()
	y++
	if y >= ylimit && v.domainIdx < len(v.domains)-ylimit {
		v.domainIdx++
		v.refreshDomain()
		return nil
	} else if y >= ylimit {
		v.logInfo("reach domain bottom")
		return nil
	} else if err := view.SetCursor(x, y); err != nil {
		return err
	} else {
		v.refreshDomain()
		return nil
	}
}

func (v *View) domainCusorUp(g *gocui.Gui, view *gocui.View) error {
	x, y := view.Cursor()
	y--
	if y < 0 && v.domainIdx > 0 {
		v.domainIdx--
		v.refreshDomain()
		return nil
	} else if y < 0 {
		v.logInfo("reach domain top")
		return nil
	} else if err := view.SetCursor(x, y); err != nil {
		return err
	} else {
		v.refreshDomain()
		return nil
	}
}

func (v *View) refreshDomain() {
	var result []string
	for _, k := range v.domainList[v.domainIdx:] {
		result = append(result, k)
	}
	v.domainView.Clear()
	fmt.Fprintln(v.domainView, strings.Join(result, "\n"))

	domainName, err := v.currentDomainName()
	if err != nil {
		return
	}

	if domainName == "" {
		return
	} else if d, exist := v.domains[domainName]; exist {
		v.currentDomain = d
		v.logInfo("select domain %s with bean size %d", domainName, len(v.currentDomain.beans))
		v.selectDomain()
	}
}

func (v *View) currentDomainName() (string, error) {
	_, y := v.domainView.Cursor()
	return v.domainView.Line(y)
}

func (v *View) selectDomain() error {
	v.currentDomain.beanIdx = 0
	v.beanView.SetCursor(0, 0)
	v.refreshBean()
	return nil
}

func (v *View) selectBean(g *gocui.Gui, view *gocui.View) error {
	_, y := view.Cursor()
	beanName, err := view.Line(y)
	if err != nil {
		return err
	}
	if beanName == "" {
		return nil
	}
	v.logInfo("select mbean %s", beanName)
	domainName, _ := v.currentDomainName()
	v.cfg.addRule(fmt.Sprintf("%s:%s", domainName, beanName))
	v.toggleBeanUsed(beanName)

	v.refreshConf()
	v.refreshBean()
	return nil
}

func (v *View) toggleBeanUsed(beanName string) {
	if v.currentDomain == nil {
		return
	}
	for _, b := range v.currentDomain.beans {
		if b.name == beanName {
			b.used = !b.used
		}
	}
}

func (v *View) beanCursorUp(g *gocui.Gui, view *gocui.View) error {
	if v.currentDomain == nil {
		return nil
	}

	x, y := view.Cursor()
	y--
	if y < 0 && v.currentDomain.beanIdx > 0 {
		v.currentDomain.beanIdx--
		v.refreshBean()
		v.logInfo("bean idx %d", v.currentDomain.beanIdx)
		return nil
	} else if y < 0 {
		v.logInfo("reach top now!!!!")
		return nil
	}
	return view.SetCursor(x, y)
}

func (v *View) beanCursorDown(g *gocui.Gui, view *gocui.View) error {
	if v.currentDomain == nil {
		return nil
	}

	x, y := view.Cursor()
	_, ylimit := view.Size()
	y++
	if y >= ylimit && v.currentDomain.beanIdx < len(v.currentDomain.beans)-ylimit {
		v.currentDomain.beanIdx++
		v.refreshBean()
		return nil
	} else if y >= ylimit {
		v.logInfo("reach bottom now!!!!")
		return nil
	}
	return view.SetCursor(x, y)
}

func (v *View) quit(g *gocui.Gui, view *gocui.View) error {
	return gocui.ErrQuit
}

func (v *View) nextView(g *gocui.Gui, view *gocui.View) error {
	nextIndex := (v.active + 1) % 2
	name := ""
	if nextIndex == 0 {
		name = "domain"
	} else if nextIndex == 1 {
		name = "bean"
	}

	if _, err := v.setCurrentViewOnTop(g, name); err != nil {
		return err
	}
	v.active = nextIndex
	return nil
}

func (v *View) setCurrentViewOnTop(g *gocui.Gui, name string) (*gocui.View, error) {
	if v, err := g.SetCurrentView(name); err != nil {
		return nil, err
	} else {
		return v, nil
	}
}

func (v *View) refreshViewContent() {
	v.logInfo("refresh content")
	v.refreshDomain()
	v.refreshBean()
	v.refreshConf()
}

func (v *View) refreshBean() {
	if v.currentDomain == nil {
		return
	}
	var result []string
	domain := v.currentDomain
	for _, b := range domain.beans[domain.beanIdx:] {
		if b.used {
			result = append(result, fmt.Sprintf("\033[35;7m%s\033[0m", b.name))
		} else {
			result = append(result, b.name)
		}
	}
	v.beanView.Clear()
	fmt.Fprint(v.beanView, strings.Join(result, "\n"))
}

func (v *View) refreshConf() {
	if c, err := v.cfg.dump(); err == nil {
		v.confView.Clear()
		fmt.Fprint(v.confView, c)
	}
}
