package cli

import "fmt"

type Output struct {
	ch     chan string
	prompt string
}

func (o *Output) Run() {
	for msg := range o.ch {
		fmt.Print("\r\033[K")
		fmt.Println(msg)
		fmt.Print(o.prompt)
	}
}

func (o *Output) Send(msg string) {
	o.ch <- msg
}
