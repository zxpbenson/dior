PREFIX=/usr/local
BINDIR=${PREFIX}/bin
DESTDIR=
BLDDIR = build
#BLDFLAGS = CGO_ENABLED=0 GOOS=linux GOARCH=amd64 #golang的交叉编译秒啊
EXT=
ifeq (${GOOS},windows)
    EXT=.exe
endif

APPS = dior some kafka-consumer
all: $(APPS)

$(BLDDIR)/dior:           $(wildcard apps/dior/*.go cache/*.go lg/*.go option/*.go pressor/*.go writer/*.go)
$(BLDDIR)/some:           $(wildcard apps/some/*.go cache/*.go lg/*.go option/*.go pressor/*.go writer/*.go)
$(BLDDIR)/kafka-consumer: $(wildcard apps/kafka-consumer/*.go cache/*.go lg/*.go option/*.go pressor/*.go writer/*.go)

$(BLDDIR)/%:
	@mkdir -p $(dir $@)
	${BLDFLAGS} go build -o $@ ./apps/$*

$(APPS): %: $(BLDDIR)/%

clean:
	rm -fr $(BLDDIR)

.PHONY: install clean all
.PHONY: $(APPS)

install: $(APPS)
	install -m 755 -d ${DESTDIR}${BINDIR}
	for APP in $^ ; do install -m 755 ${BLDDIR}/$$APP ${DESTDIR}${BINDIR}/$$APP${EXT} ; done
