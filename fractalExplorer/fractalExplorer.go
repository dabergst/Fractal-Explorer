// Daniel Bergström
// dabergst@kth.se

package main

/**
 * TODO: 
 * Implement scaleable box cursor:
 *	gdk_bitmap_create_from_data
 *	gdk_cursor_new_from_pixmap 
 * Choice of colorscheme
 * User defined colorscheme
 * Save file dialog 
 * Reset confirmation
 * Expand window instead of reallocate space.
 * Real anti-alisaing via multisampling (also downsampling)
 * Add statusbar
 */

import (
	"fmt"
	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
	"log"
	"os"
	"runtime"
	"image"
	"image/color"
	"image/png"
	"saph/graphic"
	"saph/fractal"
	"saph/graphic/palette"
	"strconv"
	"time"
	"unsafe"
)

func init() {
	log.SetFlags(0) // no extra info in log messages
	//log.SetOutput(ioutil.Discard) // turns off logging

	// Use all CPU's.
	numcpu := runtime.NumCPU()
	log.Println("CPU count:", numcpu)
	runtime.GOMAXPROCS(numcpu)
}

var window *gtk.Window
var drawingarea *gtk.DrawingArea
var pixmap *gdk.Pixmap
var gc *gdk.GC
var pixChan chan graphic.Pixel

var frac *fractal.Fractal
var imageSize graphic.Box

var maxIterationsEntry *gtk.Entry
var bailoutRadiusEntry *gtk.Entry
var normalizeCheckbutton *gtk.CheckButton
var multisampleComboBoxText *gtk.ComboBoxText
var colorFrequencyEntry *gtk.Entry
var progressBar *gtk.ProgressBar

var before time.Time

var cursor int

