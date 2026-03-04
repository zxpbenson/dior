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

# 自动检测当前系统架构，但允许通过环境变量覆盖
# 如果未设置 GOOS，则自动检测
ifeq ($(GOOS),)
  GOOS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
  # uname 返回 Darwin (macOS)，Go 使用 darwin
  # Linux 返回 linux，Go 使用 linux
  # Windows 需要特殊处理
endif
# 如果未设置 GOARCH，则自动检测
ifeq ($(GOARCH),)
  GOARCH := $(shell uname -m)
  # 将 arm64 映射为 arm64，x86_64 映射为 amd64
  ifeq ($(GOARCH),x86_64)
    GOARCH := amd64
  endif
  ifeq ($(GOARCH),aarch64)
    GOARCH := arm64
  endif
endif

BLDFLAGS = CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH)
EXT=
ifeq ($(GOOS),windows)
    EXT=.exe
endif

# 用于调试：显示当前编译配置
.PHONY: show-config
show-config:
	@echo "GOOS: $(GOOS)"
	@echo "GOARCH: $(GOARCH)"
	@echo "BLDFLAGS: $(BLDFLAGS)"
	@echo "EXT: $(EXT)"

APPS = dior some kafka-consumer
all: $(APPS)

$(BLDDIR)/dior:           $(wildcard cmd/dior/*.go internal/cache/*.go internal/lg/*.go option/*.go internal/source/*.go internal/sink/*.go component/*.go internal/version/*.go)
$(BLDDIR)/some:           $(wildcard cmd/some/*.go internal/cache/*.go internal/lg/*.go option/*.go internal/source/*.go internal/sink/*.go component/*.go internal/version/*.go)
$(BLDDIR)/kafka-consumer: $(wildcard cmd/kafka-consumer/*.go internal/cache/*.go internal/lg/*.go option/*.go internal/source/*.go internal/sink/*.go component/*.go internal/version/*.go)

$(BLDDIR)/%:
	@mkdir -p $(dir $@)
	$(BLDFLAGS) go build -o $@ ./cmd/$*

$(APPS): %: $(BLDDIR)/%

clean:
	rm -fr $(BLDDIR)

.PHONY: install clean all test show-config
.PHONY: $(APPS)

test:
	go test -v -race -cover ./...

install: $(APPS)
	install -m 755 -d ${DESTDIR}${BINDIR}
	for APP in $^ ; do install -m 755 ${BLDDIR}/$$APP ${DESTDIR}${BINDIR}/$$APP${EXT} ; done
