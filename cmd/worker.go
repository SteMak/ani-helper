package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/SteMak/ani-helper/workerTools/config"
	"github.com/SteMak/ani-helper/workerTools/database"
	"github.com/SteMak/ani-helper/workerTools/logic"
)

func main() {
	config.Init()

	database.Init()
	defer database.Close()

	rand.Seed(time.Now().UnixNano())

	logic.Init()

	for i := 1; i > 0; i++ {
		time.Sleep(25 * time.Minute)
		fmt.Println("WORKS for", 25*i, "minutes")
	}
}