func main() {


	cursor = 0

	gtk.Init(nil)
	window = gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetPosition(gtk.WIN_POS_CENTER)
	window.SetTitle("Fractal Explorer")
	window.SetIconName("gtk-zoom-fit")
	window.Connect("destroy", gtk.MainQuit)

	//frac = fractal.NewJulia(complex(-0.7, 0.3))
	frac = fractal.NewMandelbrot()

	 // Layout
	vbox1 := gtk.NewVBox(false, 5)
	vbox1.SetBorderWidth(5)
		
		frame11 := gtk.NewFrame("Left click - zoom in, right click - zoom out")
		vbox1.PackStart(frame11, true, true, 0)
		
		hbox12f := gtk.NewHBox(false, 0)
		vbox1.PackStart(hbox12f, false, false, 0)

		hbox12 := gtk.NewHBox(false, 15)
		hbox12f.PackStart(hbox12, true, false, 0)
	
			frame121f := gtk.NewFrame("Quality settings")
			hbox12.PackStart(frame121f, true, true, 0)
				
			hbox121 := gtk.NewHBox(false, 5)
			frame121f.Add(hbox121)
			
				vbox1211 := gtk.NewVBox(true, 0)
				hbox121.PackStart(vbox1211, true, true, 0)
				
				vbox1212 := gtk.NewVBox(true, 0)
				hbox121.PackStart(vbox1212, true, true, 0)
				
				vsep := gtk.NewVSeparator()
				hbox121.PackStart(vsep, false, false, 5)
				
				vbox1213 := gtk.NewVBox(true, 0)
				hbox121.PackStart(vbox1213, true, true, 0)
				
				vbox1214 := gtk.NewVBox(true, 0)
				hbox121.PackStart(vbox1214, true, true, 0)
					
			frame122f := gtk.NewFrame("Color settings")
			hbox12.PackStart(frame122f, true, true, 0)
				
			hbox122 := gtk.NewHBox(false, 5)
			frame122f.Add(hbox122)
			
				vbox1221 := gtk.NewVBox(true, 0)
				hbox122.PackStart(vbox1221, true, true, 0)
				
				vbox1222 := gtk.NewVBox(true, 0)
				hbox122.PackStart(vbox1222, true, true, 0)
				
				vsep = gtk.NewVSeparator()
				hbox122.PackStart(vsep, false, false, 5)
				
				vbox1223 := gtk.NewVBox(true, 0)
				hbox122.PackStart(vbox1223, true, true, 0)
				
				vbox1224 := gtk.NewVBox(true, 0)
				hbox122.PackStart(vbox1224, true, true, 0)
				
		hbox13 := gtk.NewHBox(true, 0)
		vbox1.PackStart(hbox13, false, false, 0)


	menubar := gtk.NewMenuBar()
	cascademenu := gtk.NewMenuItemWithMnemonic("_File")
	menubar.Append(cascademenu)
	submenu := gtk.NewMenu()
	cascademenu.SetSubmenu(submenu)
	var menuitem *gtk.MenuItem
	vbox1.PackStart(menubar, false, false, 0)

	/*menuitem = gtk.NewMenuItemWithMnemonic("_Reset")
	menuitem.Connect("activate", func() {
		if isRendering {
			return
		}
		renderLock()
		bounds := frac.Bounds()
		frac.Reset()
		frac.SetBounds(bounds)
		updateEntries()
		ch = frac.Render()
		glib.IdleAdd(printFractal)
	})
	submenu.Append(menuitem)
*/
/*	menuitem = gtk.NewMenuItemWithMnemonic("_Save image")
	menuitem.Connect("activate", func() {
		if isRendering {
			return
		}
		renderLock()
	
		temp := imageSize
		imageSize = graphic.Box{1920, 1200}
		
		go func() {
			img := image.NewRGBA(image.Rect(0, 0, 1920, 1200))
			render()

			for c := range pixChan {
				color := color.RGBA{c.R, c.G, c.B, c.A}
				img.Set(c.X, c.Y, color)
			}

			err := CreatePng("mandel111", img)
			if err != nil {
				log.Println(err)
			}

			imageSize = temp
			renderUnlock()
		}()
	})
	submenu.Append(menuitem)
*/
	menuitem = gtk.NewMenuItemWithMnemonic("E_xit")
	menuitem.Connect("activate", func() {
		gtk.MainQuit()
	})
	submenu.Append(menuitem)


	//~~~~~~~~~~~~ DrawingArea - Fractal ~~~~~~~~~~~~
	drawingarea = gtk.NewDrawingArea()
	drawingarea.Connect("configure-event", func() {
		if !frac.IsFinished() { return }
		if pixmap != nil {
			pixmap.Unref()
		}
		var allocation gtk.Allocation
		drawingarea.GetAllocation(&allocation)
		pixmap = gdk.NewPixmap(drawingarea.GetWindow().GetDrawable(), allocation.Width, allocation.Height, 24)
		gc = gdk.NewGC(pixmap.GetDrawable())
		gc.SetRgbFgColor(gdk.NewColor("white"))
		pixmap.GetDrawable().DrawRectangle(gc, true, 0, 0, -1, -1)
		imageSize = graphic.Box{allocation.Width, allocation.Height}
		render()
	})

	drawingarea.Connect("expose-event", func() {
		if pixmap != nil {
			drawingarea.GetWindow().GetDrawable().DrawDrawable(gc, pixmap.GetDrawable(), 0, 0, 0, 0, -1, -1)
		}
	})

	drawingarea.Connect("button-press-event", func(ctx *glib.CallbackContext) {
		var x, y int
		var mt gdk.ModifierType
		drawingarea.GetWindow().GetPointer(&x, &y, &mt)
		magnifySize := graphic.Box{imageSize.Width/10, imageSize.Height/10}
		frac.Magnify(imageSize, magnifySize, image.Point{x,y})
		render()		
	})

	drawingarea.SetEvents(int(gdk.BUTTON_PRESS_MASK))
	frame11.Add(drawingarea)

	//~~~~~~~~~~~~ Entry - Max iterations ~~~~~~~~~~~~
	maxIterationsEntry = gtk.NewEntry()
	maxIterationsEntry.SetEditable(true)
	maxIterationsEntry.SetMaxLength(6)
	maxIterationsEntry.SetWidthChars(10)
	maxIterationsEntry.SetText("300")
	maxIterationsEntry.Connect("insert-text", func(ctx *glib.CallbackContext) {
		if !entryInsertIsInt(ctx) {
			maxIterationsEntry.StopEmission("insert-text")
		}
	})
	vbox1211.PackStart(NewLeftAlignedLabel("Max iterations:"), true, true, 0)
	vbox1212.PackStart(maxIterationsEntry, true, true, 0)

	//~~~~~~~~~~~~ Entry - Bailout radius ~~~~~~~~~~~~
	bailoutRadiusEntry = gtk.NewEntry()
	bailoutRadiusEntry.SetEditable(true)
	bailoutRadiusEntry.SetMaxLength(10)
	bailoutRadiusEntry.SetWidthChars(10)
	bailoutRadiusEntry.SetText("20")
	bailoutRadiusEntry.Connect("insert-text", func(ctx *glib.CallbackContext) {
		if !entryInsertIsInt(ctx) {
			bailoutRadiusEntry.StopEmission("insert-text")
		}
	})	
	vbox1211.PackStart(NewLeftAlignedLabel("Bailout radius:"), true, true, 0)
	vbox1212.PackStart(bailoutRadiusEntry, true, true, 0)
		
	//~~~~~~~~~~~~ ComboBoxText - Multisample ~~~~~~~~~~~~
	multisampleComboBoxText = gtk.NewComboBoxText()
	multisampleComboBoxText.AppendText("x4")
	multisampleComboBoxText.AppendText("x3")
	multisampleComboBoxText.AppendText("x2")
	multisampleComboBoxText.AppendText("x1")
	multisampleComboBoxText.AppendText("/1")
	multisampleComboBoxText.AppendText("/2")
	multisampleComboBoxText.AppendText("/3")
	multisampleComboBoxText.AppendText("/4")
	multisampleComboBoxText.SetActive(3)
	vbox1213.PackStart(NewLeftAlignedLabel("Multisample:"), true, true, 0)
	vbox1214.PackStart(multisampleComboBoxText, true, true, 0)
		
	//~~~~~~~~~~~~ CheckButton - Normalize ~~~~~~~~~~~~
	normalizeCheckbutton = gtk.NewCheckButton()
	normalizeCheckbutton.SetActive(true)
	vbox1213.PackStart(NewLeftAlignedLabel("Normalize:"), true, true, 0)
	vbox1214.PackStart(normalizeCheckbutton, true, true, 0)
	
	paletteCombobox := gtk.NewComboBoxText()
	paletteCombobox.AppendText("Peach")
	paletteCombobox.AppendText("Banana")
	paletteCombobox.AppendText("Apple")
	vbox1221.PackStart(NewLeftAlignedLabel("Palette:"), true, true, 0)
	vbox1222.PackStart(paletteCombobox, true, true, 0)

	//~~~~~~~~~~~~ Entry - Color frequency ~~~~~~~~~~~~
	colorFrequencyEntry = gtk.NewEntry()
	colorFrequencyEntry.SetEditable(true)
	colorFrequencyEntry.SetMaxLength(10)
	colorFrequencyEntry.SetWidthChars(10)
	colorFrequencyEntry.SetText("20")
	colorFrequencyEntry.Connect("insert-text", func(ctx *glib.CallbackContext) {
		if !entryInsertIsInt(ctx) {
			colorFrequencyEntry.StopEmission("insert-text")
		}
	})
	vbox1221.PackStart(NewLeftAlignedLabel("Frequency:"), true, true, 0)
	vbox1222.PackStart(colorFrequencyEntry, true, true, 0)
	
	
	setColorCombobox := gtk.NewComboBoxText()
	setColorCombobox.AppendText("Peach")
	setColorCombobox.AppendText("Banana")
	setColorCombobox.AppendText("Apple")
	vbox1223.PackStart(NewLeftAlignedLabel("Set color:"), true, true, 0)
	vbox1224.PackStart(setColorCombobox, true, true, 0)
	
	dummyL := gtk.NewLabel("")
	dummyR := gtk.NewLabel("")
	vbox1223.PackStart(dummyL, true, true, 0)
	vbox1224.PackStart(dummyR, true, true, 0)

	//~~~~~~~~~~~~ Button - Render ~~~~~~~~~~~~
	button := gtk.NewButtonWithLabel("               Render               ")
	button.Clicked(func() {
		render()
	})
	hbox12.PackStart(button, false, false, 0)
	
	progressBar = gtk.NewProgressBar()
	progressBar.SetFraction(0.0)
	progressBar.SetText("0%")
	hbox13.PackStart(progressBar, true, true, 0)

	window.Add(vbox1)
	defaultSizeRequest()
	window.ShowAll()
	gtk.Main()
}

