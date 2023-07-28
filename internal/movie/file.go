package movie

import (
	"bufio"
	"bytes"
	"io"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gabe565/ascii-movie/internal/progressbar"
)

func (m *Movie) LoadFile(path string, src io.Reader, speed float64) error {
	m.Filename = filepath.Base(path)

	var f Frame
	var buf bytes.Buffer
	scanner := bufio.NewScanner(src)

	// Build part of every frame, excluding progress bar and bottom padding
	frameNum := -1
	frameHeadRe := regexp.MustCompile(`^\d+$`)
	for scanner.Scan() {
		if frameHeadRe.Match(scanner.Bytes()) {
			frameNum += 1
			if frameNum != 0 {
				f.Data = buf.String()
				buf.Reset()
				m.Frames = append(m.Frames, f)
			}

			f = Frame{}

			v, err := strconv.Atoi(scanner.Text())
			if err != nil {
				return err
			}

			f.Duration = time.Duration(v) * time.Second / 15
			f.Duration = time.Duration(float64(f.Duration) / speed)
		} else {
			if len(scanner.Bytes()) > m.Width {
				m.Width = len(scanner.Bytes())
			}
			buf.WriteString(scanner.Text() + "\n")
		}
	}
	m.Frames = append(m.Frames, f)
	if err := scanner.Err(); err != nil {
		return err
	}

	// Compute the total duration
	var frameCap int
	bar := progressbar.New()
	totalDuration := m.Duration()

	// Build the rest of every frame and write to disk
	var currentPosition time.Duration
	for i, f := range m.Frames {
		f.Data = strings.TrimSuffix(f.Data, "\n")
		f.Progress = bar.Generate(currentPosition+f.Duration/2, totalDuration, m.Width+2)
		m.Frames[i] = f
		if frameCap < len(f.Data) {
			frameCap = len(f.Data)
		}
		currentPosition += f.Duration
	}

	m.Cap = frameCap
	m.screenStyle = screenStyle.Copy().Width(m.Width)

	return nil
}
