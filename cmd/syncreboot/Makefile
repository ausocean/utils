BIN = syncreboot
DIR = /usr/local/bin

all:
	go build
	@echo "\nNow run sudo make install\n"

install:
	cp -pf $(BIN) $(DIR)
	sudo setcap cap_sys_boot=ep $(DIR)/$(BIN)

clean:
	rm $(BIN)
