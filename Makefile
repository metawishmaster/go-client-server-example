.PHONY: all clean server client

BINDIR := bin
SOURCES := server.go client.go
TARGETS :=  $(BINDIR)/server $(BINDIR)/client

all: $(TARGETS)

$(BINDIR)/%: %.go | $(BINDIR)
	go build -o $@ $<

$(BINDIR):
	mkdir -p $(BINDIR)

clean:
	rm -rf $(BINDIR)

server: $(BINDIR)/server
client: $(BINDIR)/client
