package main

func main() {
	m := mic{&defaultAudioInput{}, &defaultSocketNet{}, sigTermChannel()}
	m.start()
}
