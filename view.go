package main

import (
	"fmt"
	"time"

	"github.com/jroimartin/gocui"
)

type ViewLabel struct {
	name, value string
	index       int
	used        bool
}

type ViewBean struct {
	domain, name     string
	labels           []*ViewLabel
	valueName, value string
	used             bool
}

type ViewDomain struct {
	beanIdx int
	beans   []*ViewBean
}

type View struct {
	jmx    *JMX
	cfg    *Conf
	inited bool

	domainView *gocui.View
	beanView   *gocui.View
	labelView  *gocui.View
	confView   *gocui.View
	logView    *gocui.View
	active     int

	domains       map[string]*ViewDomain
	domainList    []string
	domainIdx     int
	currentDomain *ViewDomain
	currentBean   *ViewBean
}

var (
	v *View
)

func Start(jmx *JMX, cfg *Conf) {
	v = &View{
		jmx:     jmx,
		cfg:     cfg,
		inited:  false,
		domains: make(map[string]*ViewDomain),
	}
	v.init()

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		v.logInfo(err.Error())
	}
	defer g.Close()
	g.Highlight = true
	g.Cursor = true
	g.SelFgColor = gocui.ColorGreen
	g.SetManagerFunc(v.layout)

	v.bindActions(g)
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		panic(err.Error())
	}
}

func (v *View) init() {
	for _, d := range v.jmx.Domains() {
		v.domains[d] = &ViewDomain{beanIdx: 0}
	}

	for d := range v.domains {
		beans := v.jmx.Beans(d)
		v.domainList = append(v.domainList, d)
		for _, b := range beans {
			v.domains[d].beans = append(v.domains[d].beans, v.buildViewBean(b))
		}
	}
}

func (v *View) buildViewBean(bean *JMXBean) *ViewBean {
	var labels []*ViewLabel
	for _, l := range bean.Labels {
		labels = append(labels, &ViewLabel{
			name:  l.Name,
			value: l.Value,
			index: l.Index,
			used:  false,
		})
	}
	return &ViewBean{
		domain:    bean.Domain,
		name:      bean.Name,
		valueName: bean.ValueName,
		value:     bean.Value,
		used:      false,
		labels:    labels,
	}
}

func (v *View) layout(g *gocui.Gui) error {
	if err := v.domainViewLayout(g); err != nil {
		return err
	} else if err := v.beanViewLayout(g); err != nil {
		return err
	} else if err := v.confViewLayout(g); err != nil {
		return err
	} else if err := v.logViewLayout(g); err != nil {
		return err
	} else if err := v.labelViewLayout(g); err != nil {
		return err
	}
	if !v.inited {
		v.inited = true
		v.refreshViewContent()
	}
	return nil
}

func (v *View) domainViewLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if view, err := g.SetView("domain", 0, 0, maxX/2, maxY/3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		view.Title = "Domains"
		view.Highlight = true
		view.SelBgColor = gocui.ColorCyan
		view.SelFgColor = gocui.ColorMagenta
		v.domainView = view
		v.setCurrentViewOnTop(g, "domain")
	}
	return nil
}

func (v *View) beanViewLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if view, err := g.SetView("bean", 0, maxY/3, maxX/2, maxY*2/3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		view.Title = "Beans"
		view.SelBgColor = gocui.ColorCyan
		view.SelFgColor = gocui.ColorMagenta
		v.beanView = view
	}
	return nil
}

func (v *View) labelViewLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if view, err := g.SetView("label", 0, maxY*2/3, maxX/2, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		view.Title = "Lables"
		view.SelBgColor = gocui.ColorCyan
		view.SelFgColor = gocui.ColorMagenta
		v.labelView = view
	}
	return nil
}

func (v *View) confViewLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if view, err := g.SetView("conf", maxX/2, 0, maxX-1, maxY/3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		view.Title = "Config"
		v.confView = view
	}
	return nil
}

func (v *View) logViewLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if view, err := g.SetView("log", maxX/2, maxY/3, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		view.Title = "Log"
		view.Autoscroll = true
		view.Wrap = true
		v.logView = view
	}
	return nil
}

func (v *View) logInfo(format string, args ...interface{}) (int, error) {
	mesg := fmt.Sprintf(time.Now().String()+"  "+format+"\n", args...)
	return fmt.Fprintln(v.logView, mesg)
}
