package fixtures

import (
	"strings"
	"unicode/utf8"
)

type ansiParserState uint8

const (
	ansiText ansiParserState = iota
	ansiEscape
	ansiCSI
	ansiOSC
	ansiOSCEscape
)

type ansiScreen struct {
	width  int
	height int
	rows   [][]byte
	row    int
	col    int

	state        ansiParserState
	params       [2]int
	paramPresent [2]bool
	paramIndex   int
	paramInvalid bool
}

func newANSIScreen(width, height int) *ansiScreen {
	return &ansiScreen{
		width:  width,
		height: height,
		rows:   [][]byte{nil},
	}
}

func (s *ansiScreen) Write(data []byte) {
	for _, b := range data {
		s.writeByte(b)
	}
}

func (s *ansiScreen) writeByte(b byte) {
	switch s.state {
	case ansiEscape:
		s.writeEscape(b)
	case ansiCSI:
		s.writeCSI(b)
	case ansiOSC:
		switch b {
		case 0x07:
			s.state = ansiText
		case 0x1b:
			s.state = ansiOSCEscape
		}
	case ansiOSCEscape:
		switch b {
		case '\\':
			s.state = ansiText
		case 0x1b:
			s.state = ansiOSCEscape
		default:
			s.state = ansiOSC
		}
	default:
		s.writeText(b)
	}
}

func (s *ansiScreen) writeEscape(b byte) {
	switch b {
	case '[':
		s.resetCSI()
		s.state = ansiCSI
	case ']':
		s.state = ansiOSC
	case 0x1b:
		s.state = ansiEscape
	default:
		s.state = ansiText
	}
}

func (s *ansiScreen) writeCSI(b byte) {
	switch {
	case b >= '0' && b <= '9':
		if s.paramIndex < len(s.params) {
			s.paramPresent[s.paramIndex] = true
			s.params[s.paramIndex] = s.params[s.paramIndex]*10 + int(b-'0')
		}
	case b == ';':
		s.paramIndex++
	case b == '?':
	case b == ' ':
		s.paramInvalid = true
	default:
		s.executeCSI(b)
		s.state = ansiText
	}
}

func (s *ansiScreen) resetCSI() {
	s.params = [2]int{}
	s.paramPresent = [2]bool{}
	s.paramIndex = 0
	s.paramInvalid = false
}

func (s *ansiScreen) param(index, defaultValue int) int {
	if s.paramInvalid || index >= len(s.params) || !s.paramPresent[index] {
		return defaultValue
	}
	return s.params[index]
}

func (s *ansiScreen) executeCSI(final byte) {
	n := s.param(0, 1)
	switch final {
	case 'A':
		s.row -= n
		if s.row < 0 {
			s.row = 0
		}
	case 'B':
		s.moveCursorDown(n)
	case 'E':
		s.moveCursorDown(n)
		s.col = 0
	case 'F':
		s.row -= n
		if s.row < 0 {
			s.row = 0
		}
		s.col = 0
	case 'G':
		s.col = n - 1
		s.clampCol()
	case 'H', 'f':
		s.row = n - 1
		s.col = s.param(1, 1) - 1
		if s.row < 0 {
			s.row = 0
		}
		if s.height > 0 && s.row >= s.height {
			s.row = s.height - 1
		}
		s.clampCol()
		s.ensureRow(s.row)
	case 'K':
		s.eraseLine(s.param(0, 0))
	case 'J':
		if s.param(0, 0) == 2 {
			s.rows = [][]byte{nil}
			s.row = 0
			s.col = 0
		}
	}
}

func (s *ansiScreen) writeText(b byte) {
	switch b {
	case 0x1b:
		s.state = ansiEscape
	case '\r':
		s.col = 0
	case '\n':
		s.lineFeed()
		s.col = 0
	case '\b':
		if s.col > 0 {
			s.col--
		}
	case '\t':
		next := (s.col/8 + 1) * 8
		for s.col < next {
			s.writePrintable(' ')
		}
	default:
		if b >= 0x20 {
			s.writePrintable(b)
		}
	}
}

func (s *ansiScreen) writePrintable(b byte) {
	if s.width > 0 && s.col >= s.width {
		s.lineFeed()
		s.col = 0
	}
	s.ensureCol(s.row, s.col)
	if s.col < len(s.rows[s.row]) {
		s.rows[s.row][s.col] = b
	} else {
		s.rows[s.row] = append(s.rows[s.row], b)
	}
	s.col++
}

func (s *ansiScreen) lineFeed() {
	if s.height > 0 && s.row >= s.height-1 {
		if len(s.rows) < s.height {
			s.rows = append(s.rows, nil)
			s.row++
			return
		}
		copy(s.rows, s.rows[1:])
		s.rows[len(s.rows)-1] = nil
		s.row = s.height - 1
		return
	}
	s.row++
	s.ensureRow(s.row)
}

func (s *ansiScreen) moveCursorDown(n int) {
	s.row += n
	if s.height > 0 && s.row >= s.height {
		s.row = s.height - 1
	}
	s.ensureRow(s.row)
}

func (s *ansiScreen) ensureRow(row int) {
	for row >= len(s.rows) {
		s.rows = append(s.rows, nil)
	}
}

func (s *ansiScreen) ensureCol(row, col int) {
	s.ensureRow(row)
	for len(s.rows[row]) < col {
		s.rows[row] = append(s.rows[row], ' ')
	}
}

func (s *ansiScreen) clampCol() {
	if s.col < 0 {
		s.col = 0
	}
	if s.width > 0 && s.col >= s.width {
		s.col = s.width - 1
	}
}

func (s *ansiScreen) eraseLine(mode int) {
	s.ensureRow(s.row)
	switch mode {
	case 0:
		if s.col < len(s.rows[s.row]) {
			s.rows[s.row] = s.rows[s.row][:s.col]
		}
	case 1:
		s.ensureCol(s.row, s.col)
		for i := 0; i < s.col && i < len(s.rows[s.row]); i++ {
			s.rows[s.row][i] = ' '
		}
	case 2:
		s.rows[s.row] = nil
	}
}

func (s *ansiScreen) String() string {
	var output strings.Builder
	for i, row := range s.rows {
		output.Write(row)
		if i < len(s.rows)-1 {
			output.WriteByte('\n')
		}
	}
	return output.String()
}

func (s *ansiScreen) MaxLineWidth() int {
	maxWidth := 0
	for _, row := range s.rows {
		if width := utf8.RuneCount(row); width > maxWidth {
			maxWidth = width
		}
	}
	return maxWidth
}
