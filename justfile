dev:
    ag -l --go | entr -r godotenv go run .

build:
    CC=musl-gcc go build -ldflags='-linkmode external -extldflags "-static"' -o ./countries

deploy: build
    ssh root@turgot 'systemctl stop countries';
    scp countries turgot:countries/countries
    ssh root@turgot 'systemctl start countries'
