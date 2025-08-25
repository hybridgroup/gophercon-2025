package main

import (
	"image/color"
	"machine"
	"strconv"
	"time"

	"tinygo.org/x/drivers/buzzer"
	"tinygo.org/x/drivers/ssd1306"
	"tinygo.org/x/tinydraw"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
)

var (
	green  = machine.D12
	red    = machine.D10
	button = machine.D11
	touch  = machine.D9
	bzrPin = machine.D8

	bzr    buzzer.Device
	dial   = machine.ADC{machine.ADC0}
	pwm    = machine.PWM2 // PWM2 corresponds to Pin D10.
	redPwm uint8

	dialValue  uint16
	buttonPush bool
	touchPush  bool

	systemActive   bool
	alarmTriggered bool
	alarmLevel     uint16 = 32000

	systemTest bool
)

func main() {
	initDevices()
	connectWifi()

	go handleDisplay()
	go startWebServer()

	for {
		systemActivationStatusButton()
		systemActivationStatusLED()

		handleSensorReading()
		handleSystemTest()
		handleAlarm()

		time.Sleep(time.Millisecond * 100)
	}
}

func initDevices() {
	green.Configure(machine.PinConfig{Mode: machine.PinOutput})
	button.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	touch.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	bzrPin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	err := pwm.Configure(machine.PWMConfig{
		Period: 16384e3, // 16.384ms
	})
	if err != nil {
		println("failed to configure PWM")
		return
	}
	redPwm, err = pwm.Channel(red)
	if err != nil {
		println("failed to configure PWM channel")
		return
	}

	machine.InitADC()
	dial.Configure(machine.ADCConfig{})

	bzr = buzzer.New(bzrPin)
}

func handleDisplay() {
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

	black := color.RGBA{1, 1, 1, 255}

	for {
		display.ClearBuffer()

		msg := "off"
		if systemActive {
			val := strconv.Itoa(int(dialValue))
			msg = "pwr: " + val
		}

		tinyfont.WriteLine(display, &freemono.Bold9pt7b, 10, 20, msg, black)

		var radius int16 = 4
		if systemActive {
			tinydraw.FilledCircle(display, 16+32*0, 32-radius-1, radius, black)
		} else {
			tinydraw.Circle(display, 16+32*0, 32-radius-1, radius, black)
		}
		if alarmTriggered {
			tinydraw.FilledCircle(display, 16+32*1, 32-radius-1, radius, black)
		} else {
			tinydraw.Circle(display, 16+32*1, 32-radius-1, radius, black)
		}

		display.Display()

		time.Sleep(100 * time.Millisecond)
	}
}

func systemActivationStatusButton() {
	pushed := button.Get()
	switch {
	case pushed && buttonPush:
		// already pushed, do nothing
		return
	case !pushed && buttonPush:
		// we released the button
		buttonPush = false
		return
	case pushed && !buttonPush:
		// we pushed the button
		systemActive = !systemActive
		buttonPush = false
	default:
		// do nothing
	}
}

func systemActivationStatusLED() {
	if systemActive {
		green.High()
	} else {
		green.Low()
	}
}

func handleSensorReading() {
	if !systemActive {
		return
	}

	alarmTriggered = false
	dialValue = dial.Get()
	pwm.Set(redPwm, uint32(dialValue))

	if dialValue > alarmLevel {
		alarmTriggered = true
	}
}

func handleAlarm() {
	if alarmTriggered {
		bzr.On()
	} else {
		bzr.Off()
	}
}

func handleSystemTest() {
	// ignore if the system is already active
	if systemActive {
		return
	}

	pushed := touch.Get()
	switch {
	case pushed && touchPush:
		// already pushed, do nothing
		return
	case !pushed && touchPush:
		// we released the button
		touchPush = false
		return
	case pushed && !touchPush:
		// we pushed the button
		alarmTriggered = !alarmTriggered
		touchPush = false
	default:
		// do nothing
	}
}
