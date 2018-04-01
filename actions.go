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

	if err := g.SetKeybinding("label", 'j', gocui.ModNone, v.labelCursorDown); err != nil {
		panic(err.Error())
	}

	if err := g.SetKeybinding("label", 'k', gocui.ModNone, v.labelCursorUp); err != nil {
		panic(err.Error())
	}

	if err := g.SetKeybinding("bean", gocui.KeyEnter, gocui.ModNone, v.selectBean); err != nil {
		panic(err.Error())
	}

	if err := g.SetKeybinding("label", gocui.KeyEnter, gocui.ModNone, v.selectLabel); err != nil {
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

func (v *View) domainCusorUp(g *gocui.Gui, view *gocui.View) error {
	if v.cursorUp(g, view, len(v.domains), &(v.domainIdx)) {
		v.refreshDomain()
	}
	return nil
}

func (v *View) domainCusorDown(g *gocui.Gui, view *gocui.View) error {
	if v.cursorDown(g, view, len(v.domains), &(v.domainIdx)) {
		v.refreshDomain()
	}
	return nil
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

func (v *View) currentBeanName() (string, error) {
	_, y := v.beanView.Cursor()
	return v.beanView.Line(y)
}

func (v *View) selectDomain() error {
	v.currentDomain.beanIdx = 0
	v.beanView.SetCursor(0, 0)
	v.refreshBean()
	return nil
}

func (v *View) selectLabel(g *gocui.Gui, view *gocui.View) error {
	_, y := view.Cursor()
	labelName, err := view.Line(y)
	if err != nil {
		return err
	}
	if labelName == "" {
		return nil
	}
	v.logInfo("select label %s", labelName)
	v.toggleLabelUsed(labelName)
	v.refreshLabel()
	v.refreshConf()
	return nil
}

func (v *View) toggleLabelUsed(labelName string) {
	if v.currentBean == nil {
		return
	}
	for _, l := range v.currentBean.labels {
		if l.name == labelName {
			l.used = !l.used
		}
	}
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
	v.toggleBeanUsed(beanName)

	v.cfg.addRule(v.currentBean)
	v.refreshConf()
	v.refreshBean()

	return nil
}

func (v *View) setCurrentBean(beanName string) {
	if v.currentDomain == nil {
		return
	}
	for _, b := range v.currentDomain.beans {
		if b.name == beanName {
			v.currentBean = b
		}
	}
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
	if v.cursorUp(g, view, len(v.currentDomain.beans), &(v.currentDomain.beanIdx)) {
		v.refreshBean()
	}
	return nil
}

func (v *View) beanCursorDown(g *gocui.Gui, view *gocui.View) error {
	if v.currentDomain == nil {
		return nil
	}

	if v.cursorDown(g, view, len(v.currentDomain.beans), &(v.currentDomain.beanIdx)) {
		v.refreshBean()
	}
	return nil
}

func (v *View) labelCursorDown(g *gocui.Gui, view *gocui.View) error {
	if v.currentBean == nil {
		return nil
	}

	if v.cursorDown(g, view, len(v.currentBean.labels), &(v.currentBean.labelIdx)) {
		v.refreshLabel()
	}
	return nil
}

func (v *View) labelCursorUp(g *gocui.Gui, view *gocui.View) error {
	if v.currentBean == nil {
		return nil
	}
	if v.cursorUp(g, view, len(v.currentBean.labels), &(v.currentBean.labelIdx)) {
		v.refreshLabel()
	}
	return nil
}

func (v *View) cursorUp(g *gocui.Gui, view *gocui.View, contentLen int, curIndex *int) bool {
	x, y := view.Cursor()
	y--
	if y < 0 && (*curIndex) > 0 {
		(*curIndex)--
		return true
	} else if y < 0 {
		v.logInfo("reach top")
		return false
	} else if err := view.SetCursor(x, y); err != nil {
		return false
	} else {
		return true
	}
}

func (v *View) cursorDown(g *gocui.Gui, view *gocui.View, contentLen int, curIndex *int) bool {
	x, y := view.Cursor()
	_, ylimit := view.Size()
	y++

	limit := ylimit
	if ylimit > contentLen {
		limit = contentLen
	}

	if y >= ylimit && (*curIndex) < contentLen-ylimit {
		(*curIndex)++
		return true
	} else if y >= limit {
		v.logInfo("reach bottom")
		return false
	} else if err := view.SetCursor(x, y); err != nil {
		return false
	} else {
		return true
	}
}

func (v *View) quit(g *gocui.Gui, view *gocui.View) error {
	return gocui.ErrQuit
}

func (v *View) nextView(g *gocui.Gui, view *gocui.View) error {
	nextIndex := (v.active + 1) % 3
	name := ""
	if nextIndex == 0 {
		name = "domain"
	} else if nextIndex == 1 {
		name = "bean"
	} else if nextIndex == 2 {
		name = "label"
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

func (v *View) refreshLabel() {
	if v.currentBean == nil {
		return
	}
	var result []string
	bean := v.currentBean
	for _, l := range bean.labels[bean.labelIdx:] {
		if l.used {
			result = append(result, fmt.Sprintf("\033[35;7m%s\033[0m", l.name))
		} else {
			result = append(result, l.name)
		}
	}
	v.labelView.Clear()
	fmt.Fprint(v.labelView, strings.Join(result, "\n"))
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

	beanName, _ := v.currentBeanName()
	v.setCurrentBean(beanName)

	v.labelView.SetCursor(0, 0)
	v.currentBean.labelIdx = 0
	v.refreshLabel()
}

func (v *View) refreshConf() {
	if c, err := v.cfg.dump(); err == nil {
		v.confView.Clear()
		fmt.Fprint(v.confView, c)
	}
}
