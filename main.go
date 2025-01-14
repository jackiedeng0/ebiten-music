package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type MusicBarsSettings struct {
	NumBars   uint
	MaxWidth  float64
	MaxHeight float64
	StartFreq float64
	EndFreq   float64
}

type Bar struct {
	Freq      float64
	HeightPct float64
}

func (settings *MusicBarsSettings) initBars() []Bar {
	bars := make([]Bar, settings.NumBars)
	for i := range settings.NumBars {
		freq := (settings.StartFreq * math.Pow(
			settings.EndFreq/settings.StartFreq, float64(i/settings.NumBars)))
		// TODO: set initial HeightPct to 0
		bars[i] = Bar{Freq: freq, HeightPct: (float64(i) * 10) / 100}
	}
	return bars
}

var settings MusicBarsSettings
var bars []Bar
var barWidth float64

type Game struct{}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	rectColor := color.RGBA{255, 255, 255, 255}
	for i, bar := range bars {
		ebitenutil.DrawRect(screen,
			float64(i)*barWidth, 240,
			barWidth, -(bar.HeightPct * settings.MaxHeight),
			rectColor)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 320, 240
}

func main() {
	settings = MusicBarsSettings{
		NumBars:   10,
		MaxWidth:  320,
		MaxHeight: 240,
		StartFreq: 20,
		EndFreq:   20,
	}
	bars = settings.initBars()
	barWidth = settings.MaxWidth / float64(settings.NumBars)
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Music Visualizer")
	if err := ebiten.RunGame(&Game{}); err != nil {
		panic(err)
	}
}
