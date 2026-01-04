package main

func main() {
	cfg.init()
	db := NewDB()
	launchServer(db)
}

