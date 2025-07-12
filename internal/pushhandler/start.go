package pushhandler

func Start() {
	go func() {
		handle()
	}()
}
