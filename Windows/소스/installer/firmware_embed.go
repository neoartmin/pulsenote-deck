package main

import _ "embed"

// The installer installs the reference sketches but does not flash hardware.
//
//go:embed firmware_pn12.ino
var firmwarePN12 []byte

//go:embed firmware_wiring.txt
var firmwareWiring []byte

//go:embed firmware_pn15.ino
var firmwarePN15 []byte

//go:embed firmware_pn15_left.ino
var firmwarePN15Left []byte
