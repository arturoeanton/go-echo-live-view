cd cmd/wasm/
GOOS=js GOARCH=wasm go build -o  ../../assets/json.wasm 
echo "build wasm success"
cd -
go run example/example2/example2.go