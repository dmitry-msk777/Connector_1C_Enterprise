set GOOS=linux
set GOARCH=amd64
go build -o go-example-sqlquerybuilder-linux64
docker build --no-cache -t go-example-sqlquerybuilder:dev .
docker run -p 8080:8080 --name sqlquerybuilder go-example-sqlquerybuilder:dev