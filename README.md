# How to Set Up Golang Web Project (Windows)

1. Downloaded Go from https://go.dev/doc/install
2. Set up two environment variables. This defines the architecture to build for (webassembly) and the operating system (javascript environment)  
    * ```$env:GOARCH = "wasm"```
    * ```$env:GOOS = "js"```
3. Next, I built ```main.go``` into webassembly:
    * ```go build -o main.wasm```
4. Finally, to allow my html to have any idea what's going on, I added a web assembly executable file to the directory:
    * ```cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .```
    * In html, this is called via before anything else:
```js
<script src="js/wasm_exec.js"></script>
```
5. The following three lines load all of the Go functions, now able to be called in js:
```js
const go = new Go();
            const wasm = await WebAssembly.instantiateStreaming(fetch('main.wasm'), go.importObject);
            go.run(wasm.instance);
```