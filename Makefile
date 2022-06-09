DETA_VERSION = DEV
LINUX_PLATFORM = x86_64-linux
LINUX_ARM_PLATFORM= arm64-linux
MAC_PLATFORM = x86_64-darwin
MAC_ARM_PLATFORM = arm64-darwin
WINDOWS_PLATFORM_I386 = x86-windows
WINDOWS_PLATFORM_AMD64 = x86_64-windows

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

build-linux-arm:
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS) -X github.com/deta/deta-cli/cmd.platform=$(LINUX_ARM_PLATFORM)" -o build/deta	
	cd build && zip -FSr deta-$(LINUX_ARM_PLATFORM).zip deta

build-win-i386:
	GOOS=windows GOARCH=386 go build -ldflags="$(LDFLAGS) -X github.com/deta/deta-cli/cmd.platform=$(WINDOWS_PLATFORM_I386)" -o build/deta.exe	
	cd build && zip -FSr deta-$(WINDOWS_PLATFORM_I386).zip deta.exe

build-win-amd64:
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS) -X github.com/deta/deta-cli/cmd.platform=$(WINDOWS_PLATFORM_AMD64)" -o build/deta.exe	
	cd build && zip -FSr deta-$(WINDOWS_PLATFORM_AMD64).zip deta.exe

build-mac:
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS) -X github.com/deta/deta-cli/cmd.platform=$(MAC_PLATFORM)" -o build/deta	
	cd build && zip -FSr deta-$(MAC_PLATFORM).zip deta

build-mac-arm:
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS) -X github.com/deta/deta-cli/cmd.platform=$(MAC_ARM_PLATFORM)" -o build/deta	
	cd build && zip -FSr deta-$(MAC_ARM_PLATFORM).zip deta

build: build-linux build-win-i386 build-win-amd64 build-mac build-mac-arm build-linux-arm

clean:
	rm -rf build
