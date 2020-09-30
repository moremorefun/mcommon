package mcommon

import (
	"image"
	"image/color"
	"io/ioutil"
	"strconv"
	"strings"

	"golang.org/x/image/math/fixed"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

func getFontFamilyFont(fontPath string) (*truetype.Font, error) {
	// 这里需要读取中文字体，否则中文文字会变成方格
	fontBytes, err := ioutil.ReadFile(fontPath)
	if err != nil {
		return &truetype.Font{}, err
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return &truetype.Font{}, err
	}
	return f, err
}

// GetStringSize 获取制定字体文字大小
func GetStringSize(str string, fontPath string, fontSize float64, dpi float64) (fixed.Rectangle26_6, fixed.Int26_6, error) {
	b, err := ioutil.ReadFile(fontPath)
	if err != nil {
		return fixed.Rectangle26_6{}, 0, err
	}
	f, err := truetype.Parse(b)
	if err != nil {
		return fixed.Rectangle26_6{}, 0, err
	}
	// Truetype stuff
	opts := truetype.Options{
		Size: fontSize,
		DPI:  dpi,
	}
	face := truetype.NewFace(f, &opts)
	rb, ra := font.BoundString(face, str)
	return rb, ra, nil
}

// GetWriteSizeFont 获取指定字体的文字大小
func GetWriteSizeFont(content string, fontSize float64, fontPath string) (int, int, error) {
	c := freetype.NewContext()

	c.SetDPI(72)
	c.SetHinting(font.HintingFull)
	// 设置文字颜色、字体、字大小
	c.SetSrc(image.NewUniform(color.RGBA{R: 51, G: 51, B: 51, A: 254}))
	c.SetFontSize(fontSize)
	fontFam, err := getFontFamilyFont(fontPath)
	if err != nil {
		return 0, 0, err
	}
	c.SetFont(fontFam)

	pt := freetype.Pt(0, 0)

	newPt, err := c.DrawString(content, pt)
	if err != nil {
		return 0, 0, err
	}
	const shift, mask = 6, 1<<6 - 1
	x := int32(newPt.X)
	if x >= 0 {
		return int(int32(x >> shift)), int(int32(x & mask)), nil
	}
	x = -x
	if x >= 0 {
		return -int(int32(x >> shift)), int(int32(x & mask)), nil
	}
	return 0, 0, nil
}

// WriteOnImageFont 根据指定字体写入图片
func WriteOnImageFont(target *image.NRGBA, content string, fontSize float64, x int, y int, color color.RGBA, fontPath string) error {
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetClip(target.Bounds())
	c.SetDst(target)
	c.SetHinting(font.HintingFull)
	// 设置文字颜色、字体、字大小
	c.SetSrc(image.NewUniform(color))
	c.SetFontSize(fontSize)
	fontFam, err := getFontFamilyFont(fontPath)
	if err != nil {
		return err
	}
	c.SetFont(fontFam)
	pt := freetype.Pt(x, y)
	_, err = c.DrawString(content, pt)
	if err != nil {
		return err
	}
	return nil
}

// GetWriteLinesWidthFont 获取行数
func GetWriteLinesWidthFont(content string, fontSize float64, width int, fontPath string) ([]string, error) {
	lines := make([]string, 0)
	contents := strings.Split(content, "")
	lastStart := 0
	for i := range contents {
		word := strings.Join(contents[lastStart:i+1], "")
		_, ra, err := GetStringSize(
			word,
			fontPath,
			fontSize,
			72,
		)
		if err != nil {
			return nil, err
		}
		raStr := ra.String()
		raStrs := strings.Split(raStr, ":")
		raX, err := strconv.ParseInt(raStrs[0], 10, 64)
		if err != nil {
			return nil, err
		}
		posX := int(raX)
		if posX > width {
			lines = append(lines, word)
			lastStart = i + 1
		}
		if i == len(contents)-1 && lastStart != i+1 {
			lines = append(lines, word)
		}
	}
	return lines, nil
}

// WriteOnImageLinesFontWithGap 写入多行文字到图片
func WriteOnImageLinesFontWithGap(target *image.NRGBA, lines []string, fontSize float64, x int, y int, color color.RGBA, fontPath string, gap float64) error {
	c := freetype.NewContext()

	c.SetDPI(72)
	c.SetClip(target.Bounds())
	c.SetDst(target)
	c.SetHinting(font.HintingFull)

	// 设置文字颜色、字体、字大小
	c.SetSrc(image.NewUniform(color))
	c.SetFontSize(fontSize)
	fontFam, err := getFontFamilyFont(fontPath)
	if err != nil {
		return err
	}
	c.SetFont(fontFam)

	for i, line := range lines {
		pt := freetype.Pt(x, y+i*(int(fontSize+gap)))
		_, err = c.DrawString(line, pt)
		if err != nil {
			return err
		}
	}
	return nil
}
