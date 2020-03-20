VERSION        ?= 0.4.1
ORACLE_VERSION ?= 19.5
LDFLAGS        := -X main.Version=$(VERSION)
GOFLAGS        := -ldflags "$(LDFLAGS) -s -w"
ARCH           ?= $(shell uname -m)
GOARCH         ?= $(subst x86_64,amd64,$(patsubst i%86,386,$(ARCH)))
RPM_VERSION    ?= $(ORACLE_VERSION).0.0.0-1
ORA_RPM         = oracle-instantclient$(ORACLE_VERSION)-devel-$(RPM_VERSION).$(ARCH).rpm oracle-instantclient$(ORACLE_VERSION)-basic-$(RPM_VERSION).$(ARCH).rpm
ORA_ZIP         = instantclient-basic-linux.x64-19.5.0.0.0dbru.zip
LD_LIBRARY_PATH = /usr/lib/oracle/$(ORACLE_VERSION)/client64/lib
BUILD_ARGS      = --build-arg VERSION=$(VERSION) --build-arg ORACLE_VERSION=$(ORACLE_VERSION)
DIST_DIR        = prometheus_oracle.$(VERSION)-ora$(ORACLE_VERSION).linux-${GOARCH}
ARCHIVE         = prometheus_oracle.$(VERSION)-ora$(ORACLE_VERSION).linux-${GOARCH}.tar.gz

#export LD_LIBRARY_PATH ORACLE_VERSION

%.rpm:
	wget -q http://yum.oracle.com/repo/OracleLinux/OL7/oracle/instantclient/$(ARCH)/getPackage/$@

%.zip:
	wget -q https://download.oracle.com/otn_software/linux/instantclient/195000/$@

download-rpms: $(ORA_RPM)

download-zips: $(ORA_ZIP)

prereq: download-rpms download-zips
	@echo deps
	sudo apt-get update
	sudo apt-get install --no-install-recommends -qq libaio1 rpm
	sudo rpm -Uvh --nodeps --force oracle*rpm
	echo $(LD_LIBRARY_PATH) | sudo tee /etc/ld.so.conf.d/oracle.conf
	sudo ldconfig

oci.pc:
	sed "s/@ORACLE_VERSION@/$(ORACLE_VERSION)/g" oci8.pc.template > oci8.pc

linux: oci.pc
	@echo build linux
	mkdir -p ./dist/$(DIST_DIR)
	PKG_CONFIG_PATH=${PWD} GOOS=linux go build $(GOFLAGS) -o ./dist/$(DIST_DIR)/prometheus_oracle
	(cd dist ; tar cfz $(ARCHIVE) $(DIST_DIR))

darwin: oci.pc
	@echo build darwin
	mkdir -p ./dist/prometheus_oracle.$(VERSION).darwin-${GOARCH}
	PKG_CONFIG_PATH=${PWD} GOOS=darwin go build $(GOFLAGS) -o ./dist/prometheus_oracle.$(VERSION).darwin-${GOARCH}/prometheus_oracle
	(cd dist ; tar cfz prometheus_oracle.$(VERSION).darwin-${GOARCH}.tar.gz prometheus_oracle.$(VERSION).darwin-${GOARCH})

local-build:  linux

build: docker

deps:
	@PKG_CONFIG_PATH=${PWD} go get

test:
	@echo test
	@PKG_CONFIG_PATH=${PWD} go test $$(go list ./... | grep -v /vendor/)

clean:
	rm -rf ./dist sgerrand.rsa.pub glibc-2.29-r0.apk oci8.pc

docker: ubuntu-image

ubuntu-image: $(ORA_RPM) $(ORA_ZIP)
	docker build $(BUILD_ARGS)  -t "juev/prometheus_oracle:$(VERSION)" .
	docker tag "juev/prometheus_oracle:$(VERSION)" "juev/prometheus_oracle:latest"

travis: oci.pc prereq deps test linux docker
	@true

.PHONY: build deps test clean docker travis oci.pc
