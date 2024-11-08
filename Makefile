all:
	go build -ldflags "-X main.version=1.0.9" -o ds_un 

clean:
	rm -rf ds_un