func defaultSizeRequest() {
	window.SetSizeRequest(1000, 700)
}

func entryInsertIsInt(ctx *glib.CallbackContext) bool {
	a := (*[2000]uint8)(unsafe.Pointer(ctx.Args(0)))
	i := 0
	for a[i] != 0 {
		i++
	}
	s := string(a[0:i])
	_, err := strconv.Atoi(s)
	if err != nil {
		return false
	}
	return true
}

func NewLeftAlignedLabel(labelText string) gtk.IWidget {
	hbox := gtk.NewHBox(false, 0)
	LAbox := gtk.NewHBox(false, 0)
	hbox.PackStart(LAbox, true, true, 0)
	label := gtk.NewLabel(labelText)
	LAbox.PackStart(label, false, false, 0)
	return hbox
}


func render() {
	renderLock()
	maxIterations, _ := strconv.Atoi(maxIterationsEntry.GetText())
	bailoutRadius, _ := strconv.ParseFloat(bailoutRadiusEntry.GetText(), 64)
	normalize := normalizeCheckbutton.GetActive()
	sampleRatio := 4 - multisampleComboBoxText.GetActive() // OBS 0
	colorScheme := palette.CyclicPalette([]color.RGBA{palette.Orange, palette.White, palette.OrangeRed, palette.Red})
	colorFrequency, _ := strconv.ParseFloat(colorFrequencyEntry.GetText(), 64)
	setColor := palette.Black
	pixChan = frac.Render(imageSize, maxIterations, bailoutRadius, normalize,	sampleRatio, setColor, colorScheme, colorFrequency)
	before = time.Now()
	glib.IdleAdd(printPixChan)
}

