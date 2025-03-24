package conv

import (
	"bufio"
	"fmt"
	"io"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type Converter struct {
	r *transform.Reader
	enc encoding.Encoding
	toUtf8 bool
	err error
}

func (c *Converter) Scan() bool {
	return c.scanner.Scan()
}

func (c *Converter) Text() string {
	transform.NewReader()
}

func NewConverter( enc encoding.Encoding, r io.Reader ) (*Converter, error) {
	return &Converter{ s: bufio.NewReader(r), enc: enc }, nil
}

func getEncoding(enc string) (encoding.Encoding, error) {
	switch enc{
		case "sjis":
			return japanese.ShiftJIS, nil
		case "euc":
			return japanese.EUCJP, nil
		default:
			return nil, fmt.Errorf("ERROR")
	}
}