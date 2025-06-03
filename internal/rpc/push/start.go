package push

import (
	push "github.com/roc/roc-im-server/internal/kitex_gen/push/pushservice"
	"log"
)

func main() {
	svr := push.NewServer(new(PushServiceImpl))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
