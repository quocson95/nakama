dev:
	go mod vendor
	go mod tidy
	docker run --rm -w "/app" -v "$(pwd):/app" heroiclabs/nakama-pluginbuilder:3.11.0 build --trimpath --mod=vendor -o ./bin/nakama
copy:
	sudo cp ./bin/nakama ../cgb-lobby-module/bin/

cp:
	rm ../../cgp/dev/nakama || true && rm ../../cgp/dev-1/nakama || true
	cp ./bin/nakama ../../cgp/dev/nakama &&	cp ./bin/nakama ../../cgp/dev-1/nakama
	docker restart nakama_dev && docker restart nakama_dev_1

cgp-dev:
	rsync -aurv --delete ./bin/nakama root@cgpdev:/root/cgp-server/dev