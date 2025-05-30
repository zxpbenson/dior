PREFIX=/usr/local
BINDIR=${PREFIX}/bin
DESTDIR=
BLDDIR = build

# GOOS 的可选项包括：
# darwin：macOS
# linux：Linux
# windows：Windows
# freebsd：FreeBSD
# netbsd：NetBSD
# openbsd：OpenBSD
# dragonfly：DragonFly BSD
# android：Android
# ios：iOS
# js：JavaScript（用于 WebAssembly）

# GOARCH 的可选项包括：
# amd64：64 位 x86 架构
# 386：32 位 x86 架构
# arm：32 位 ARM 架构
# arm64：64 位 ARM 架构（也称为 AArch64）
# mips：32 位 MIPS 架构
# mipsle：32 位小端 MIPS 架构
# mips64：64 位 MIPS 架构
# mips64le：64 位小端 MIPS 架构
# ppc：32 位 PowerPC 架构
# ppc64：64 位 PowerPC 架构
# ppc64le：64 位小端 PowerPC 架构
# s390x：64 位 IBM Z 架构

BLDFLAGS = CGO_ENABLED=0 GOOS=linux GOARCH=amd64 #golang的交叉编译妙啊
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
