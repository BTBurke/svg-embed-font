package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// Font is a single font family and associated Base64 encoded font file.
type Font struct {
	Family string
	//CmdLineSet is set when the user specifies a particular font file on the command line
	CmdLineSet  bool
	EncodedFont string
	File        string
}

// Document represents an SVG document containing one or more fonts
type Document struct {
	Fonts []Font
}

// Add will add a new font definition to the document
func (d *Document) Add(f Font) {
	d.Fonts = append(d.Fonts, f)
	return
}

// FontMap is an alias type for a map of the font-family to a font specification
type FontMap map[string]Font

// FindEmbedFonts will analyze the SVG file looking for all unique fonts used.  It will then walk the directory tree
// starting from the working directory looking for fonts that match the font-family names.  It will then Base64 encode
// the font and embed it in the SVG file.
func FindEmbedFonts(svg []byte, dir string) ([]byte, error) {
	fonts := make(FontMap)
	re := regexp.MustCompile("font-family:(.*?);")
	matches := re.FindAllStringSubmatch(string(svg), -1)
	for _, match := range matches {
		name := strings.Trim(match[1], " '\"\t\n\r")
		if (name == "sans-serif") || (name == "serif") {
			continue
		}
		fonts[name] = Font{Family: name}
	}
	if len(os.Args) > 2 {
		err := ProcessCmdLineFonts(fonts, os.Args[2:], dir)
		if err != nil {
			return svg, fmt.Errorf("Error processing command line font: %s", err)
		}
	}
	if CheckAllFontsSet(fonts) {
		svgEmbed, err := Embed(fonts, svg)
		if err != nil {
			return svg, err
		}
		PrintResults(fonts)
		return svgEmbed, nil
	}
	if err := Walk(fonts, dir); err != nil {
		return svg, err
	}
	if !CheckAllFontsSet(fonts) {
		for family, f := range fonts {
			if len(f.EncodedFont) == 0 {
				return svg, fmt.Errorf("No matching font file found for font-family: %s", family)
			}
		}
	}
	svgEmbed, err := Embed(fonts, svg)
	if err != nil {
		return svg, err
	}
	PrintResults(fonts)
	return svgEmbed, nil
}

// ProcessCmdLineFonts will embed any fonts specified on the command line in the SVG before looking for others
// in the file system
func ProcessCmdLineFonts(fm FontMap, fonts []string, baseDir string) error {
	for _, fontFile := range fonts {
		p := path.Join(baseDir, fontFile)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return err
		}
		for family := range fm {
			familyT := strings.Replace(family, " ", "", -1)
			if strings.Contains(strings.ToLower(fontFile), strings.ToLower(familyT)) {
				data, err := ioutil.ReadFile(p)
				if err != nil {
					return err
				}
				fm[family] = Font{
					Family:      family,
					CmdLineSet:  true,
					File:        fontFile,
					EncodedFont: base64.StdEncoding.EncodeToString(data),
				}
			}
		}
	}
	return nil
}

// CheckAllFontsSet will determine if all fonts have been resolved to associated font files
func CheckAllFontsSet(fm FontMap) bool {
	for _, f := range fm {
		if len(f.EncodedFont) == 0 {
			return false
		}
	}
	return true
}

// Embed puts the encoded fonts within the <defs> section of the SVG
func Embed(fm FontMap, svg []byte) ([]byte, error) {
	doc := new(Document)
	for family := range fm {
		doc.Add(fm[family])
	}
	t := template.Must(template.New("embed").Parse(embedTemplate))

	buf := new(bytes.Buffer)
	if err := t.Execute(buf, doc); err != nil {
		return svg, err
	}
	svgEmbed := strings.Replace(string(svg), "</defs>", buf.String(), 1)
	return []byte(svgEmbed), nil
}

func genWalkFunc(fm FontMap) filepath.WalkFunc {
	return func(p string, info os.FileInfo, err error) error {
		if (info.IsDir()) || (err != nil) {
			return nil
		}
		_, f := path.Split(p)
		for family := range fm {
			familyT := strings.Replace(family, " ", "", -1)
			if strings.Contains(strings.ToLower(f), strings.ToLower(familyT)) {
				if fm[family].CmdLineSet {
					return nil
				}
				data, err := ioutil.ReadFile(p)
				if err != nil {
					return err
				}
				if len(fm[family].EncodedFont) > 0 {
					return fmt.Errorf("Multiple font files found as a match to font-family: %s (%s, %s)", family, fm[family].File, f)
				}
				fm[family] = Font{
					Family:      family,
					CmdLineSet:  true,
					File:        f,
					EncodedFont: base64.StdEncoding.EncodeToString(data),
				}
			}
		}
		return nil
	}
}

// Walk walks the file system looking for font files that match the font-families in the document
func Walk(fm FontMap, baseDir string) error {
	if err := filepath.Walk(baseDir, genWalkFunc(fm)); err != nil {
		return err
	}
	return nil
}

//PrintResults generates a report
func PrintResults(fm FontMap) {
	fmt.Printf("Found %d fonts to be embedded.  Using the following font files:\n", len(fm))
	for family, f := range fm {
		fmt.Printf("%s: %s\n", family, f.File)
	}
	return
}
