const http = require('http');
const fs = require('fs');
const path = require('path');

// Create an HTTP server that handles incoming requests
const server = http.createServer((req, res) => {
  // Determine the requested file path, defaulting to 'index.html' for '/'
  const filePath = req.url === '/' ? '/index.html' : req.url;

  // Set the root directory where your public files are located
  const rootDir = path.join(__dirname, 'public');

  // Create a readable stream for the requested file
  const file = fs.createReadStream(path.join(rootDir, filePath));

  // Set the correct MIME type for WebAssembly files
  if (path.extname(filePath) === '.wasm') {
    res.setHeader('Content-Type', 'application/wasm');
  }

  // Send a 200 OK response to the client
  res.writeHead(200);

  // Pipe the file's content to the response, serving the file to the client
  file.pipe(res);
});

// Define the port for the server, defaulting to 3000
const port = process.env.PORT || 3000;

// Start the server and listen on the specified port
server.listen(port, () => {
  console.log(`Server is running on port ${port}`);
});
