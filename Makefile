GO_VERSION = 1.13
DETA_VERSION = DEV
PLATFORM = linux/amd64

LDFLAGS := -X github.com/deta/deta-cli/cmd.detaVersion=$(DETA_VERSION) $(LDFLAGS)
LDFLAGS := -X github.com/deta/deta-cli/cmd.goVersion=$(GO_VERSION) $(LDFLAGS)
LDFLAGS := -X github.com/deta/deta-cli/cmd.platform=$(PLATFORM) $(LDFLAGS)
LDFLAGS := -X github.com/deta/deta-cli/auth.loginURL=$(LOGIN_URL) $(LDFLAGS)
LDFLAGS := -X github.com/deta/deta-cli/api.version=$(DETA_VERSION) $(LDFLAGS)

.PHONY: build clean

build:
	go build -ldflags="$(LDFLAGS)" -o build/deta	

clean:
	rm -rf build