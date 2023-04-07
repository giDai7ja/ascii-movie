package main

import (
	"bytes"
	"fmt"
	"github.com/gabe565/ascii-movie/config"
	"github.com/gabe565/ascii-movie/internal/frame"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
)

var frameTemplate *template.Template

func init() {
	var err error

	if frameTemplate, err = template.New("").Parse(frameTmpl); err != nil {
		panic(err)
	}
}

func writeFrame(f frame.Frame) error {
	filename := filepath.Join(config.OutputDir, fmt.Sprintf("frame%d.go", f.Num))

	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		_ = out.Close()
	}(out)

	var buf bytes.Buffer

	err = frameTemplate.Execute(&buf, map[string]any{
		"Package": config.OutputDir,
		"Frame":   f,
	})
	if err != nil {
		return err
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}

	if _, err := out.Write(formatted); err != nil {
		return err
	}

	if err := out.Close(); err != nil {
		return err
	}

	return nil
}

func writeFrameList(frames []frame.Frame, cap int) error {
	filename := filepath.Join(config.OutputDir, "0_frame_list.go")

	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		_ = out.Close()
	}(out)

	tmpl, err := template.New("").Parse(allTmpl)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	err = tmpl.Execute(&buf, map[string]any{
		"Package": config.OutputDir,
		"Frames":  frames,
		"Cap":     cap,
	})
	if err != nil {
		return err
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}

	if _, err := out.Write(formatted); err != nil {
		return err
	}

	if err := out.Close(); err != nil {
		return err
	}

	return nil
}
