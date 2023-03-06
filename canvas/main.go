package main

import (
	"image/color"
	"log"

	"github.com/fzipp/canvas"
)

func main() {
	err := canvas.ListenAndServe(":8080", run,
		canvas.Size(100, 80),
		canvas.Title("Example 1: Drawing"),
	)
	if err != nil {
		log.Fatal(err)
	}
}

func run(ctx *canvas.Context) {
	ctx.SetFillStyle(color.RGBA{R: 200, A: 255})
	ctx.FillRect(10, 10, 50, 50)
	// ...
	ctx.Flush()
}
