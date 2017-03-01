Embed Fonts in SVG Assets
===

SVG is useful for device-independent resolution, but can often be a pain because fonts must be embedded in the file to render properly across all browsers.  `svg-embed-font` is a command line tool to easily determine what fonts are used in an SVG file and encode them as Base64 assets within it.

If your SVG assets look great on your computer and messed up on everyone else's, it's because the fonts aren't embedded properly in the file.

### Usage

```
svg-embed-font input.svg
```

In the default mode, the tool will scan the SVG file for all font-family declarations then attempt to locate matching font files (any font file format).  Matches are defined as a case-insensitive substring match for the font family name ignoring any spaces.  So if you declare:

```
font-family: 'Permanent Marker'

Matches:
permanentmarker.ttf
PermanentMarker-700.otf
```

In this case, there are two possible matches, which can often happen when a font comes in multiple weights.  To specify which should be used, list the font on the command line after the input file.  Multiple possible matches must be resolved by listing the correct one on the command line.

```
svg-embed-font input.svg permanentmarker.ttf
```

One or more preferred font files can be listed on the command line and it will use those files instead of any other matches it finds.

### Font File Path Search

If you don't specify the exact font files, it will look in the current directory and all subdirectories for a match, so you can lay out your files in a logical hierarchy and it will find them.  If it exhausts all possible files without finding a match to every font in the SVG file, it will return an error.

### Installation

Download the release appropriate for your operating system on the [releases page](https://github.com/BTBurke/svg-embed-font/releases).

### License

MIT
