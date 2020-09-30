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
