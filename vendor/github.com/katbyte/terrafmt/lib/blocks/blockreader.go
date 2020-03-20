package blocks

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/katbyte/terrafmt/lib/common"
	"github.com/sirupsen/logrus"
)

type Reader struct {
	FileName string

	//io
	Reader io.Reader
	Writer io.Writer

	//stats
	LineCount        int // total lines processed
	LinesBlock       int // total block lines processed
	BlockCount       int // total blocks found
	BlockCurrentLine int // current block line count

	ErrorBlocks int

	ReadOnly bool

	//callbacks
	LineRead  func(*Reader, int, string) error
	BlockRead func(*Reader, int, string) error
}

func ReaderPassthrough(br *Reader, number int, line string) error {
	_, err := br.Writer.Write([]byte(line))
	return err
}

func ReaderIgnore(br *Reader, number int, line string) error {
	return nil
}

func IsStartLine(line string) bool {
	if strings.HasSuffix(line, "return fmt.Sprintf(`\n") { // acctest
		return true
	} else if strings.HasPrefix(line, "```hcl") { // documentation
		return true
	} else if strings.HasPrefix(line, "```tf") { // documentation
		return true
	}

	return false
}

func IsFinishLine(line string) bool {
	if line == "`)\n" { // acctest
		return true
	} else if strings.HasPrefix(line, "`,") { // acctest
		return true
	} else if strings.HasPrefix(line, "```") { // documentation
		return true
	}

	return false
}

func (br *Reader) DoTheThing(filename string) error {
	var buf *bytes.Buffer

	if filename != "" {
		br.FileName = filename
		common.Log.Debugf("opening src file %s", filename)
		fs, err := os.Open(filename) // For read access.
		if err != nil {
			return err
		}
		defer fs.Close()
		br.Reader = fs

		// for now write to buffer
		if !br.ReadOnly {
			buf = bytes.NewBuffer([]byte{})
			br.Writer = buf
		} else {
			br.Writer = ioutil.Discard
		}
	} else {
		br.FileName = "stdin"
		br.Reader = os.Stdin
		br.Writer = os.Stdout

		if br.ReadOnly {
			br.Writer = ioutil.Discard
		}
	}

	br.LineCount = 0
	br.BlockCount = 0
	s := bufio.NewScanner(br.Reader)
	for s.Scan() { // scan file
		br.LineCount += 1
		//br.CurrentLine = s.Text()+"\n"
		l := s.Text() + "\n"

		if err := br.LineRead(br, br.LineCount, l); err != nil {
			return fmt.Errorf("NB LineRead failed @ %s:%d for %s: %v", br.FileName, br.LineCount, l, err)
		}

		if IsStartLine(l) {
			block := ""
			br.BlockCurrentLine = 0
			br.BlockCount += 1

			for s.Scan() { // scan block
				br.LineCount += 1
				br.BlockCurrentLine += 1
				l2 := s.Text() + "\n"

				// make sure we don't run into another block
				if IsStartLine(l2) {
					// the end of current block must be malformed, so lets pass it through and log an error
					logrus.Errorf("block %d @ %s:%d failed to find end of block", br.BlockCount, br.FileName, br.LineCount-br.BlockCurrentLine)
					if err := ReaderPassthrough(br, br.LineCount, block); err != nil { // is this ok or should we loop with LineRead?
						return err
					}

					if err := br.LineRead(br, br.LineCount, l2); err != nil {
						return fmt.Errorf("NB LineRead failed @ %s#%d for %s: %v", br.FileName, br.LineCount, l, err)
					}

					block = ""
					br.BlockCount += 1
					continue
				}

				if IsFinishLine(l2) {
					br.LinesBlock += br.BlockCurrentLine

					// todo configure this behaviour with switch's
					if err := br.BlockRead(br, br.LineCount, block); err != nil {
						//for now ignore block errors and output unformatted
						br.ErrorBlocks += 1
						logrus.Errorf("block %d @ %s:%d failed to process with: %v", br.BlockCount, br.FileName, br.LineCount-br.BlockCurrentLine, err)
						if err := ReaderPassthrough(br, br.LineCount, block); err != nil {
							return err
						}
					}

					if err := br.LineRead(br, br.LineCount, l2); err != nil {
						return fmt.Errorf("NB LineRead failed @ %s:%d for %s: %v", br.FileName, br.LineCount, l2, err)
					}

					block = ""
					break
				} else {
					block += l2
				}
			}

			// ensure last block in the file was property handled
			if block != "" {
				//for each line { Lineread()?
				logrus.Errorf("block %d @ %s:%d failed to find end of block", br.BlockCount, br.FileName, br.LineCount-br.BlockCurrentLine)
				if err := ReaderPassthrough(br, br.LineCount, block); err != nil { // is this ok or should we loop with LineRead?
					return err
				}
			}
		}
	}

	// If not read-only, need to write back to file.
	if !br.ReadOnly {
		destination, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer destination.Close()

		common.Log.Debugf("copying..")
		_, err = io.Copy(destination, buf)
		return err
	}

	// todo should this be at the end of a command?
	//fmt.Fprintf(os.Stderr, c.Sprintf("\nFinished processing <cyan>%d</> lines <yellow>%d</> blocks!\n", br.LineCount, br.BlockCount))
	return nil
}
