package main

import (
	"encoding/binary"
	"image/color"
	"io"
	"math"

	"math/cmplx"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"gonum.org/v1/gonum/dsp/fourier"
)

type MusicBarsSettings struct {
	NumBars   uint
	MaxWidth  float64
	MaxHeight float64
}

func (settings *MusicBarsSettings) initBars() []float64 {
	bars := make([]float64, settings.NumBars)
	for i := range settings.NumBars {
		bars[i] = 0
	}
	return bars
}

var settings MusicBarsSettings
var bars []float64
var barWidth float64
var visualizerStream *mp3.Stream

var startTime time.Time

type Game struct{}

func invariant(err error) {
	if err != nil {
		panic(err)
	}
}

type Stereo struct {
	Left  float32
	Right float32
}

var stereo []Stereo
var fft *fourier.FFT
var leftF64 []float64

const (
	BytesPerSample     = 8
	TicksPerSecond     = 60
	Bars               = 20
	MaxLogAmplitude    = 4.0
	MinLogAmplitude    = -2.0
	MovingAverageRatio = 0.3
)

var frequenciesPerBar int

func VisualizerSetup() {
	var samplesPerTick int = visualizerStream.SampleRate() / TicksPerSecond
	stereo = make([]Stereo, samplesPerTick)
	fft = fourier.NewFFT(samplesPerTick)
	leftF64 = make([]float64, samplesPerTick)
	fftFrequencies := fft.Len()/2 + 1
	frequenciesPerBar = fftFrequencies / Bars
}

func (g *Game) Update() error {
	visualizerStream.Seek(int64(float64(time.Since(startTime).Milliseconds())*
		float64(visualizerStream.SampleRate())/1000.0)*BytesPerSample,
		io.SeekStart)

	err := binary.Read(visualizerStream, binary.LittleEndian, &stereo)
	for i, v := range stereo {
		leftF64[i] = float64(v.Left)
	}
	coeff := fft.Coefficients(nil, leftF64)
	runningSum := 0.0
	for i, c := range coeff[1:] {
		runningSum += cmplx.Abs(c)
		if i%frequenciesPerBar == 0 {
			if i/frequenciesPerBar == Bars {
				break
			}
			logAverageFrequency := math.Log(runningSum / float64(frequenciesPerBar))
			currentCappedFrequency := 0.0
			if logAverageFrequency > MaxLogAmplitude {
				currentCappedFrequency = 1.0
				bars[i/frequenciesPerBar] = 1.0
			} else if logAverageFrequency < MinLogAmplitude {
				currentCappedFrequency = 0.0
				bars[i/frequenciesPerBar] = 0.0
			} else {
				currentCappedFrequency = ((logAverageFrequency - MinLogAmplitude) / (MaxLogAmplitude - MinLogAmplitude))
			}
			bars[i/frequenciesPerBar] = (bars[i/frequenciesPerBar]*MovingAverageRatio + currentCappedFrequency*(1-MovingAverageRatio))
			runningSum = 0.0
		}
	}
	return err
}

func (g *Game) Draw(screen *ebiten.Image) {
	rectColor := color.RGBA{255, 255, 255, 255}
	for i, bar := range bars {
		ebitenutil.DrawRect(screen,
			float64(i)*barWidth, 240,
			barWidth, -(bar * settings.MaxHeight),
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
	defer file.Close()
	invariant(err)
	audioStream, err := mp3.DecodeF32(file)
	invariant(err)

	file2, err := os.Open(os.Args[1])
	defer file2.Close()
	invariant(err)
	visualizerStream, err = mp3.DecodeF32(file2)
	invariant(err)
	VisualizerSetup()

	context := audio.NewContext(audioStream.SampleRate())
	musicPlayer, err := context.NewPlayerF32(audioStream)
	invariant(err)
	musicPlayer.Play()
	startTime = time.Now()

	settings = MusicBarsSettings{
		NumBars:   20,
		MaxWidth:  320,
		MaxHeight: 240,
	}
	bars = settings.initBars()
	barWidth = settings.MaxWidth / float64(settings.NumBars)
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Music Visualizer")
	if err := ebiten.RunGame(&Game{}); err != nil {
		panic(err)
	}
}
