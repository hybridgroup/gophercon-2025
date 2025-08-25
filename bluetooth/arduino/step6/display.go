package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/ssd1306"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyterm"
)

var (
	font = &tinyfont.TomThumb
)

func initTerminal() {
	display := initDisplay()

	time.Sleep(time.Millisecond * 100)

	terminal = tinyterm.NewTerminal(display)
	terminal.Configure(&tinyterm.Config{
		Font:              font,
		FontHeight:        6,
		FontOffset:        4,
		UseSoftwareScroll: true,
	})
}

func initDisplay() tinyterm.Displayer {
	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: machine.TWI_FREQ_400KHZ,
	})

	display := ssd1306.NewI2C(machine.I2C0)
	display.Configure(ssd1306.Config{
		Address: ssd1306.Address_128_32,
		Width:   128,
		Height:  32,
	})

	display.ClearDisplay()

	return display
}
