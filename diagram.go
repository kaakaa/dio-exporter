package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Diagram struct {
	Path     string
	RelPath  string
	Elements []DiagramElement
	XMLBody  string
}
type DiagramElement struct {
	Id         string `xml:"id,attr"`
	Name       string `xml:"name,attr`
	GraphModel struct {
		Width  string `xml:"dx"`
		Height string `xml:"dy"`
		Scale  string `xml:"pageScale"`
	} `xml:"mxGraphModel"`
}

type ConvertData struct {
	XMLString       string
	Format          string
	Width           int
	Height          int
	Border          int
	BackgroundColor string
	// From            int
	// To              int
	PageId string
	Scale  float64
	Extras string
}

// ToRenderparams returns JavaScript styled JSON String that all key aren't enclosed by double quote
func (d *ConvertData) ToRenderParams() string {
	return fmt.Sprintf(
		`{format:'%s', w:%d, h:%d, border:%d, bg: '%s', pageId:'%s', scale:%f, extras: '%s', xml:'%s'}`,
		d.Format,
		d.Width,
		d.Height,
		d.Border,
		d.BackgroundColor,
		d.PageId,
		d.Scale,
		d.Extras,
		d.XMLString,
	)
}

func (d *Diagram) Export(addr, format, bgColor, dest string) error {
	for idx, tab := range d.Elements {
		pageWidth, _ := strconv.Atoi(tab.GraphModel.Width)
		pageHeight, _ := strconv.Atoi(tab.GraphModel.Height)
		scale, _ := strconv.ParseFloat(tab.GraphModel.Scale, 64)
		data := ConvertData{
			XMLString:       d.XMLBody,
			Format:          format,
			Width:           pageWidth,
			Height:          pageHeight,
			Border:          1,
			BackgroundColor: bgColor,
			PageId:          tab.Id,
			Scale:           scale,
			Extras:          "",
		}

		// drawio url for exporting (such as `http://drawio_host/export3.html`)
		u := url.URL{
			Scheme: "http",
			Host:   addr,
			Path:   "export3.html",
		}
		img, err := capture(u.String(), data)
		if err != nil {
			logf("[WARN] failed to capture image: %v", err)
			continue
		}

		// save an image
		dummy := filepath.Join(dest, d.RelPath)
		ext := filepath.Ext(dummy)
		suffix := tab.Name
		if len(suffix) == 0 {
			suffix = strconv.Itoa(idx)
		}
		path := fmt.Sprintf("%s_%s.%s", dummy[0:len(dummy)-len(ext)], suffix, format)

		logf("  Save exported diagram to [%s]", path)
		if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			return err
		}
		if err := ioutil.WriteFile(path, img, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func ReadDir(root string, extList []string) ([]*Diagram, error) {
	var walker = &DiagramWalker{}
	err := filepath.Walk(root, walker.getWalker(root, extList))
	if err != nil {
		return nil, err
	}
	logf("Read %d diagrams", len(walker.Diagrams))
	if *debug {
		for _, d := range walker.Diagrams {
			debugf("  - %s", d.Path)
		}
	}
	return walker.Diagrams, nil
}

type DiagramWalker struct {
	Diagrams []*Diagram
}

func (dw *DiagramWalker) getWalker(root string, extList []string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if include(filepath.Ext(path), extList) {
			d, err := readDiagram(root, path)
			if err != nil {
				logf("[WARN] failed to read diagramfile: %v", err)
			} else {
				dw.Diagrams = append(dw.Diagrams, d)
			}
		}
		return nil
	}
}

func readDiagram(root, path string) (*Diagram, error) {
	debugf("Start to read %s", path)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var xmldata struct {
		Diagrams []DiagramElement `xml:"diagram"`
	}
	if err := xml.Unmarshal(b, &xmldata); err != nil {
		return nil, err
	}

	abs, _ := filepath.Abs(path)
	rel, _ := filepath.Rel(root, path)
	return &Diagram{
		Path:     abs,
		RelPath:  rel,
		Elements: xmldata.Diagrams,
		XMLBody:  strings.Replace(string(b), "\n", "", -1),
	}, nil
}

func include(s string, list []string) bool {
	for _, v := range list {
		if s == v {
			return true
		}
	}
	return false
}
