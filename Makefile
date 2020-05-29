GO_VERSION = 1.13
DETA_VERSION = 1.0
PLATFORM = linux/amd64

LDFLAGS := -X github.com/deta/deta-cli/cmd.detaVersion=$(DETA_VERSION) $(LDFLAGS)
LDFLAGS := -X github.com/deta/deta-cli/cmd.goVersion=$(GO_VERSION) $(LDFLAGS)
LDFLAGS := -X github.com/deta/deta-cli/cmd.platform=$(PLATFORM) $(LDFLAGS)
LDFLAGS := -X github.com/deta/deta-cli/auth.cognitoClientID=$(COGNITO_CLIENT_ID) $(LDFLAGS)
LDFLAGS := -X github.com/deta/deta-cli/auth.accessKeyID=$(ACCESS_KEY_ID) $(LDFLAGS)
LDFLAGS := -X github.com/deta/deta-cli/auth.acessKeySecret=$(ACCESS_KEY_SECRET) $(LDFLAGS)
LDFLAGS := -X github.com/deta/deta-cli/auth.userpoolRegion=$(USERPOOL_REGION) $(LDFLAGS)

.PHONY: build clean

build:
	go build -ldflags="$(LDFLAGS)" -o build/deta	

clean:
	rm -rf build