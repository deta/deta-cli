DETA_VERSION = v1.1.2-beta
LINUX_PLATFORM = linux_x86_64
RASPI_PLATFORM = linux_arm
MAC_PLATFORM = darwin_x86_64
WINDOWS_PLATFORM = windows_x86_64

LDFLAGS := -X github.com/deta/deta-cli/cmd.detaVersion=$(DETA_VERSION) $(LDFLAGS)
LDFLAGS := -X github.com/deta/deta-cli/cmd.gatewayDomain=$(GATEWAY_DOMAIN) $(LDFLAGS)
LDFLAGS := -X github.com/deta/deta-cli/cmd.visorURL=$(VISOR_URL) $(LDFLAGS)
LDFLAGS := -X github.com/deta/deta-cli/auth.loginURL=$(LOGIN_URL) $(LDFLAGS)
LDFLAGS := -X github.com/deta/deta-cli/auth.cognitoClientID=$(COGNITO_CLIENT_ID) $(LDFLAGS)
LDFLAGS := -X github.com/deta/deta-cli/auth.cognitoRegion=$(COGNITO_REGION) $(LDFLAGS)
LDFLAGS := -X github.com/deta/deta-cli/auth.detaSignVersion=$(DETA_SIGN_VERSION) $(LDFLAGS)
LDFLAGS := -X github.com/deta/deta-cli/api.version=$(DETA_VERSION) $(LDFLAGS)

.PHONY: build clean

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS) -X github.com/deta/deta-cli/cmd.platform=$(LINUX_PLATFORM)" -o build/deta	
	cd build && zip -FSr deta-$(LINUX_PLATFORM).zip deta

build-raspi:
	GOOS=linux GOARCH=arm go build -ldflags="$(LDFLAGS) -X github.com/deta/deta-cli/cmd.platform=$(RASPI_PLATFORM)" -o build/deta	
	cd build && zip -FSr deta-$(RASPI_PLATFORM).zip deta

build-win:
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS) -X github.com/deta/deta-cli/cmd.platform=$(WINDOWS_PLATFORM)" -o build/deta.exe	
	cd build && zip -FSr deta-$(WINDOWS_PLATFORM).zip deta.exe

build-mac:
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS) -X github.com/deta/deta-cli/cmd.platform=$(MAC_PLATFORM)" -o build/deta	
	cd build && zip -FSr deta-$(MAC_PLATFORM).zip deta

build: build-linux build-win build-mac build-raspi

clean:
	rm -rf build
