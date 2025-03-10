dev:
    ag -l --go | entr -r godotenv go run .

build:
    CC=musl-gcc go build -ldflags='-linkmode external -extldflags "-static"' -o ./countries

deploy: build
    ssh root@cantillon 'systemctl stop countries';
    scp countries cantillon:countries/countries
    ssh root@cantillon 'systemctl start countries'
