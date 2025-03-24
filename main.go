package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// getEncoding はエンコーディング名に対応する変換器を返す
func getEncoding(name string) (encoding.Encoding, error) {
	switch name {
	case "utf8":
		return unicode.UTF8, nil
	case "sjis", "shiftjis":
		return japanese.ShiftJIS, nil
	case "eucjp":
		return japanese.EUCJP, nil
	default:
		return nil, fmt.Errorf("未対応のエンコーディング: %s", name)
	}
}

// replaceUnsupportedCharacters は Shift_JIS に変換できない文字を適切に置換
func replaceUnsupportedCharacters(r rune) rune {
	replacements := map[rune]rune{
		'～': '〜', '‐': '－', '−': '－', // 波ダッシュ・ハイフン・マイナス
		'¥': '￥',                        // 国際YEN記号
		'“': '"', '”': '"', '„': '"', // ダブルクォート
		'‘': '\'', '’': '\'', '‚': '\'', // シングルクォート
	}

	var ret rune

	if newVal, ok := replacements[r]; ok {
		ret = newVal // 置換文字を追加
		fmt.Println(newVal)
	} else if r < 0x20 || r > 0x7E && !utf8.ValidRune(r) { // 制御文字・サロゲートペア・不正な文字
		ret = '?'
	}else {
		ret = r
	}
	return ret
}

// SafeTransformer は変換できない文字を "?" に置換する Transformer
type SafeEncoder struct {
	transform.Transformer
}

// NewSafeTransformer は変換器をラップして SafeTransformer を返す
func NewSafeEncoder(enc encoding.Encoding) transform.Transformer {
	return SafeEncoder{ enc.NewEncoder() }
}

// Transform メソッドをオーバーライドし、変換できない文字を "?" に置換
func (se SafeEncoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	for len(src) > 0 {
		r, size := utf8.DecodeRune(src)
		if r == utf8.RuneError {
			break
		}

		replaceUnsupportedCharacters(r)

		buf := make([]byte, 1024) // 一時バッファ
		n, _, e := se.Transformer.Transform(buf, src[:size], true)
		if e != nil {
			// 変換失敗時は "?" を代わりに挿入
			if len(dst) > nDst {
				dst[nDst] = '?'
				nDst++
			}
		} else {
			// 正常に変換できたらコピー
			if len(dst) >= nDst+n {
				copy(dst[nDst:], buf[:n])
				nDst += n
			}
		}
		nSrc += size
		src = src[size:] // 次の文字へ
	}
	return nDst, nSrc, nil
}

func main() {
    flag.Usage = func() {
        fmt.Println("使用方法:")
        fmt.Println("echo '入力テキスト' | ./jpconv")
        flag.PrintDefaults()
    }
	toUtf8 := flag.Bool("d", false, "" )
    flag.Parse()

    info, err := os.Stdin.Stat()
    if err != nil {
        fmt.Fprintln(os.Stderr, "エラー:", err)
        return
    }

    if (info.Mode() & os.ModeCharDevice) != 0 {
        flag.Usage()
        return
    }

	if *toUtf8 {
		trans := japanese.ShiftJIS.NewDecoder()
		io.Copy(os.Stdout, transform.NewReader(os.Stdin, trans))
	} else {
		trans := NewSafeEncoder(japanese.ShiftJIS)
		io.Copy(os.Stdout, transform.NewReader(os.Stdin, trans))
	}

}
