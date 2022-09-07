package main

import (
	"github.com/Walchand-Linux-Users-Group/wargames/api/helpers/env"
)

func main() {

	env.InitEnv()

	InitDatabase()

	InitAPI()
}
