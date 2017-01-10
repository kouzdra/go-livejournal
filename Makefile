all: ljmain
	./$<

ljmain::
	go build ljmain.go
