package main

import (
	"github.com/Walchand-Linux-Users-Group/wargames/api/helpers"
)

func main() {

	helpers.InitEnv()

	InitDatabase()

	InitAPI()
}
