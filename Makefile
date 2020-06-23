DETA_VERSION = v0.1-beta
LINUX_PLATFORM = x86_64-linux
MAC_PLATFORM = x86_64-darwin
WINDOWS_PLATFORM = x86_64-windows

LDFLAGS := -X github.com/deta/deta-cli/cmd.detaVersion=$(DETA_VERSION) $(LDFLAGS)
LDFLAGS := -X github.com/deta/deta-cli/auth.loginURL=$(LOGIN_URL) $(LDFLAGS)
LDFLAGS := -X github.com/deta/deta-cli/auth.cognitoClientID=$(COGNITO_CLIENT_ID) $(LDFLAGS)
LDFLAGS := -X github.com/deta/deta-cli/auth.cognitoRegion=$(COGNITO_REGION) $(LDFLAGS)
LDFLAGS := -X github.com/deta/deta-cli/api.version=$(DETA_VERSION) $(LDFLAGS)
LDFLAGS := -X github.com/deta/deta-cli/cmd.gatewayDomain=$(GATEWAY_DOMAIN) $(LDFLAGS)

.PHONY: build clean

build:
	go build -ldflags="$(LDFLAGS) -X github.com/deta/deta-cli/cmd.platform=$(LINUX_PLATFORM)" -o build/deta	
	cd build && rm deta-$(LINUX_PLATFORM).zip && zip deta-$(LINUX_PLATFORM).zip deta

clean:
	rm -rf build