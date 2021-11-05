package webui

import (
	"log"
	"os"
)

var logger = log.New(os.Stdout, "realtime: ", log.LstdFlags)
