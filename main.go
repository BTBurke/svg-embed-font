package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

const (
	embedTemplate = `
	{{- range .Fonts}}
		<style type="text/css">
		<![CDATA[
		@font-face {
			font-family: '{{.Family}}';
			src: url('data:application/x-font-ttf;base64,{{.EncodedFont}}');
			}
		]]>
		</style>
	{{- end}}
	</defs>
	`
	usage = `
Usage:
svg-font-embed input.svg [font1.ttf font2.ttf]

Required Arguments:
input.svg - The SVG file to embed fonts within

Optional Arguments:
font.ttf - Specify one or more font files to embed within the SVG document.  Fonts do not have to be specified unless it is not obvious which file matches the fonts in the SVG file.

If no fonts are specified, the current directory and all subdirectories will be walked to look for a matching font file.  A match is defined as a font file which has the font-family name in its file name.  When multiple font files exists that would match (such as when the font comes in different weights), an error will be thrown unless you specify which font file to use on the command line.
`
)

func main() {
	if (len(os.Args) < 2) || (os.Args[1] == "help") || (os.Args[1] == "--help") {
		fmt.Println(usage)
		os.Exit(1)
	}
	svgFile := os.Args[1]

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("Could not get current working directory")
	}

	if _, err := os.Stat(path.Join(wd, svgFile)); os.IsNotExist(err) {
		fmt.Printf("Error: Could not find SVG file %s in current directory\n", svgFile)
		os.Exit(1)
	}

	if !strings.HasSuffix(os.Args[1], "svg") {
		fmt.Printf("Error: Input file %s does not end with a .svg extension\n", os.Args[1])
		os.Exit(1)
	}
	svg, err := ioutil.ReadFile(path.Join(wd, os.Args[1]))

	svgEmbed, err := FindEmbedFonts(svg, wd)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	svgEmbedFileName := strings.Replace(svgFile, ".svg", ".embed.svg", 1)
	if err := ioutil.WriteFile(svgEmbedFileName, svgEmbed, 0644); err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}
	fmt.Printf("Output saved: %s\n", svgEmbedFileName)

}
