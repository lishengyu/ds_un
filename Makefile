all:
	go build -ldflags "-X main.version=1.0.8" -o ds_un 

clean:
	rm -rf ds_un
