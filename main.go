package main

import (
        "image"
        "image/draw"
        "log"
        "time"

        "golang.org/x/image/font"
        "golang.org/x/image/font/basicfont"
        "golang.org/x/image/math/fixed"

        "periph.io/x/conn/v3/spi/spireg"
        "periph.io/x/devices/v3/ssd1306/image1bit"
        "periph.io/x/devices/v3/waveshare2in13v2"
        "periph.io/x/host/v3"
)

func main() {
        // Make sure periph is initialized.
        if _, err := host.Init(); err != nil {
                log.Fatal(err)
        }

        // Use spireg SPI bus registry to find the first available SPI bus.
        b, err := spireg.Open("")
        if err != nil {
                log.Fatal(err)
        }
        defer b.Close()

        // Use EPD2in13v2 configuration
        dev, err := waveshare2in13v2.NewHat(b, &waveshare2in13v2.EPD2in13v2)
        dev.SetUpdateMode(waveshare2in13v2.Partial)
        if err != nil {
                log.Fatalf("Failed to initialize driver: %v", err)
        }

        err = dev.Init()
        if err != nil {
                log.Fatalf("Failed to initialize display: %v", err)
        }

        // Function to create the display image
        createDisplayImage := func() (image.Rectangle, *image1bit.VerticalLSB) {
                bounds := dev.Bounds()
                tempImg := image1bit.NewVerticalLSB(image.Rect(0, 0, bounds.Dy(), bounds.Dx()))

                draw.Draw(tempImg, tempImg.Bounds(), &image.Uniform{image1bit.On}, image.Point{}, draw.Src)

                f := basicfont.Face7x13
                lineHeight := f.Metrics().Height.Floor()

                timeStr := time.Now().UTC().Format("15:04:05")

                drawer := font.Drawer{
                        Dst:  tempImg,
                        Src:  &image.Uniform{image1bit.Off},
                        Face: f,
                }

                startX := 10
                startY := 30

                drawer.Dot = fixed.P(startX, startY)
                drawer.DrawString(timeStr)

                drawer.Dot = fixed.P(startX, startY+lineHeight)
                drawer.DrawString("haxrob")

                displayImg := image1bit.NewVerticalLSB(bounds)
                draw.Draw(displayImg, displayImg.Bounds(), &image.Uniform{image1bit.On}, image.Point{}, draw.Src)

                for y := 0; y < tempImg.Bounds().Dy(); y++ {
                        for x := 0; x < tempImg.Bounds().Dx(); x++ {
                                if tempImg.BitAt(x, y) == image1bit.Off {
                                        newX := tempImg.Bounds().Dy() - y - 1
                                        newY := x
                                        displayImg.SetBit(newX, newY, image1bit.Off)
                                }
                        }
                }

                return bounds, displayImg
        }

        // Initial display with full Draw
        bounds, displayImg := createDisplayImage()
        if err := dev.DrawPartial(bounds, displayImg, image.Point{}); err != nil {
                log.Printf("Error in initial display update: %v", err)
        }

        // Update function using DrawPartial
        updateDisplay := func() {
                bounds, displayImg := createDisplayImage()
                if err := dev.DrawPartial(bounds, displayImg, image.Point{}); err != nil {
                        log.Printf("Error updating display: %v", err)
                }
        }

        // Update every second
        ticker := time.NewTicker(1 * time.Second)
        defer ticker.Stop()

        log.Println("Updating display every second. Press Ctrl+C to stop.")

        // Main loop
        for range ticker.C {
                updateDisplay()
        }
}
