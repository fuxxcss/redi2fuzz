package analyze

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fuxxcss/redi2fuzz/pkg/db"
	"github.com/fuxxcss/redi2fuzz/pkg/utils"
	"github.com/fuxxcss/redi2fuzz/pkg/model"

)

func Analyze(target utils.TargetType, path string) {

	// Analyze Target (redis, keydb, redis-stack)
	feature := utils.Targets[target]
	context, err := os.ReadFile(path)

	if err != nil {
		log.Fatalln("err: bug file failed.")
	}

	// interface
	var DBtarget db.DB

	switch target {
	// Redi
	case utils.REDI_REDIS, utils.REDI_KEYDB, utils.REDI_STACK:
		DBtarget = db.NewRedi(feature)
	}
	
	// StartUp target first
	err = DBtarget.StartUp()
	defer DBtarget.ShutDown()
	
	if err != nil {
		log.Println("err: db startup failed.")
		return
	}

	// test bug
	lines := strings.Split(string(context), model.LineSep)
	index := -1
	var bug string

	for i, line := range lines {

		// execute each line
		tokens := strings.Split(line, model.TokenSep)
		DBtarget.Execute(tokens)

		alive := DBtarget.CheckAlive()

		if !alive {
			index = i
			bug = line
			break
		}
	}

	// trigger bug
	if index >= 0 {

		fmt.Printf("line %d trigger bug\n", index+1)
		fmt.Println(bug)
		fmt.Println(DBtarget.Stderr())

	} else {
		fmt.Println("not a bug")
	}

}
