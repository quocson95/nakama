dev:
	go mod vendor
	go mod tidy
	docker run --rm -w "/app" -v "$(pwd):/app" heroiclabs/nakama-pluginbuilder:3.11.0 build --trimpath --mod=vendor -o ./bin/nakama
copy:
	sudo cp ./bin/nakama ../cgb-lobby-module/bin/