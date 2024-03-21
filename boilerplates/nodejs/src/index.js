// Import the http module
const http = require("http");

// Define the hostname and port
const port = 3000;

// Create the HTTP server
const server = http.createServer((_, res) => {
  // Set the response status code and headers
  res.statusCode = 200;
  res.setHeader("Content-Type", "text/plain");

  // Send the response body
  res.end("Hello, World!\n");
});

// Start listening for incoming requests
server.listen(port, () => {
  console.log(`Server running at http://*:${port}/`);
});
