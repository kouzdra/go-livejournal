all: ljmain
	./$<

ljmain::
	go build ljmain.go

clean:
	go clean ljmain.go