all:	logger ring filemap

logger:
	cd logger;	go build

ring:
	cd ring;	go build

filemap:
	cd filemap;	go build
