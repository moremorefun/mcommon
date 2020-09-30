package mcommon

import (
	"image"
	"image/color"
	"io/ioutil"

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
