package main

import (
	"backD/internal/device"
	"fmt"
)

func main() {

	extDevice, err := device.Detect()
	if err != nil {
		fmt.Printf("Error : %v", err)
	}

	for _, o := range extDevice {
		fmt.Print(o)
	}
}
