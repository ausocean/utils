BIN = syncreboot
BASEDIR = /opt/ausocean
BINDIR = $(BASEDIR)/bin

all:
	go build
	@echo "\nNow run sudo make install\n"

install:
	mkdir -p $(BINDIR)
	chmod -R guo+rwx $(BASEDIR)
	cp -pf $(BIN) $(BINDIR)
	sudo setcap cap_sys_boot=ep $(BINDIR)/$(BIN)

clean:
	rm $(BIN)
