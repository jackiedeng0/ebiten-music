package main

import (
	"encoding/binary"
	"fmt"
	"image/color"
	"io"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
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
	Freq       float64
	Proportion float64
}

func (settings *MusicBarsSettings) initBars() []Bar {
	bars := make([]Bar, settings.NumBars)
	for i := range settings.NumBars {
		freq := (settings.StartFreq * math.Pow(
			settings.EndFreq/settings.StartFreq, float64(i/settings.NumBars)))
		bars[i] = Bar{Freq: freq, Proportion: 0}
	}
	return bars
}

var settings MusicBarsSettings
var bars []Bar
var barWidth float64
var stream2 *mp3.Stream

type Game struct{}

func invariant(err error) {
	if err != nil {
		panic(err)
	}
}

func (g *Game) Update() error {
	bytes := make([]byte, 1)
	_, err := stream2.Read(bytes)
	invariant(err)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	rectColor := color.RGBA{255, 255, 255, 255}
	for i, bar := range bars {
		ebitenutil.DrawRect(screen,
			float64(i)*barWidth, 240,
			barWidth, -(bar.Proportion * settings.MaxHeight),
			rectColor)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 320, 240
}

func main() {
	if len(os.Args) < 2 {
		panic("Takes an mp3 file as argument.")
	}
	file, err := os.Open(os.Args[1])
	invariant(err)
	stream, err := mp3.DecodeF32(file)
	invariant(err)
	stream2, err = mp3.DecodeF32(file)
	invariant(err)

	var L float32 = 0
	var R float32 = 0
	for err == nil {
		err = binary.Read(stream2, binary.LittleEndian, &L)
		err = binary.Read(stream2, binary.LittleEndian, &R)
		fmt.Printf("L: %f, R: %f\n", L, R)
	}
	stream2.Seek(0, io.SeekStart)

	context := audio.NewContext(stream.SampleRate())
	player, err := context.NewPlayerF32(stream)
	invariant(err)
	player.Play()

	settings = MusicBarsSettings{
		NumBars:   20,
		MaxWidth:  320,
		MaxHeight: 240,
		StartFreq: 20,
		EndFreq:   20000,
	}
	bars = settings.initBars()
	barWidth = settings.MaxWidth / float64(settings.NumBars)
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Music Visualizer")
	if err := ebiten.RunGame(&Game{}); err != nil {
		panic(err)
	}
}
