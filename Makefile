default: test

fmt: 
	go fmt ./...

coverage: fmt
	go test ./ -coverprofile=coverage.out
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out
	rm coverage.out

test: fmt 
	go vet ./...
	go test ./...

pprof:
	go test -c
	./gorbac.test -test.cpuprofile cpu.prof -test.bench .
	go tool pprof gorbac.test cpu.prof
	rm cpu.prof gorbac.test

flamegraph:
	go test -c
	./gorbac.test -test.cpuprofile cpu.prof -test.bench .
	go-torch ./gorbac.test cpu.prof
	xdg-open torch.svg
	sleep 5
	rm cpu.prof gorbac.test torch.svg

pack:
	mkdir -p _dist
	go build -ldflags "-X main.version=`date +%Y-%m-%d_%H-%M_``git log -1 --format=%h`" ./cmd/gleam/
	mv ./gleam ./_dist/
	cp ./utils/* ./_dist/
	cp -r ./scripts ./_dist/

pack-x86:
	mkdir -p _dist
	env GOOS=linux GOARCH=386 go build -ldflags "-X main.version=`date +%Y-%m-%d_%H-%M_``git log -1 --format=%h`" ./cmd/gleam/
	mv ./gleam ./_dist/
	cp ./utils/* ./_dist/
	cp -r ./scripts ./_dist/

pack-docker: pack
	CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' ./cmd/gleam/
	mv ./gleam ./_dist/
	sudo docker build -t mikespook/gleam _dist/

docker:
	sudo docker run mikespook/gleam
