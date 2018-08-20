package main

func main() {
	runner := NewRunner(NewApp)
	runner.Start()
}