func renderLock() {
	window.SetSensitive(false)
	window.SetResizable(false)
	var x, y int
	window.GetSize(&x, &y)
	window.SetSizeRequest(x, y)
	drawingarea.GetWindow().SetCursor(gdk.NewCursor(gdk.WATCH))
}

func renderUnlock() {
	defaultSizeRequest()
	window.SetSensitive(true)
	window.SetResizable(true)
	drawingarea.GetWindow().SetCursor(gdk.NewCursor(gdk.TCROSS))
}

func printPixChan() bool {
	defer drawingarea.GetWindow().Invalidate(nil, false)
	defer func(){
		progress := frac.GetProgress()
		progressBar.SetFraction(progress)
		progressBar.SetText(fmt.Sprintf("%d%%", int(progress*100)))
	}()
	timeout := time.After(time.Millisecond*100)
	for {
		select {
		case <-timeout:
			return true
		case c, ok := <-pixChan:
			if !ok {
				renderUnlock()
				log.Println("Time: ", time.Now().Sub(before))
				return false
			}
			// EXPENSIVE NON ALLOCATED!!!!!!!
			gdkColor := gdk.NewColor(fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)) 
			gc.SetRgbFgColor(gdkColor)
			pixmap.GetDrawable().DrawPoint(gc, c.X, c.Y)
		}
	}
	return false
}

func CreatePng(filename string, img image.Image) (err error) {
	file, err := os.Create(filename + ".png")
	if err != nil {
		log.Fatalf("CreatePng: %s", err)
	}
	defer file.Close()
	err = png.Encode(file, img)
	return
}
