package demo

import (
	"html/template"

	"github.com/dmundt/stitch/css"
)

func init() {
	css.MustRegister(css.NewStaticProvider("minstyle", template.HTML(`<link rel="stylesheet" href="https://unpkg.com/minstyle.io@latest/dist/css/minstyle.io.min.css">`)))
	css.MustRegister(css.NewStaticProvider("milligram", template.HTML(`<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/milligram@1.4.1/dist/milligram.min.css">`)))
}
