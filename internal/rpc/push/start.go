package push

import (
	push "github.com/rhp-QE/roc-im-server/internal/kitex_gen/push/pushservice"
	"log"
)

func main() {
	svr := push.NewServer(new(PushServiceImpl))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
